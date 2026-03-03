import {useEffect, useState} from 'react';
import './App.css';
import {EventsOffAll, EventsOn} from "../wailsjs/runtime/runtime";
import {
    ApplyValetLinuxRemediation,
    DumpEventChannelName,
    EnableCLIHook,
    GetCollectorStatus,
    GetRecentEvents,
    GetSetupDiagnostics,
    GetValetLinuxVerification,
} from "../wailsjs/go/main/App";

type DumpEvent = {
    id: string;
    timestamp: string;
    sourceType: string;
    projectRoot: string;
    isDd: boolean;
    payload: unknown;
};

type CollectorStatus = {
    running: boolean;
    socketPath: string;
    lastError: string;
    dropped: number;
};

type SetupDiagnostics = {
    generatedAt: string;
    phpFound: boolean;
    phpVersion: string;
    phpIniOutput: string;
    serviceManager: string;
    lastError: string;
};

type HookInstallResult = {
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

type FPMServiceStatus = {
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

type ValetLinuxVerification = {
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

type ValetRemediationTarget = {
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

type ValetLinuxRemediationResult = {
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

const MAX_RENDERED_EVENTS = 500;

function App() {
    const [events, setEvents] = useState<DumpEvent[]>([]);
    const [channelName, setChannelName] = useState<string>('');
    const [status, setStatus] = useState<CollectorStatus | null>(null);
    const [diagnostics, setDiagnostics] = useState<SetupDiagnostics | null>(null);
    const [valetVerification, setValetVerification] = useState<ValetLinuxVerification | null>(null);
    const [hookResult, setHookResult] = useState<HookInstallResult | null>(null);
    const [installingHook, setInstallingHook] = useState(false);
    const [refreshingValet, setRefreshingValet] = useState(false);
    const [valetRemediationResult, setValetRemediationResult] = useState<ValetLinuxRemediationResult | null>(null);
    const [applyingValetRemediation, setApplyingValetRemediation] = useState(false);
    const [confirmValetRemediation, setConfirmValetRemediation] = useState(false);

    useEffect(() => {
        let disposed = false;
        let unsubscribe: (() => void) | null = null;

        EventsOffAll();

        const appendEvent = (event: DumpEvent) => {
            setEvents((prev) => {
                if (prev.some((existing) => existing.id === event.id)) {
                    return prev;
                }

                const next = [...prev, event];
                if (next.length <= MAX_RENDERED_EVENTS) {
                    return next;
                }
                return next.slice(next.length - MAX_RENDERED_EVENTS);
            });
        };

        const load = async () => {
            const [resolvedChannel, collectorStatus, recentEvents, setupDiagnostics, valetStatus] = await Promise.all([
                DumpEventChannelName(),
                GetCollectorStatus(),
                GetRecentEvents(MAX_RENDERED_EVENTS),
                GetSetupDiagnostics(),
                GetValetLinuxVerification(),
            ]);

            if (disposed) {
                return;
            }

            setChannelName(resolvedChannel);
            setStatus(collectorStatus);
            setEvents(recentEvents);
            setDiagnostics(setupDiagnostics);
            setValetVerification(valetStatus);

            unsubscribe = EventsOn(resolvedChannel, (event: DumpEvent) => {
                appendEvent(event);
            });
        };

        void load();

        const interval = setInterval(() => {
            if (disposed) {
                return;
            }

            void GetCollectorStatus().then((nextStatus) => {
                if (!disposed) {
                    setStatus(nextStatus);
                }
            });
        }, 2000);

        return () => {
            disposed = true;
            clearInterval(interval);

            if (unsubscribe !== null) {
                unsubscribe();
                unsubscribe = null;
            }

            EventsOffAll();
        };
    }, []);

    const clearEvents = () => setEvents([]);
    const refreshDiagnostics = () => {
        void GetSetupDiagnostics().then(setDiagnostics);
    };

    const refreshValetVerification = async () => {
        setRefreshingValet(true);
        try {
            const result = await GetValetLinuxVerification();
            setValetVerification(result);
        } finally {
            setRefreshingValet(false);
        }
    };

    const enableCLIHook = async () => {
        setInstallingHook(true);
        try {
            const result = await EnableCLIHook();
            setHookResult(result);
            const [latestDiagnostics, latestValetStatus] = await Promise.all([
                GetSetupDiagnostics(),
                GetValetLinuxVerification(),
            ]);
            setDiagnostics(latestDiagnostics);
            setValetVerification(latestValetStatus);
        } finally {
            setInstallingHook(false);
        }
    };

    const applyValetRemediation = async () => {
        setApplyingValetRemediation(true);
        try {
            const result = await ApplyValetLinuxRemediation(confirmValetRemediation);
            setValetRemediationResult(result);
            const latestValetStatus = await GetValetLinuxVerification();
            setValetVerification(latestValetStatus);
        } finally {
            setApplyingValetRemediation(false);
        }
    };

    return (
        <div id="app" className="app-shell">
            <header className="app-header">
                <h1>Phant Live Dumps</h1>
                <button className="btn" onClick={clearEvents}>Clear</button>
            </header>

            <section className="status-grid">
                <div><strong>Runtime channel:</strong> {channelName || 'loading...'}</div>
                <div><strong>Collector running:</strong> {status?.running ? 'yes' : 'no'}</div>
                <div><strong>Dropped events:</strong> {status?.dropped ?? 0}</div>
                <div><strong>Socket:</strong> {status?.socketPath || 'n/a'}</div>
                {status?.lastError ? <div><strong>Last error:</strong> {status.lastError}</div> : null}
            </section>

            <section className="status-grid diagnostics-grid">
                <div className="diagnostics-header">
                    <strong>Setup diagnostics</strong>
                    <div className="actions-row">
                        <button className="btn" onClick={refreshDiagnostics}>Refresh</button>
                        <button className="btn" onClick={enableCLIHook} disabled={installingHook}>
                            {installingHook ? 'Enabling...' : 'Enable CLI Hook'}
                        </button>
                    </div>
                </div>
                <div><strong>Generated:</strong> {diagnostics?.generatedAt || 'n/a'}</div>
                <div><strong>PHP found:</strong> {diagnostics?.phpFound ? 'yes' : 'no'}</div>
                <div><strong>PHP version:</strong> {diagnostics?.phpVersion || 'n/a'}</div>
                <div><strong>Service manager:</strong> {diagnostics?.serviceManager || 'unknown'}</div>
                {diagnostics?.lastError ? <div><strong>Diagnostics error:</strong> {diagnostics.lastError}</div> : null}
                {hookResult ? (
                    <>
                        <div><strong>Hook status:</strong> {hookResult.success ? 'enabled' : 'failed'}</div>
                        <div><strong>Hook message:</strong> {hookResult.message || hookResult.error || 'n/a'}</div>
                        <div><strong>php.ini:</strong> {hookResult.phpIniPath || 'n/a'}</div>
                        <div><strong>prepend file:</strong> {hookResult.prependPath || 'n/a'}</div>
                        <div><strong>backup:</strong> {hookResult.backupPath || 'n/a'}</div>
                        {hookResult.requiresSudo && hookResult.suggestedCmd ? (
                            <>
                                <div><strong>Manual sudo command:</strong></div>
                                <pre className="event-item">{hookResult.suggestedCmd}</pre>
                            </>
                        ) : null}
                    </>
                ) : null}
                <div><strong>Note:</strong> CLI hook may require sudo when PHP config is under /etc. Web/FPM support comes next.</div>
            </section>

            <section className="status-grid diagnostics-grid">
                <div className="diagnostics-header">
                    <strong>Valet Linux verification</strong>
                    <div className="actions-row">
                        <button className="btn" onClick={refreshValetVerification} disabled={refreshingValet}>
                            {refreshingValet ? 'Checking...' : 'Refresh Valet'}
                        </button>
                        <button
                            className="btn"
                            onClick={applyValetRemediation}
                            disabled={applyingValetRemediation || !confirmValetRemediation}
                        >
                            {applyingValetRemediation ? 'Applying...' : 'Apply Remediation'}
                        </button>
                    </div>
                </div>
                <label>
                    <input
                        type="checkbox"
                        checked={confirmValetRemediation}
                        onChange={(event) => setConfirmValetRemediation(event.target.checked)}
                    />{' '}
                    Confirm I want to modify FPM hook ini files and attempt service restarts.
                </label>
                <div><strong>Generated:</strong> {valetVerification?.generatedAt || 'n/a'}</div>
                <div><strong>Supported OS:</strong> {valetVerification?.supported ? 'yes' : 'no'}</div>
                <div><strong>Valet detected:</strong> {valetVerification?.valetDetected ? 'yes' : 'no'}</div>
                <div><strong>Service manager:</strong> {valetVerification?.serviceManager || 'n/a'}</div>
                <div><strong>CLI conf.d path:</strong> {valetVerification?.cliConfDPath || 'n/a'}</div>
                <div><strong>CLI auto_prepend_file:</strong> {valetVerification?.cliAutoPrepend || 'n/a'}</div>
                <div><strong>Expected prepend:</strong> {valetVerification?.expectedPrependPath || 'n/a'}</div>
                {valetVerification?.lastError ? <div><strong>Verification error:</strong> {valetVerification.lastError}</div> : null}

                {valetVerification?.fpmServices?.length ? (
                    <>
                        <div><strong>PHP-FPM services:</strong></div>
                        {valetVerification.fpmServices.map((service) => (
                            <pre key={service.serviceName} className="event-item">{JSON.stringify(service, null, 2)}</pre>
                        ))}
                    </>
                ) : null}

                {valetVerification?.recommendations?.length ? (
                    <>
                        <div><strong>Recommendations:</strong></div>
                        {valetVerification.recommendations.map((item, index) => (
                            <div key={`${index}-${item}`}>- {item}</div>
                        ))}
                    </>
                ) : null}

                {valetRemediationResult ? (
                    <>
                        <div><strong>Remediation status:</strong> {valetRemediationResult.applied ? 'applied' : 'not applied'}</div>
                        <div><strong>Remediation message:</strong> {valetRemediationResult.message || valetRemediationResult.error || 'n/a'}</div>
                        {valetRemediationResult.targets?.length ? (
                            <>
                                <div><strong>Remediation targets:</strong></div>
                                {valetRemediationResult.targets.map((target) => (
                                    <pre key={target.serviceName} className="event-item">{JSON.stringify(target, null, 2)}</pre>
                                ))}
                            </>
                        ) : null}
                        {valetRemediationResult.suggestedCommands?.length ? (
                            <>
                                <div><strong>Suggested commands:</strong></div>
                                {valetRemediationResult.suggestedCommands.map((command, index) => (
                                    <pre key={`${index}-${command}`} className="event-item">{command}</pre>
                                ))}
                            </>
                        ) : null}
                    </>
                ) : null}
            </section>

            <section className="events-section">
                <div className="events-title">Events ({events.length})</div>
                <div className="events-list">
                    {events.length === 0 ? (
                        <div className="empty-state">No dumps yet. Trigger dump() or dd() in Laravel.</div>
                    ) : (
                        events.map((event) => (
                            <pre key={`${event.id}-${event.timestamp}`} className="event-item">
                                {JSON.stringify(event, null, 2)}
                            </pre>
                        ))
                    )}
                </div>
            </section>
        </div>
    );
}

export default App
