import { type ReactNode, useEffect, useState } from 'react';
import { NavLink, Navigate, Route, Routes } from 'react-router-dom';
import { EventsOffAll, EventsOn } from '../wailsjs/runtime/runtime';
import {
    ApplyValetLinuxRemediation,
    DumpEventChannelName,
    EnableCLIHook,
    GetCollectorStatus,
    GetRecentEvents,
    GetSetupDiagnostics,
    GetValetLinuxVerification,
} from '../wailsjs/go/main/App';

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

const navGroups = [
    {
        title: 'Product',
        items: [
            { to: '/php-versions', label: 'PHP Versions' },
            { to: '/sites', label: 'Sites' },
            { to: '/valet-setups', label: 'Valet Setups' },
        ],
    },
    {
        title: 'Tools',
        items: [
            { to: '/dumps', label: 'Live Dumps' },
            { to: '/setup', label: 'PHP Setup' },
            { to: '/valet', label: 'Valet Linux' },
        ],
    },
];

function PlaceholderPage({
    title,
    subtitle,
    nextSteps,
}: {
    title: string;
    subtitle: string;
    nextSteps: string[];
}) {
    return (
        <PageCard title={title}>
            <p className="text-sm text-slate-300">{subtitle}</p>
            <div className="rounded-md border border-slate-800 bg-slate-950 p-3">
                <p className="mb-2 text-xs uppercase tracking-wide text-slate-500">Planned MVP scope</p>
                <ul className="list-inside list-disc space-y-1 text-sm text-slate-300">
                    {nextSteps.map((item) => (
                        <li key={item}>{item}</li>
                    ))}
                </ul>
            </div>
        </PageCard>
    );
}

function ValueRow({ label, value }: { label: string; value: string }) {
    return (
        <div className="flex items-start justify-between gap-4 rounded-md border border-slate-800 bg-slate-900/40 px-3 py-2">
            <span className="text-slate-400">{label}</span>
            <span className="text-right text-slate-100">{value}</span>
        </div>
    );
}

function JsonBox({ value }: { value: unknown }) {
    return (
        <pre className="overflow-x-auto rounded-md border border-slate-800 bg-slate-950 p-3 text-xs text-slate-200">
            {JSON.stringify(value, null, 2)}
        </pre>
    );
}

function PageCard({ title, actions, children }: { title: string; actions?: ReactNode; children: ReactNode }) {
    return (
        <section className="space-y-4 rounded-lg border border-slate-800 bg-slate-900/50 p-4">
            <div className="flex items-center justify-between gap-3">
                <h2 className="text-base font-semibold text-slate-100">{title}</h2>
                {actions ? <div className="flex items-center gap-2">{actions}</div> : null}
            </div>
            {children}
        </section>
    );
}

function ActionButton({
    onClick,
    disabled,
    children,
}: {
    onClick: () => void;
    disabled?: boolean;
    children: ReactNode;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            disabled={disabled}
            className="rounded-md border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-slate-100 transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
        >
            {children}
        </button>
    );
}

function DumpsPage({
    channelName,
    status,
    events,
    onClear,
}: {
    channelName: string;
    status: CollectorStatus | null;
    events: DumpEvent[];
    onClear: () => void;
}) {
    return (
        <div className="space-y-4">
            <PageCard
                title="Runtime"
                actions={<ActionButton onClick={onClear}>Clear Events</ActionButton>}
            >
                <div className="grid gap-2 md:grid-cols-2">
                    <ValueRow label="Channel" value={channelName || 'loading...'} />
                    <ValueRow label="Collector" value={status?.running ? 'running' : 'stopped'} />
                    <ValueRow label="Dropped" value={String(status?.dropped ?? 0)} />
                    <ValueRow label="Socket" value={status?.socketPath || 'n/a'} />
                    {status?.lastError ? <ValueRow label="Last error" value={status.lastError} /> : null}
                </div>
            </PageCard>

            <PageCard title={`Live dumps (${events.length})`}>
                {events.length === 0 ? (
                    <p className="text-sm text-slate-400">No dumps yet. Trigger `dump()` or `dd()` in Laravel.</p>
                ) : (
                    <div className="space-y-3">
                        {events.map((event) => (
                            <JsonBox key={event.id} value={event} />
                        ))}
                    </div>
                )}
            </PageCard>
        </div>
    );
}

