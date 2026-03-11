import { useEffect, useState } from 'react';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { BaseLayout } from './components/layout/BaseLayout';
import { Events } from '@wailsio/runtime';
import {
    DumpEventChannelName,
    GetCollectorStatus,
    GetRecentEvents,
} from '../bindings/phant/internal/services/dumpservice';
import {
    ApplyValetLinuxRemediation,
    EnableCLIHook,
    GetSetupDiagnostics,
    GetValetSites,
    GetValetLinuxVerification,
} from '../bindings/phant/internal/services/setupservice';

import type {
    CollectorStatus,
    DumpEvent,
    HookInstallResult,
    SetupDiagnostics,
    ValetSitesResult,
    ValetLinuxRemediationResult,
    ValetLinuxVerification,
} from './types';

import { ThemeProvider } from './components/theme-provider';
import { PhpManagerPage } from './pages/PhpManagerPage';
import { ValetSitesPage } from './pages/ValetSitesPage';
import { ServicesPage } from './pages/ServicesPage';
import { SettingsPage } from './pages/SettingsPage';
import { DumpsPage } from './pages/DumpsPage';
import { SetupPage } from './pages/SetupPage';
import { ValetPage } from './pages/ValetPage';

const MAX_RENDERED_EVENTS = 500;

const isSameCollectorStatus = (
    previous: CollectorStatus | null,
    next: CollectorStatus,
): boolean => {
    if (previous === null) {
        return false;
    }

    return (
        previous.running === next.running
        && previous.socketPath === next.socketPath
        && previous.lastError === next.lastError
        && previous.dropped === next.dropped
    );
};

function App() {
    const location = useLocation();
    const [events, setEvents] = useState<DumpEvent[]>([]);
    const [channelName, setChannelName] = useState<string>('');
    const [status, setStatus] = useState<CollectorStatus | null>(null);
    const [diagnostics, setDiagnostics] = useState<SetupDiagnostics | null>(null);
    const [valetVerification, setValetVerification] = useState<ValetLinuxVerification | null>(null);
    const [valetSites, setValetSites] = useState<ValetSitesResult | null>(null);
    const [loadingValetSites, setLoadingValetSites] = useState(false);
    const [hookResult, setHookResult] = useState<HookInstallResult | null>(null);
    const [installingHook, setInstallingHook] = useState(false);
    const [refreshingValet, setRefreshingValet] = useState(false);
    const [valetRemediationResult, setValetRemediationResult] = useState<ValetLinuxRemediationResult | null>(null);
    const [applyingValetRemediation, setApplyingValetRemediation] = useState(false);
    const [confirmValetRemediation, setConfirmValetRemediation] = useState(false);

    useEffect(() => {
        let disposed = false;

        const load = async () => {
            setLoadingValetSites(true);

            const [setupDiagnostics, valetStatus, sites] = await Promise.all([
                GetSetupDiagnostics(),
                GetValetLinuxVerification(),
                GetValetSites(),
            ]);

            if (disposed) {
                return;
            }

            setDiagnostics(setupDiagnostics);
            setValetVerification(valetStatus);
            setValetSites(sites);
            setLoadingValetSites(false);
        };

        void load();

        return () => {
            disposed = true;
        };
    }, []);

    useEffect(() => {
        if (location.pathname !== '/dumps') {
            Events.OffAll();
            return;
        }

        let disposed = false;
        let unsubscribe: (() => void) | null = null;

        const appendEvent = (event: DumpEvent) => {
            setEvents((previousEvents) => {
                if (previousEvents.some((existing) => existing.id === event.id)) {
                    return previousEvents;
                }

                const next = [...previousEvents, event];
                if (next.length <= MAX_RENDERED_EVENTS) {
                    return next;
                }

                return next.slice(next.length - MAX_RENDERED_EVENTS);
            });
        };

        const loadDumps = async () => {
            const [resolvedChannel, collectorStatus, recentEvents] = await Promise.all([
                DumpEventChannelName(),
                GetCollectorStatus(),
                GetRecentEvents(MAX_RENDERED_EVENTS),
            ]);

            if (disposed) {
                return;
            }

            setChannelName(resolvedChannel);
            setStatus((previous) => (isSameCollectorStatus(previous, collectorStatus) ? previous : collectorStatus));
            setEvents(recentEvents);

            unsubscribe = Events.On(resolvedChannel, (event) => {
                appendEvent(event.data as DumpEvent);
            });
        };

        void loadDumps();

        const interval = setInterval(() => {
            void GetCollectorStatus().then((nextStatus) => {
                if (!disposed) {
                    setStatus((previous) => (isSameCollectorStatus(previous, nextStatus) ? previous : nextStatus));
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

            Events.OffAll();
        };
    }, [location.pathname]);

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

    const refreshValetSites = async () => {
        setLoadingValetSites(true);
        try {
            const result = await GetValetSites();
            setValetSites(result);
        } finally {
            setLoadingValetSites(false);
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
                <Route
                    path="/sites"
                    element={(
                        <ValetSitesPage
                            valetSites={valetSites}
                            loadingValetSites={loadingValetSites}
                            onRefresh={refreshValetSites}
                        />
                    )}
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
