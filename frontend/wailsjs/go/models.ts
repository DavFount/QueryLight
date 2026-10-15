export namespace application {
	
	export class Failure {
	    code: string;
	    message: string;
	    sessionLost: boolean;
	    outcomeUnknown: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Failure(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.sessionLost = source["sessionLost"];
	        this.outcomeUnknown = source["outcomeUnknown"];
	    }
	}
	export class ActionReply {
	    id: string;
	    error?: Failure;
	
	    static createFrom(source: any = {}) {
	        return new ActionReply(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.error = this.convertValues(source["error"], Failure);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Cell {
	    kind: string;
	    value: string;
	    unavailable: boolean;
	    byteLength: number;
	
	    static createFrom(source: any = {}) {
	        return new Cell(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.value = source["value"];
	        this.unavailable = source["unavailable"];
	        this.byteLength = source["byteLength"];
	    }
	}
	export class Column {
	    name: string;
	    type: string;
	    nullable: boolean;
	    nullableKnown: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Column(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.nullable = source["nullable"];
	        this.nullableKnown = source["nullableKnown"];
	    }
	}
	export class ConnectionConfig {
	    name: string;
	    host: string;
	    port: number;
	    user: string;
	    password: string;
	    database: string;
	    tls: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.database = source["database"];
	        this.tls = source["tls"];
	    }
	}
	export class ConnectionInfo {
	    id: string;
	    name: string;
	    address: string;
	    version: string;
	    database: string;
	    tls: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.address = source["address"];
	        this.version = source["version"];
	        this.database = source["database"];
	        this.tls = source["tls"];
	    }
	}
	export class ConnectionReply {
	    connection?: ConnectionInfo;
	    error?: Failure;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionReply(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connection = this.convertValues(source["connection"], ConnectionInfo);
	        this.error = this.convertValues(source["error"], Failure);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExecutionRequest {
	    editorID: string;
	    sql: string;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new ExecutionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.editorID = source["editorID"];
	        this.sql = source["sql"];
	        this.mode = source["mode"];
	    }
	}
	
	export class TableInfo {
	    name: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new TableInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}
	export class MetadataReply {
	    schemas: string[];
	    tables: TableInfo[];
	    error?: Failure;
	
	    static createFrom(source: any = {}) {
	        return new MetadataReply(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemas = source["schemas"];
	        this.tables = this.convertValues(source["tables"], TableInfo);
	        this.error = this.convertValues(source["error"], Failure);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TableColumn {
	    name: string;
	    dataType: string;
	    nullable: boolean;
	    generated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TableColumn(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.dataType = source["dataType"];
	        this.nullable = source["nullable"];
	        this.generated = source["generated"];
	    }
	}
	export class TableMetadata {
	    engine: string;
	    primaryKey: string[];
	    columns: TableColumn[];
	    editCandidate: boolean;
	    readOnlyReason: string;
	
	    static createFrom(source: any = {}) {
	        return new TableMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.engine = source["engine"];
	        this.primaryKey = source["primaryKey"];
	        this.columns = this.convertValues(source["columns"], TableColumn);
	        this.editCandidate = source["editCandidate"];
	        this.readOnlyReason = source["readOnlyReason"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QueryResult {
	    table?: TableMetadata;
	    columns: Column[];
	    rows: Cell[][];
	    truncated: boolean;
	    notice: string;
	    affectedRows: string;
	    elapsedMS: number;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.table = this.convertValues(source["table"], TableMetadata);
	        this.columns = this.convertValues(source["columns"], Column);
	        this.rows = this.convertValues(source["rows"], Cell);
	        this.truncated = source["truncated"];
	        this.notice = source["notice"];
	        this.affectedRows = source["affectedRows"];
	        this.elapsedMS = source["elapsedMS"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OperationView {
	    id: string;
	    status: string;
	    result?: QueryResult;
	    error?: Failure;
	
	    static createFrom(source: any = {}) {
	        return new OperationView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status = source["status"];
	        this.result = this.convertValues(source["result"], QueryResult);
	        this.error = this.convertValues(source["error"], Failure);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	

}

