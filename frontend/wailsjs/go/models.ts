export namespace dump {
	
	export class CommandMeta {
	    name: string;
	    args?: string[];
	    cwd?: string;
	
	    static createFrom(source: any = {}) {
	        return new CommandMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.args = source["args"];
	        this.cwd = source["cwd"];
	    }
	}
	export class HostMeta {
	    hostname: string;
	    pid: number;
	
	    static createFrom(source: any = {}) {
	        return new HostMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hostname = source["hostname"];
	        this.pid = source["pid"];
	    }
	}
	export class TraceFrame {
	    file?: string;
	    line?: number;
	    func?: string;
	
	    static createFrom(source: any = {}) {
	        return new TraceFrame(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.line = source["line"];
	        this.func = source["func"];
	    }
	}
	export class HTTPMeta {
	    method: string;
	    scheme: string;
	    host: string;
	    path: string;
	    query?: string;
	    statusCode?: number;
	    clientIp?: string;
	    userAgent?: string;
	
	    static createFrom(source: any = {}) {
	        return new HTTPMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.method = source["method"];
	        this.scheme = source["scheme"];
	        this.host = source["host"];
	        this.path = source["path"];
	        this.query = source["query"];
	        this.statusCode = source["statusCode"];
	        this.clientIp = source["clientIp"];
	        this.userAgent = source["userAgent"];
	    }
	}
	export class Event {
	    schemaVersion: number;
	    id: string;
	    timestamp: string;
	    sourceType: string;
	    projectRoot: string;
	    phpSapi: string;
	    requestId?: string;
	    http?: HTTPMeta;
	    command?: CommandMeta;
	    isDd: boolean;
	    payloadFormat: string;
	    payload: number[];
	    trace: TraceFrame[];
	    host: HostMeta;
	
	    static createFrom(source: any = {}) {
	        return new Event(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.schemaVersion = source["schemaVersion"];
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.sourceType = source["sourceType"];
	        this.projectRoot = source["projectRoot"];
	        this.phpSapi = source["phpSapi"];
	        this.requestId = source["requestId"];
	        this.http = this.convertValues(source["http"], HTTPMeta);
	        this.command = this.convertValues(source["command"], CommandMeta);
	        this.isDd = source["isDd"];
	        this.payloadFormat = source["payloadFormat"];
	        this.payload = source["payload"];
	        this.trace = this.convertValues(source["trace"], TraceFrame);
	        this.host = this.convertValues(source["host"], HostMeta);
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

export namespace main {
	
	export class CollectorStatus {
	    running: boolean;
	    socketPath: string;
	    lastError: string;
	    dropped: number;
	
	    static createFrom(source: any = {}) {
	        return new CollectorStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.socketPath = source["socketPath"];
	        this.lastError = source["lastError"];
	        this.dropped = source["dropped"];
	    }
	}

}

export namespace setup {
	
	export class Diagnostics {
	    generatedAt: string;
	    phpFound: boolean;
	    phpVersion: string;
	    phpIniOutput: string;
	    serviceManager: string;
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new Diagnostics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generatedAt = source["generatedAt"];
	        this.phpFound = source["phpFound"];
	        this.phpVersion = source["phpVersion"];
	        this.phpIniOutput = source["phpIniOutput"];
	        this.serviceManager = source["serviceManager"];
	        this.lastError = source["lastError"];
	    }
	}
	export class FPMServiceStatus {
	    serviceName: string;
	    version: string;
	    confDPath: string;
	    hookIniPath: string;
	    hookIniExists: boolean;
	    autoPrependFile: string;
	    matchesExpected: boolean;
	    systemdActive: boolean;
	    systemdEnabled: boolean;
	    restartCommand: string;
	    verificationCommand: string;
	
	    static createFrom(source: any = {}) {
	        return new FPMServiceStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceName = source["serviceName"];
	        this.version = source["version"];
	        this.confDPath = source["confDPath"];
	        this.hookIniPath = source["hookIniPath"];
	        this.hookIniExists = source["hookIniExists"];
	        this.autoPrependFile = source["autoPrependFile"];
	        this.matchesExpected = source["matchesExpected"];
	        this.systemdActive = source["systemdActive"];
	        this.systemdEnabled = source["systemdEnabled"];
	        this.restartCommand = source["restartCommand"];
	        this.verificationCommand = source["verificationCommand"];
	    }
	}
	export class HookInstallResult {
	    success: boolean;
	    alreadyEnabled: boolean;
	    phpIniPath: string;
	    prependPath: string;
	    backupPath: string;
	    socketPath: string;
	    requiresSudo: boolean;
	    suggestedCmd: string;
	    privilegeStrategy: string;
	    privilegeAttempted: boolean;
	    message: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new HookInstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.alreadyEnabled = source["alreadyEnabled"];
	        this.phpIniPath = source["phpIniPath"];
	        this.prependPath = source["prependPath"];
	        this.backupPath = source["backupPath"];
	        this.socketPath = source["socketPath"];
	        this.requiresSudo = source["requiresSudo"];
	        this.suggestedCmd = source["suggestedCmd"];
	        this.privilegeStrategy = source["privilegeStrategy"];
	        this.privilegeAttempted = source["privilegeAttempted"];
	        this.message = source["message"];
	        this.error = source["error"];
	    }
	}
	export class ValetRemediationTarget {
	    serviceName: string;
	    hookIniPath: string;
	    writeAttempted: boolean;
	    written: boolean;
	    writeError: string;
	    restartAttempted: boolean;
	    restarted: boolean;
	    restartError: string;
	    restartCommand: string;
	
	    static createFrom(source: any = {}) {
	        return new ValetRemediationTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serviceName = source["serviceName"];
	        this.hookIniPath = source["hookIniPath"];
	        this.writeAttempted = source["writeAttempted"];
	        this.written = source["written"];
	        this.writeError = source["writeError"];
	        this.restartAttempted = source["restartAttempted"];
	        this.restarted = source["restarted"];
	        this.restartError = source["restartError"];
	        this.restartCommand = source["restartCommand"];
	    }
	}
	export class ValetLinuxRemediationResult {
	    generatedAt: string;
	    supported: boolean;
	    confirmed: boolean;
	    applied: boolean;
	    expectedPrependPath: string;
	    requiresSudo: boolean;
	    suggestedCommands: string[];
	    targets: ValetRemediationTarget[];
	    message: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new ValetLinuxRemediationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generatedAt = source["generatedAt"];
	        this.supported = source["supported"];
	        this.confirmed = source["confirmed"];
	        this.applied = source["applied"];
	        this.expectedPrependPath = source["expectedPrependPath"];
	        this.requiresSudo = source["requiresSudo"];
	        this.suggestedCommands = source["suggestedCommands"];
	        this.targets = this.convertValues(source["targets"], ValetRemediationTarget);
	        this.message = source["message"];
	        this.error = source["error"];
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
	export class ValetLinuxVerification {
	    generatedAt: string;
	    supported: boolean;
	    valetDetected: boolean;
	    serviceManager: string;
	    cliConfDPath: string;
	    cliAutoPrepend: string;
	    expectedPrependPath: string;
	    fpmServices: FPMServiceStatus[];
	    recommendations: string[];
	    lastError: string;
	
	    static createFrom(source: any = {}) {
	        return new ValetLinuxVerification(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.generatedAt = source["generatedAt"];
	        this.supported = source["supported"];
	        this.valetDetected = source["valetDetected"];
	        this.serviceManager = source["serviceManager"];
	        this.cliConfDPath = source["cliConfDPath"];
	        this.cliAutoPrepend = source["cliAutoPrepend"];
	        this.expectedPrependPath = source["expectedPrependPath"];
	        this.fpmServices = this.convertValues(source["fpmServices"], FPMServiceStatus);
	        this.recommendations = source["recommendations"];
	        this.lastError = source["lastError"];
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

