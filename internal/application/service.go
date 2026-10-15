package application

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	values "querylight/internal/database"
)

const MaxRows = 500
const MaxResultBytes = 5 * 1024 * 1024
const MaxColumns = 256
const MaxSQLBytes = 128 * 1024
const QueryTimeout = 30 * time.Second

var errResultLimit = errors.New("result column or metadata limit exceeded")
var errMultipleResults = errors.New("multiple result sets are not supported")

// Do not return editor connections to the metadata pool with session variables,
// schema changes, locks, or manually opened transactions still attached.
func discardConnection(conn *sql.Conn) {
	_ = conn.Raw(func(any) error { return driver.ErrBadConn })
	_ = conn.Close()
}

type connection struct {
	info   ConnectionInfo
	db     *sql.DB
	ctx    context.Context
	cancel context.CancelFunc
}
type editor struct {
	connectionID string
	conn         *sql.Conn
	busy         bool
	lost         bool
}
type operation struct {
	view         OperationView
	cancel       context.CancelFunc
	connectionID string
	editorID     string
}
type Service struct {
	mu          sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	connections map[string]*connection
	editors     map[string]*editor
	operations  map[string]*operation
	connectGate chan struct{}
}

func NewService(ctx context.Context) *Service {
	ctx, cancel := context.WithCancel(ctx)
	return &Service{ctx: ctx, cancel: cancel, connections: map[string]*connection{}, editors: map[string]*editor{}, operations: map[string]*operation{}, connectGate: make(chan struct{}, 2)}
}
func failure(code, message string) *Failure { return &Failure{Code: code, Message: message} }
func safeError(err error) *Failure {
	if errors.Is(err, errResultLimit) {
		return failure("limit", "Result exceeds the 256-column or metadata limit. Select fewer columns.")
	}
	if errors.Is(err, errMultipleResults) {
		return failure("unsupported", "Multiple result sets are not supported yet. The SQL was executed and was not retried.")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return failure("timeout", "The operation timed out. Check the server and try again.")
	}
	if errors.Is(err, context.Canceled) {
		return failure("cancelled", "Operation cancelled.")
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case 1045:
			return failure("authentication", "Authentication failed. Check the username and password.")
		case 1044, 1142:
			return failure("permission", "This account does not have permission for that action.")
		case 1049:
			return failure("schema", "The selected database does not exist or is unavailable.")
		case 1064:
			return failure("sql", "SQL syntax error (1064). Check the statement and server dialect.")
		case 1146:
			return failure("table", "The referenced table does not exist.")
		case 1062:
			return failure("constraint", "A unique key value already exists (1062).")
		default:
			return failure("database", fmt.Sprintf("Database rejected the operation (error %d). Check the statement, schema, and permissions.", me.Number))
		}
	}
	return failure("connection", "Could not communicate with the server. Check the address, port, TLS configuration, and server availability.")
}

type silentLogger struct{}

