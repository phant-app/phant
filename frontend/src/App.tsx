import { type ReactNode, useEffect, useState } from 'react';
import { NavLink, Navigate, Route, Routes } from 'react-router-dom';
import { BaseLayout } from './components/layout/BaseLayout';
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
    trace?: DumpTraceFrame[];
};

type DumpTraceFrame = {
    file?: string;
    line?: number;
    func?: string;
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
            { to: '/php-versions', label: 'PHP' },
            { to: '/sites', label: 'Sites' },
            { to: '/valet-setups', label: 'Valet Setups' },
        ],
    },
    {
        title: 'Tools',
        items: [
            { to: '/dumps', label: 'Dumps' },
            { to: '/setup', label: 'PHP Setup' },
            { to: '/valet', label: 'Valet Linux' },
        ],
    },
];

import { ThemeProvider } from './components/theme-provider';
import { PhpManagerPage } from './pages/PhpManagerPage';
import { ValetSitesPage } from './pages/ValetSitesPage';
import { ServicesPage } from './pages/ServicesPage';
import { SettingsPage } from './pages/SettingsPage';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

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
        <div className="flex items-start justify-between gap-4 rounded-md border border-border bg-muted/40 px-3 py-2 text-sm">
            <span className="text-muted-foreground">{label}</span>
            <span className="text-right font-medium text-foreground">{value}</span>
        </div>
    );
}

function JsonBox({ value }: { value: unknown }) {
    return (
        <pre className="overflow-x-auto rounded-md border border-border bg-muted/50 p-3 text-xs text-muted-foreground font-mono">
            {JSON.stringify(value, null, 2)}
        </pre>
    );
}

function PageCard({ title, actions, children }: { title: string; actions?: ReactNode; children: ReactNode }) {
    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-base font-semibold">{title}</CardTitle>
                {actions && <div className="flex items-center gap-2">{actions}</div>}
            </CardHeader>
            <CardContent className="space-y-4 pt-4">
                {children}
            </CardContent>
        </Card>
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
        <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={onClick}
            disabled={disabled}
        >
            {children}
        </Button>
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
    const getCallsiteLabel = (event: DumpEvent): string | null => {
        const frame = event.trace?.[0];
        if (!frame?.file || !frame?.line) {
            return null;
        }

        if (frame.func) {
            return `${frame.file}:${frame.line} (${frame.func})`;
        }

        return `${frame.file}:${frame.line}`;
    };

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Live Dumps</h1>
                <p className="text-muted-foreground mt-2">
                    Monitor dump() and dd() calls from your Laravel application in real-time.
                </p>
            </div>
            
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

            <PageCard title={`Dumps (${events.length})`}>
                {events.length === 0 ? (
                    <p className="text-sm text-slate-400">No dumps yet. Trigger `dump()` or `dd()` in Laravel.</p>
                ) : (
                    <div className="space-y-3">
                        {events.map((event) => {
                            const callsiteLabel = getCallsiteLabel(event);

                            return (
                                <div key={event.id} className="space-y-2">
                                    {callsiteLabel ? (
                                        <div className="rounded-md border border-slate-800 bg-slate-950 px-3 py-2 text-xs text-slate-300">
                                            <span className="text-slate-500">Callsite:</span> {callsiteLabel}
                                        </div>
                                    ) : null}
                                    <JsonBox value={event} />
                                </div>
                            );
                        })}
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
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">CLI Diagnostics</h1>
                <p className="text-muted-foreground mt-2">
                    Verify that your PHP CLI setup includes the correct prepended hooks to capture dumps.
                </p>
            </div>

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
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Valet Linux</h1>
                <p className="text-muted-foreground mt-2">
                    Diagnose and manage your Valet configuration safely. Review FPM hook targets and apply remediation.
                </p>
            </div>

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
        <ThemeProvider defaultTheme="dark" storageKey="phant-ui-theme">
            <BaseLayout>
                <Routes>
                    <Route path="/" element={<Navigate to="/dumps" replace />} />
                    <Route
                        path="/php"
                    element={
                        <div className="space-y-8">
                            <PhpManagerPage />
                            <SetupPage
                                diagnostics={diagnostics}
                                hookResult={hookResult}
                                installingHook={installingHook}
                                onRefresh={refreshDiagnostics}
                                onEnable={enableCLIHook}
                            />
                        </div>
                    }
                />
                <Route path="/sites" element={<ValetSitesPage />} />
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
                <Route path="/services" element={<ServicesPage />} />
                <Route
                    path="/dumps"
                    element={<DumpsPage channelName={channelName} status={status} events={events} onClear={clearEvents} />}
                />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="*" element={<Navigate to="/dumps" replace />} />
            </Routes>
        </BaseLayout>
        </ThemeProvider>
    );
}

export default App
