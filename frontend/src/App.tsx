import { useEffect, useState } from 'react';
import { Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import { BaseLayout } from './components/layout/BaseLayout';
import { Events } from '@wailsio/runtime';
import { toast } from "sonner";
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
import {
    GetLicenseKey,
    SaveLicenseKey,
} from '../bindings/phant/internal/services/licenseservice';
import {
    CheckForUpdate,
    CurrentVersion,
    DownloadLatest,
} from '../bindings/phant/internal/services/updateservice';

import type {
    CollectorStatus,
    DumpEvent,
    HookInstallResult,
    SetupDiagnostics,
    UpdateCheckResult,
    UpdateDownloadResult,
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
import { ValetPage } from './pages/ValetPage';
import { OnboardingPage } from './pages/OnboardingPage';

const MAX_RENDERED_EVENTS = 500;
const ONBOARDING_COMPLETED_KEY = 'phant:onboarding:v1:completed';
const ONBOARDING_SEEN_LEGACY_KEY = 'phant:onboarding:v1:seen';
const LEGACY_LICENSE_KEY_STORAGE = 'phant:license:v1:key';
const ONBOARDING_PATH = '/onboarding';

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
    const navigate = useNavigate();
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
    const [installingFPMHook, setInstallingFPMHook] = useState(false);
    const [confirmValetRemediation, setConfirmValetRemediation] = useState(false);
    const [onboardingReady, setOnboardingReady] = useState(false);
    const [onboardingCompleted, setOnboardingCompleted] = useState(false);
    const [licenseKey, setLicenseKey] = useState('');
    const [updateStatus, setUpdateStatus] = useState<UpdateCheckResult | null>(null);
    const [updateDownloadResult, setUpdateDownloadResult] = useState<UpdateDownloadResult | null>(null);
    const [checkingForUpdates, setCheckingForUpdates] = useState(false);
    const [downloadingUpdate, setDownloadingUpdate] = useState(false);

    useEffect(() => {
        let disposed = false;

        const loadOnboardingState = async () => {
            const completedOnboarding = window.localStorage.getItem(ONBOARDING_COMPLETED_KEY) === 'true'
                || window.localStorage.getItem(ONBOARDING_SEEN_LEGACY_KEY) === 'true';

            let resolvedLicenseKey = '';
            try {
                const result = await GetLicenseKey();
                if (result.error) {
                    console.error('Failed to load license key from backend:', result.error);
                } else {
                    resolvedLicenseKey = result.licenseKey || '';
                }

                if (!resolvedLicenseKey) {
                    const legacyLicenseKey = (window.localStorage.getItem(LEGACY_LICENSE_KEY_STORAGE) || '').trim();
                    if (legacyLicenseKey.length > 0) {
                        const saveResult = await SaveLicenseKey(legacyLicenseKey);
                        if (saveResult.success) {
                            resolvedLicenseKey = saveResult.licenseKey;
                            window.localStorage.removeItem(LEGACY_LICENSE_KEY_STORAGE);
                        } else if (saveResult.error) {
                            console.error('Failed to migrate legacy license key:', saveResult.error);
                        }
                    }
                }
            } catch (error) {
                console.error('Failed to initialize onboarding state:', error);
            }

            if (disposed) {
                return;
            }

            setOnboardingCompleted(completedOnboarding);
            setLicenseKey(resolvedLicenseKey);
            setOnboardingReady(true);
        };

        void loadOnboardingState();

        return () => {
            disposed = true;
        };
    }, []);

    const checkForUpdates = async (silent = false) => {
        setCheckingForUpdates(true);
        try {
            const [currentVersion, checkResult] = await Promise.all([
                CurrentVersion(),
                CheckForUpdate(""),
            ]);
            const normalized: UpdateCheckResult = {
                currentVersion: checkResult.currentVersion || currentVersion || "unknown",
                latestVersion: checkResult.latestVersion || "",
                updateAvailable: Boolean(checkResult.updateAvailable),
                downloadURL: checkResult.downloadURL || "",
                notes: checkResult.notes || "",
                error: checkResult.error || "",
            };
            setUpdateStatus(normalized);

            if (!silent) {
                if (normalized.error) {
                    toast.error(normalized.error);
                } else if (normalized.updateAvailable) {
                    toast.info(`New update available: ${normalized.latestVersion}`);
                } else {
                    toast.success("Phant is up to date.");
                }
            } else if (!normalized.error && normalized.updateAvailable) {
                toast.info(`New update available: ${normalized.latestVersion}`);
            }
        } finally {
            setCheckingForUpdates(false);
        }
    };

    const downloadUpdate = async () => {
        setDownloadingUpdate(true);
        try {
            const result = await DownloadLatest("");
            const normalized: UpdateDownloadResult = {
                currentVersion: result.currentVersion || "",
                latestVersion: result.latestVersion || "",
                updateAvailable: Boolean(result.updateAvailable),
                downloaded: Boolean(result.downloaded),
                filePath: result.filePath || "",
                finalURL: result.finalURL || "",
                statusCode: result.statusCode || 0,
                bytesWritten: result.bytesWritten || 0,
                notes: result.notes || "",
                error: result.error || "",
            };
            setUpdateDownloadResult(normalized);
            if (normalized.error) {
                toast.error(normalized.error);
            } else if (normalized.downloaded) {
                toast.success("Update downloaded. Restart after replacing the executable.");
            } else {
                toast.info("No new update to download.");
            }
        } finally {
            setDownloadingUpdate(false);
        }
    };

    useEffect(() => {
        if (!onboardingReady) {
            return;
        }

        if (!onboardingCompleted && location.pathname !== ONBOARDING_PATH) {
            navigate(ONBOARDING_PATH, { replace: true });
            return;
        }

        if (onboardingCompleted && location.pathname === ONBOARDING_PATH) {
            navigate('/dumps', { replace: true });
        }
    }, [location.pathname, navigate, onboardingCompleted, onboardingReady]);

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

    const completeOnboarding = () => {
        window.localStorage.setItem(ONBOARDING_COMPLETED_KEY, 'true');
        window.localStorage.removeItem(ONBOARDING_SEEN_LEGACY_KEY);
        setOnboardingCompleted(true);
    };

    const saveLicenseFromOnboarding = () => {
        const normalized = licenseKey.trim();
        void (async () => {
            try {
                const result = await SaveLicenseKey(normalized);
                if (result.error) {
                    console.error('Failed to save license key:', result.error);
                    return;
                }
                setLicenseKey(result.licenseKey);
            } catch (error) {
                console.error('Failed to save license key:', error);
            }
        })();
    };

    const runHookSetupFromOnboarding = async () => {
        await enableCLIHook();
    };

    const runFPMSetupFromOnboarding = async () => {
        setInstallingFPMHook(true);
        try {
            const result = await ApplyValetLinuxRemediation(true);
            setValetRemediationResult(result);
            const latestValetStatus = await GetValetLinuxVerification();
            setValetVerification(latestValetStatus);
        } finally {
            setInstallingFPMHook(false);
        }
    };

    useEffect(() => {
        if (!onboardingReady || !onboardingCompleted) {
            return;
        }
        void checkForUpdates(true);
    }, [onboardingReady, onboardingCompleted]);

    if (!onboardingReady) {
        return null;
    }

    if (!onboardingCompleted && location.pathname === ONBOARDING_PATH) {
        return (
            <ThemeProvider defaultTheme="dark" storageKey="phant-ui-theme">
                <OnboardingPage
                    diagnostics={diagnostics}
                    valetVerification={valetVerification}
                    hookResult={hookResult}
                    fpmHookResult={valetRemediationResult}
                    installingHook={installingHook}
                    installingFPMHook={installingFPMHook}
                    licenseKey={licenseKey}
                    onLicenseKeyChange={setLicenseKey}
                    onSetupHook={runHookSetupFromOnboarding}
                    onSetupFPMHook={runFPMSetupFromOnboarding}
                    onSaveLicense={saveLicenseFromOnboarding}
                    onComplete={completeOnboarding}
                />
            </ThemeProvider>
        );
    }

    return (
        <ThemeProvider defaultTheme="dark" storageKey="phant-ui-theme">
            <BaseLayout>
                <Routes>
                    <Route path="/" element={<Navigate to="/dumps" replace />} />
                    <Route path={ONBOARDING_PATH} element={<Navigate to="/dumps" replace />} />
                    <Route
                        path="/php"
                        element={<PhpManagerPage />}
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
                <Route
                    path="/settings"
                    element={(
                        <SettingsPage
                            diagnostics={diagnostics}
                            hookResult={hookResult}
                            installingHook={installingHook}
                            onRefreshDiagnostics={refreshDiagnostics}
                            onEnableCLIHook={enableCLIHook}
                            licenseKey={licenseKey}
                            onLicenseKeyChange={setLicenseKey}
                            onSaveLicense={saveLicenseFromOnboarding}
                            updateStatus={updateStatus}
                            updateDownloadResult={updateDownloadResult}
                            checkingForUpdates={checkingForUpdates}
                            downloadingUpdate={downloadingUpdate}
                            onCheckForUpdates={() => { void checkForUpdates(false); }}
                            onDownloadUpdate={() => { void downloadUpdate(); }}
                            valetVerification={valetVerification}
                            refreshingValet={refreshingValet}
                            onRefreshValet={refreshValetVerification}
                            confirmValetRemediation={confirmValetRemediation}
                            onConfirmValetRemediation={setConfirmValetRemediation}
                            applyingValetRemediation={applyingValetRemediation}
                            onApplyValetRemediation={applyValetRemediation}
                            valetRemediationResult={valetRemediationResult}
                        />
                    )}
                />
                <Route path="*" element={<Navigate to="/dumps" replace />} />
            </Routes>
        </BaseLayout>
        </ThemeProvider>
    );
}

export default App