func (silentLogger) Print(...any) {}
func openDB(ctx context.Context, cfg ConnectionConfig) (*sql.DB, *Failure) {
	if strings.TrimSpace(cfg.Host) == "" || strings.TrimSpace(cfg.User) == "" || cfg.Port < 1 || cfg.Port > 65535 || len(cfg.Host) > 255 || len(cfg.User) > 256 || len(cfg.Database) > 256 || len(cfg.Password) > 4096 {
		return nil, failure("validation", "Enter a valid host, port (1–65535), and username.")
	}
	c := mysql.NewConfig()
	c.Net = "tcp"
	c.Addr = net.JoinHostPort(strings.TrimSpace(cfg.Host), strconv.Itoa(cfg.Port))
	c.User = cfg.User
	c.Passwd = cfg.Password
	c.DBName = cfg.Database
	c.Timeout = 8 * time.Second
	c.ReadTimeout = 35 * time.Second
	c.WriteTimeout = 10 * time.Second
	c.ParseTime = false
	c.MultiStatements = false
	_ = c.Apply(mysql.Charset("utf8mb4", "utf8mb4_unicode_ci"))
	c.Logger = silentLogger{}
	if cfg.TLS {
		c.TLSConfig = "true"
	}
	connector, err := mysql.NewConnector(c)
	if err != nil {
		return nil, failure("validation", "The connection configuration is invalid.")
	}
	db := sql.OpenDB(connector)
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(5 * time.Minute)
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, safeError(err)
	}
	return db, nil
}
func (s *Service) Connect(cfg ConnectionConfig, testOnly bool) ConnectionReply {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()
	select {
	case s.connectGate <- struct{}{}:
		defer func() { <-s.connectGate }()
	case <-ctx.Done():
		return ConnectionReply{Error: safeError(ctx.Err())}
	}
	db, fail := openDB(ctx, cfg)
	if fail != nil {
		return ConnectionReply{Error: fail}
	}
	var version string
	if err := db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		db.Close()
		return ConnectionReply{Error: safeError(err)}
	}
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = cfg.Host
	}
	if len(name) > 100 {
		name = name[:100]
	}
	info := ConnectionInfo{ID: uuid.NewString(), Name: name, Address: net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)), Version: version, Database: cfg.Database, TLS: cfg.TLS}
	if testOnly {
		db.Close()
		info.ID = ""
		return ConnectionReply{Connection: &info}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ctx.Err() != nil || len(s.connections) >= 4 {
		db.Close()
		return ConnectionReply{Error: failure("limit", "Close an existing connection before opening another (maximum 4).")}
	}
	connCtx, connCancel := context.WithCancel(s.ctx)
	s.connections[info.ID] = &connection{info: info, db: db, ctx: connCtx, cancel: connCancel}
	return ConnectionReply{Connection: &info}
}
func (s *Service) connection(id string) (*connection, *Failure) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.connections[id]
	if c == nil {
		return nil, failure("disconnected", "The connection is closed. Connect again.")
	}
	return c, nil
}
func (s *Service) Disconnect(id string) ActionReply {
	s.mu.Lock()
	c := s.connections[id]
	delete(s.connections, id)
	var closeConns []*sql.Conn
	for key, e := range s.editors {
		if e.connectionID == id {
			closeConns = append(closeConns, e.conn)
			delete(s.editors, key)
		}
	}
	for key, o := range s.operations {
		if o.connectionID == id {
			o.cancel()
			delete(s.operations, key)
		}
	}
	s.mu.Unlock()
	if c != nil {
		c.cancel()
		for _, conn := range closeConns {
			discardConnection(conn)
		}
		c.db.Close()
	}
	return ActionReply{}
}
func (s *Service) Shutdown() {
	s.cancel()
	s.mu.Lock()
	ids := make([]string, 0, len(s.connections))
	for id := range s.connections {
		ids = append(ids, id)
	}
	s.mu.Unlock()
	for _, id := range ids {
		s.Disconnect(id)
	}
}