function SetupPage({
    diagnostics,
    hookResult,
    installingHook,
    onRefresh,
    onEnable,
}: {
    diagnostics: SetupDiagnostics | null;
    hookResult: HookInstallResult | null;
    installingHook: boolean;
    onRefresh: () => void;
    onEnable: () => void;
}) {
    return (
        <div className="space-y-4">
            <PageCard
                title="CLI hook"
                actions={
                    <>
                        <ActionButton onClick={onRefresh}>Refresh</ActionButton>
                        <ActionButton onClick={onEnable} disabled={installingHook}>
                            {installingHook ? 'Enabling...' : 'Enable CLI Hook'}
                        </ActionButton>
                    </>
                }
            >
                <div className="grid gap-2 md:grid-cols-2">
                    <ValueRow label="Generated" value={diagnostics?.generatedAt || 'n/a'} />
                    <ValueRow label="PHP found" value={diagnostics?.phpFound ? 'yes' : 'no'} />
                    <ValueRow label="PHP version" value={diagnostics?.phpVersion || 'n/a'} />
                    <ValueRow label="Service manager" value={diagnostics?.serviceManager || 'unknown'} />
                    {diagnostics?.lastError ? <ValueRow label="Diagnostics error" value={diagnostics.lastError} /> : null}
                </div>
            </PageCard>

            {hookResult ? (
                <PageCard title="Last install attempt">
                    <div className="grid gap-2 md:grid-cols-2">
                        <ValueRow label="Status" value={hookResult.success ? 'enabled' : 'failed'} />
                        <ValueRow label="Message" value={hookResult.message || hookResult.error || 'n/a'} />
                        <ValueRow label="php.ini" value={hookResult.phpIniPath || 'n/a'} />
                        <ValueRow label="prepend file" value={hookResult.prependPath || 'n/a'} />
                        <ValueRow label="backup" value={hookResult.backupPath || 'n/a'} />
                    </div>
                    {hookResult.requiresSudo && hookResult.suggestedCmd ? <JsonBox value={hookResult.suggestedCmd} /> : null}
                </PageCard>
            ) : null}
        </div>
    );
}

