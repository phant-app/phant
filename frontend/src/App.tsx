import {useEffect, useState} from 'react';
import './App.css';
import {EventsOffAll, EventsOn} from "../wailsjs/runtime/runtime";
import {
    DumpEventChannelName,
    EnableCLIHook,
    GetCollectorStatus,
    GetRecentEvents,
    GetSetupDiagnostics,
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

const MAX_RENDERED_EVENTS = 500;

function App() {
    const [events, setEvents] = useState<DumpEvent[]>([]);
    const [channelName, setChannelName] = useState<string>('');
    const [status, setStatus] = useState<CollectorStatus | null>(null);
    const [diagnostics, setDiagnostics] = useState<SetupDiagnostics | null>(null);
    const [hookResult, setHookResult] = useState<HookInstallResult | null>(null);
    const [installingHook, setInstallingHook] = useState(false);

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
            const [resolvedChannel, collectorStatus, recentEvents, setupDiagnostics] = await Promise.all([
                DumpEventChannelName(),
                GetCollectorStatus(),
                GetRecentEvents(MAX_RENDERED_EVENTS),
                GetSetupDiagnostics(),
            ]);

            if (disposed) {
                return;
            }

            setChannelName(resolvedChannel);
            setStatus(collectorStatus);
            setEvents(recentEvents);
            setDiagnostics(setupDiagnostics);

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

    const enableCLIHook = async () => {
        setInstallingHook(true);
        try {
            const result = await EnableCLIHook();
            setHookResult(result);
            const latestDiagnostics = await GetSetupDiagnostics();
            setDiagnostics(latestDiagnostics);
        } finally {
            setInstallingHook(false);
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