// Native close protection does not need to transfer SQL text to Go.
func (s *Service) HasOpenWork() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Table drafts live only in the frontend. Conservatively guard every open
	// connection so closing all SQL editors cannot bypass draft protection.
	if len(s.connections) > 0 {
		return true
	}
	if len(s.editors) > 0 {
		return true
	}
	for _, o := range s.operations {
		if o.view.Status == "running" {
			return true
		}
	}
	return false
}
func (s *Service) ListSchemas(id string) MetadataReply {
	c, fail := s.connection(id)
	if fail != nil {
		return MetadataReply{Error: fail}
	}
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()
	rows, err := c.db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return MetadataReply{Error: safeError(err)}
	}
	defer rows.Close()
	names := []string{}
	for rows.Next() {
		var n string
		if err = rows.Scan(&n); err != nil {
			return MetadataReply{Error: safeError(err)}
		}
		if len(names) >= 2000 {
			return MetadataReply{Error: failure("limit", "Schema list exceeds the initial 2000-object limit.")}
		}
		names = append(names, n)
	}
	if err = rows.Err(); err != nil {
		return MetadataReply{Error: safeError(err)}
	}
	return MetadataReply{Schemas: names, Tables: []TableInfo{}}
}
func (s *Service) ListTables(id, schema string) MetadataReply {
	if len(schema) > 256 {
		return MetadataReply{Error: failure("validation", "Database name is too long.")}
	}
	c, fail := s.connection(id)
	if fail != nil {
		return MetadataReply{Error: fail}
	}
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()
	rows, err := c.db.QueryContext(ctx, "SELECT TABLE_NAME,TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_SCHEMA=? ORDER BY TABLE_NAME LIMIT 2001", schema)
	if err != nil {
		return MetadataReply{Error: safeError(err)}
	}
	defer rows.Close()
	tables := []TableInfo{}
	for rows.Next() {
		var t TableInfo
		if err = rows.Scan(&t.Name, &t.Type); err != nil {
			return MetadataReply{Error: safeError(err)}
		}
		if len(tables) >= 2000 {
			return MetadataReply{Error: failure("limit", "Table list exceeds the initial 2000-object limit.")}
		}
		tables = append(tables, t)
	}
	if err = rows.Err(); err != nil {
		return MetadataReply{Error: safeError(err)}
	}
	return MetadataReply{Schemas: []string{}, Tables: tables}
}
func (s *Service) OpenEditor(id, schema string) ActionReply {
	if len(schema) > 256 {
		return ActionReply{Error: failure("validation", "Database name is too long.")}
	}
	c, fail := s.connection(id)
	if fail != nil {
		return ActionReply{Error: fail}
	}
	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()
	conn, err := c.db.Conn(ctx)
	if err != nil {
		return ActionReply{Error: safeError(err)}
	}
	if schema != "" {
		if _, err = conn.ExecContext(ctx, "USE "+values.QuoteIdentifier(schema)); err != nil {
			discardConnection(conn)
			return ActionReply{Error: safeError(err)}
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, e := range s.editors {
		if e.connectionID == id {
			count++
		}
	}
	if s.connections[id] != c || len(s.editors) >= 8 || count >= 6 {
		discardConnection(conn)
		return ActionReply{Error: failure("limit", "Close an editor before opening another (maximum 6 per connection, 8 total).")}
	}
	editorID := uuid.NewString()
	s.editors[editorID] = &editor{connectionID: id, conn: conn}
	return ActionReply{ID: editorID}
}
func (s *Service) CloseEditor(id string) ActionReply {
	s.mu.Lock()
	e := s.editors[id]
	delete(s.editors, id)
	for _, o := range s.operations {
		if o.editorID == id && o.view.Status == "running" {
			o.cancel()
		}
	}
	s.mu.Unlock()
	if e != nil {
		discardConnection(e.conn)
	}
	return ActionReply{}
}

// register requires s.mu. IDs exist before asynchronous execution starts.
func (s *Service) register(c *connection, editorID string) (*operation, context.Context, *Failure) {
	if len(s.operations) >= 16 {
		return nil, nil, failure("limit", "Close result tabs before starting more work (maximum 16 retained operations).")
	}
	running := 0
	for _, o := range s.operations {
		if o.connectionID == c.info.ID && o.view.Status == "running" {
			running++
		}
	}
	if running >= 4 {
		return nil, nil, failure("busy", "Wait for an operation to finish (maximum 4 per server).")
	}
	ctx, cancel := context.WithTimeout(c.ctx, QueryTimeout)
	o := &operation{view: OperationView{ID: uuid.NewString(), Status: "running"}, cancel: cancel, connectionID: c.info.ID, editorID: editorID}
	s.operations[o.view.ID] = o
	return o, ctx, nil
}
func (s *Service) StartExecution(req ExecutionRequest) ActionReply {
	if strings.TrimSpace(req.SQL) == "" || len(req.SQL) > MaxSQLBytes {
		return ActionReply{Error: failure("validation", "Enter SQL up to 128 KiB.")}
	}
	if req.Mode != "rows" && req.Mode != "statement" {
		return ActionReply{Error: failure("validation", "Choose Rows or Statement execution mode.")}
	}
	s.mu.Lock()
	e := s.editors[req.EditorID]
	if e == nil || e.lost {
		s.mu.Unlock()
		return ActionReply{Error: &Failure{Code: "session_lost", Message: "This editor session is closed. Open a new SQL tab.", SessionLost: true}}
	}
	if e.busy {
		s.mu.Unlock()
		return ActionReply{Error: failure("busy", "This editor already has a running operation.")}
	}
	c := s.connections[e.connectionID]
	if c == nil {
		s.mu.Unlock()
		return ActionReply{Error: failure("disconnected", "Connect again.")}
	}
	o, ctx, fail := s.register(c, req.EditorID)
	if fail != nil {
		s.mu.Unlock()
		return ActionReply{Error: fail}
	}
	e.busy = true
	s.mu.Unlock()
	go func() {
		start := time.Now()
		var result *QueryResult
		var err error
		if req.Mode == "statement" {
			var r sql.Result
			r, err = e.conn.ExecContext(ctx, req.SQL)
			if err == nil {
				var affected int64
				affected, err = r.RowsAffected()
				if err == nil {
					result = &QueryResult{Columns: []Column{}, Rows: [][]Cell{}, AffectedRows: strconv.FormatInt(affected, 10), Notice: "Statement executed. No automatic retry."}
				}
			}
		} else {
			result, err = query(ctx, e.conn, req.SQL)
		}
		if result != nil {
			result.ElapsedMS = time.Since(start).Milliseconds()
		}
		s.finish(o, e, result, err, ctx)
	}()
	return ActionReply{ID: o.view.ID}
}
func (s *Service) StartInspection(id, schema, table string) ActionReply {
	if schema == "" || table == "" || len(schema) > 256 || len(table) > 256 {
		return ActionReply{Error: failure("validation", "Choose a schema and table.")}
	}
	s.mu.Lock()
	c := s.connections[id]
	if c == nil {
		s.mu.Unlock()
		return ActionReply{Error: failure("disconnected", "Connect again.")}
	}
	o, ctx, fail := s.register(c, "")
	s.mu.Unlock()
	if fail != nil {
		return ActionReply{Error: fail}
	}
	go func() {
		start := time.Now()
		var result *QueryResult
		conn, err := c.db.Conn(ctx)
		if err == nil {
			defer conn.Close()
			var exists int
			err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA=? AND TABLE_NAME=?", schema, table).Scan(&exists)
			if err == nil && exists != 1 {
				err = &mysql.MySQLError{Number: 1146}
			}
			var keys []string
			if err == nil {
				var rows *sql.Rows
				rows, err = conn.QueryContext(ctx, "SELECT COLUMN_NAME FROM information_schema.KEY_COLUMN_USAGE WHERE TABLE_SCHEMA=? AND TABLE_NAME=? AND CONSTRAINT_NAME='PRIMARY' ORDER BY ORDINAL_POSITION", schema, table)
				if err == nil {
					for rows.Next() {
						var key string
						if err = rows.Scan(&key); err != nil {
							break
						}
						keys = append(keys, values.QuoteIdentifier(key))
					}
					if err == nil {
						err = rows.Err()
					}
					rows.Close()
				}
			}
			if err == nil {
				text := "SELECT * FROM " + values.QuoteIdentifier(schema) + "." + values.QuoteIdentifier(table)
				if len(keys) > 0 {
					text += " ORDER BY " + strings.Join(keys, ",")
				}
				text += " LIMIT 501"
				var metadata *TableMetadata
				metadata, err = tableMetadata(ctx, conn, schema, table)
				if err == nil {
					result, err = query(ctx, conn, text)
				}
				if result != nil {
					result.Table = metadata
					result.Notice = "Read-only preview · first 500 rows."
					if len(keys) == 0 {
						result.Notice += " No primary key: order may change."
					}
					result.ElapsedMS = time.Since(start).Milliseconds()
					encoded, _ := json.Marshal(result)
					if len(encoded) > MaxResultBytes {
						result, err = nil, errResultLimit
					}
				}
			}
		}
		s.finish(o, nil, result, err, ctx)
	}()
	return ActionReply{ID: o.view.ID}
}
func (s *Service) finish(o *operation, e *editor, result *QueryResult, err error, ctx context.Context) {
	var fail *Failure
	if err != nil {
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		fail = safeError(err)
	}
	lost := false
	if e != nil && err != nil {
		var me *mysql.MySQLError
		lost = !errors.As(err, &me) && !errors.Is(err, errResultLimit) && !errors.Is(err, errMultipleResults)
		if lost {
			fail.SessionLost = true
			fail.OutcomeUnknown = true
			fail.Message += " The editor session was lost. A write may have executed; inspect the server before retrying."
			discardConnection(e.conn)
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	defer o.cancel()
	if e != nil {
		e.busy = false
		e.lost = lost
	}
	if s.operations[o.view.ID] != o {
		return
	}
	o.view.Result = result
	o.view.Error = fail
	o.view.Status = "succeeded"
	if fail != nil {
		o.view.Status = "failed"
		if fail.Code == "cancelled" {
			o.view.Status = "cancelled"
		}
	}
}
func (s *Service) GetOperation(id string) OperationView {
	s.mu.Lock()
	defer s.mu.Unlock()
	o := s.operations[id]
	if o == nil {
		return OperationView{ID: id, Status: "failed", Error: failure("expired", "This operation was released or its connection closed.")}
	}
	return o.view
}
func (s *Service) CancelOperation(id string) ActionReply {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o := s.operations[id]; o != nil && o.view.Status == "running" {
		o.cancel()
	}
	return ActionReply{}
}
func (s *Service) ReleaseOperation(id string) ActionReply {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o := s.operations[id]; o != nil {
		o.cancel()
		delete(s.operations, id)
	}
	return ActionReply{}
}

func query(ctx context.Context, conn *sql.Conn, text string) (*QueryResult, error) {
	rows, err := conn.QueryContext(ctx, text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	if len(types) > MaxColumns {
		return nil, errResultLimit
	}
	result := &QueryResult{Columns: []Column{}, Rows: [][]Cell{}}
	for _, t := range types {
		nullable, known := t.Nullable()
		result.Columns = append(result.Columns, Column{Name: t.Name(), Type: t.DatabaseTypeName(), Nullable: nullable, NullableKnown: known})
	}
	encoded, _ := json.Marshal(result)
	used := len(encoded) + 1024
	if used > MaxResultBytes {
		return nil, errResultLimit
	}
	if len(types) == 0 {
		result.Notice = "No result columns. For affected-row counts choose Statement mode before execution. SQL has already run; it was not retried."
	}
	for rows.Next() {
		if len(result.Rows) >= MaxRows {
			result.Truncated = true
			break
		}
		raw := make([]any, len(types))
		ptr := make([]any, len(types))
		for i := range raw {
			ptr[i] = &raw[i]
		}
		if err = rows.Scan(ptr...); err != nil {
			return nil, err
		}
		cells := make([]Cell, len(types))
		for i, v := range raw {
			kind, payload, unavailable, length := values.Value(v, types[i].DatabaseTypeName())
			cells[i] = Cell{Kind: kind, Value: payload, Unavailable: unavailable, ByteLength: length}
		}
		b, _ := json.Marshal(cells)
		if used+len(b)+1 > MaxResultBytes {
			result.Truncated = true
			break
		}
		used += len(b) + 1
		result.Rows = append(result.Rows, cells)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if rows.NextResultSet() {
		return nil, errMultipleResults
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return result, nil
}