function ValetPage({
    valetVerification,
    refreshingValet,
    onRefresh,
    confirmValetRemediation,
    onConfirm,
    applyingValetRemediation,
    onApply,
    valetRemediationResult,
}: {
    valetVerification: ValetLinuxVerification | null;
    refreshingValet: boolean;
    onRefresh: () => void;
    confirmValetRemediation: boolean;
    onConfirm: (checked: boolean) => void;
    applyingValetRemediation: boolean;
    onApply: () => void;
    valetRemediationResult: ValetLinuxRemediationResult | null;
}) {
    return (
        <div className="space-y-4">
            <PageCard
                title="Verification"
                actions={
                    <>
                        <ActionButton onClick={onRefresh} disabled={refreshingValet}>
                            {refreshingValet ? 'Checking...' : 'Refresh'}
                        </ActionButton>
                        <ActionButton
                            onClick={onApply}
                            disabled={applyingValetRemediation || !confirmValetRemediation}
                        >
                            {applyingValetRemediation ? 'Applying...' : 'Apply Remediation'}
                        </ActionButton>
                    </>
                }
            >
                <label className="flex items-center gap-2 text-sm text-slate-300">
                    <input
                        type="checkbox"
                        checked={confirmValetRemediation}
                        onChange={(event) => onConfirm(event.target.checked)}
                    />
                    Confirm I want to modify FPM hook INI files and attempt service restarts.
                </label>

                <div className="grid gap-2 md:grid-cols-2">
                    <ValueRow label="Generated" value={valetVerification?.generatedAt || 'n/a'} />
                    <ValueRow label="Supported OS" value={valetVerification?.supported ? 'yes' : 'no'} />
                    <ValueRow label="Valet detected" value={valetVerification?.valetDetected ? 'yes' : 'no'} />
                    <ValueRow label="Service manager" value={valetVerification?.serviceManager || 'n/a'} />
                    <ValueRow label="CLI conf.d" value={valetVerification?.cliConfDPath || 'n/a'} />
                    <ValueRow label="CLI prepend" value={valetVerification?.cliAutoPrepend || 'n/a'} />
                    <ValueRow label="Expected prepend" value={valetVerification?.expectedPrependPath || 'n/a'} />
                    {valetVerification?.lastError ? (
                        <ValueRow label="Verification error" value={valetVerification.lastError} />
                    ) : null}
                </div>
            </PageCard>

            {valetVerification?.fpmServices?.length ? (
                <PageCard title="PHP-FPM services">
                    <div className="space-y-3">
                        {valetVerification.fpmServices.map((service) => (
                            <JsonBox key={service.serviceName} value={service} />
                        ))}
                    </div>
                </PageCard>
            ) : null}

            {valetVerification?.recommendations?.length ? (
                <PageCard title="Recommendations">
                    <ul className="list-inside list-disc space-y-1 text-sm text-slate-300">
                        {valetVerification.recommendations.map((item, index) => (
                            <li key={`${index}-${item}`}>{item}</li>
                        ))}
                    </ul>
                </PageCard>
            ) : null}

            {valetRemediationResult ? (
                <PageCard title="Last remediation run">
                    <div className="grid gap-2 md:grid-cols-2">
                        <ValueRow label="Status" value={valetRemediationResult.applied ? 'applied' : 'not applied'} />
                        <ValueRow
                            label="Message"
                            value={valetRemediationResult.message || valetRemediationResult.error || 'n/a'}
                        />
                    </div>
                    {valetRemediationResult.targets?.length ? (
                        <div className="space-y-3">
                            {valetRemediationResult.targets.map((target) => (
                                <JsonBox key={target.serviceName} value={target} />
                            ))}
                        </div>
                    ) : null}
                    {valetRemediationResult.suggestedCommands?.length ? (
                        <div className="space-y-2">
                            {valetRemediationResult.suggestedCommands.map((command, index) => (
                                <JsonBox key={`${index}-${command}`} value={command} />
                            ))}
                        </div>
                    ) : null}
                </PageCard>
            ) : null}
        </div>
    );
}

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
        <div className="min-h-screen bg-slate-950 text-slate-100">
            <div className="mx-auto flex min-h-screen max-w-7xl">
                <aside className="w-64 border-r border-slate-800 bg-slate-900/40 p-4">
                    <div className="mb-6">
                        <h1 className="text-xl font-semibold">Phant</h1>
                        <p className="text-sm text-slate-400">Developer-first PHP diagnostics</p>
                    </div>

                    <nav className="space-y-5">
                        {navGroups.map((group) => (
                            <div key={group.title} className="space-y-2">
                                <p className="px-2 text-xs uppercase tracking-wide text-slate-500">{group.title}</p>
                                <div className="space-y-2">
                                    {group.items.map((item) => (
                                        <NavLink
                                            key={item.to}
                                            to={item.to}
                                            className={({ isActive }) =>
                                                [
                                                    'block rounded-md px-3 py-2 text-sm transition',
                                                    isActive
                                                        ? 'bg-slate-200 text-slate-900'
                                                        : 'bg-slate-800/40 text-slate-300 hover:bg-slate-800 hover:text-slate-100',
                                                ].join(' ')
                                            }
                                        >
                                            {item.label}
                                        </NavLink>
                                    ))}
                                </div>
                            </div>
                        ))}
                    </nav>
                </aside>

                <main className="flex-1 space-y-4 overflow-y-auto p-6">
                    <header className="space-y-1">
                        <p className="text-xs uppercase tracking-wide text-slate-500">MVP Workspace</p>
                        <h2 className="text-lg font-semibold">Laravel / PHP runtime tools</h2>
                    </header>

                    <Routes>
                        <Route path="/" element={<Navigate to="/dumps" replace />} />
                        <Route
                            path="/php-versions"
                            element={
                                <PlaceholderPage
                                    title="PHP Versions"
                                    subtitle="Install, switch, and configure global/local PHP runtimes for Laravel projects."
                                    nextSteps={[
                                        'Detect installed versions from system and Valet.',
                                        'Mark one version as active with a safe switch action.',
                                        'Show active CLI/FPM alignment diagnostics.',
                                    ]}
                                />
                            }
                        />
                        <Route
                            path="/sites"
                            element={
                                <PlaceholderPage
                                    title="Sites"
                                    subtitle="List your Valet sites and quickly jump into diagnostics per project."
                                    nextSteps={[
                                        'Read parked and linked sites from Valet config.',
                                        'Show project path, PHP version, and health status.',
                                        'Open site logs and live dumps filtered by project.',
                                    ]}
                                />
                            }
                        />
                        <Route
                            path="/valet-setups"
                            element={
                                <PlaceholderPage
                                    title="Valet Setups"
                                    subtitle="Configure and verify Valet runtime details with a guided developer flow."
                                    nextSteps={[
                                        'Detect Valet driver, paths, and service manager state.',
                                        'Run an end-to-end check for DNS, TLS, and PHP-FPM.',
                                        'Offer one-click remediation with explicit confirmations.',
                                    ]}
                                />
                            }
                        />
                        <Route
                            path="/dumps"
                            element={<DumpsPage channelName={channelName} status={status} events={events} onClear={clearEvents} />}
                        />
                        <Route
                            path="/setup"
                            element={
                                <SetupPage
                                    diagnostics={diagnostics}
                                    hookResult={hookResult}
                                    installingHook={installingHook}
                                    onRefresh={refreshDiagnostics}
                                    onEnable={enableCLIHook}
                                />
                            }
                        />
                        <Route
                            path="/valet"
                            element={
                                <ValetPage
                                    valetVerification={valetVerification}
                                    refreshingValet={refreshingValet}
                                    onRefresh={refreshValetVerification}
                                    confirmValetRemediation={confirmValetRemediation}
                                    onConfirm={setConfirmValetRemediation}
                                    applyingValetRemediation={applyingValetRemediation}
                                    onApply={applyValetRemediation}
                                    valetRemediationResult={valetRemediationResult}
                                />
                            }
                        />
                        <Route path="*" element={<Navigate to="/dumps" replace />} />
                    </Routes>
                </main>
            </div>
        </div>
    );
}

export default App
