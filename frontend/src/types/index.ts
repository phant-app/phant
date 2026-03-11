export type DumpEvent = {
    id: string;
    timestamp: string;
    sourceType: string;
    projectRoot: string;
    isDd: boolean;
    payload: unknown;
    trace?: DumpTraceFrame[];
};

export type DumpTraceFrame = {
    file?: string;
    line?: number;
    func?: string;
};

export type CollectorStatus = {
    running: boolean;
    socketPath: string;
    lastError: string;
    dropped: number;
};

export type SetupDiagnostics = {
    generatedAt: string;
    phpFound: boolean;
    phpVersion: string;
    phpIniOutput: string;
    serviceManager: string;
    lastError: string;
};

export type HookInstallResult = {
    success: boolean;
    alreadyEnabled: boolean;
    phpIniPath: string;
    prependPath: string;
    backupPath: string;
    socketPath: string;
    requiresSudo?: boolean;
    suggestedCmd?: string;
    message: string;
    error: string;
};

export type FPMServiceStatus = {
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
};

export type ValetLinuxVerification = {
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
};

export type ValetRemediationTarget = {
    serviceName: string;
    hookIniPath: string;
    writeAttempted: boolean;
    written: boolean;
    writeError: string;
    restartAttempted: boolean;
    restarted: boolean;
    restartError: string;
    restartCommand: string;
};

export type ValetLinuxRemediationResult = {
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
};

export type ValetSite = {
    name: string;
    path: string;
    url: string;
    isSecure: boolean;
    phpVersion?: string;
};

export type ValetSitesResult = {
    generatedAt: string;
    supported: boolean;
    os: string;
    source: string;
    sites: ValetSite[];
    parkedDirectories: string[];
    warnings: string[];
    error: string;
};