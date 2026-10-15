package application

type ConnectionConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	TLS      bool   `json:"tls"`
}
type Failure struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	SessionLost    bool   `json:"sessionLost"`
	OutcomeUnknown bool   `json:"outcomeUnknown"`
}
type ConnectionInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Version  string `json:"version"`
	Database string `json:"database"`
	TLS      bool   `json:"tls"`
}
type ConnectionReply struct {
	Connection *ConnectionInfo `json:"connection"`
	Error      *Failure        `json:"error"`
}
type ActionReply struct {
	ID    string   `json:"id"`
	Error *Failure `json:"error"`
}
type TableInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type MetadataReply struct {
	Schemas []string    `json:"schemas"`
	Tables  []TableInfo `json:"tables"`
	Error   *Failure    `json:"error"`
}
type Column struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Nullable      bool   `json:"nullable"`
	NullableKnown bool   `json:"nullableKnown"`
}
type Cell struct {
	Kind        string `json:"kind"`
	Value       string `json:"value"`
	Unavailable bool   `json:"unavailable"`
	ByteLength  int    `json:"byteLength"`
}
type QueryResult struct {
	Columns      []Column `json:"columns"`
	Rows         [][]Cell `json:"rows"`
	Truncated    bool     `json:"truncated"`
	Notice       string   `json:"notice"`
	AffectedRows string   `json:"affectedRows"`
	ElapsedMS    int64    `json:"elapsedMS"`
}
type ExecutionRequest struct {
	EditorID string `json:"editorID"`
	SQL      string `json:"sql"`
	Mode     string `json:"mode"`
}
type OperationView struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Result *QueryResult `json:"result"`
	Error  *Failure     `json:"error"`
}
