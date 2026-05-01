import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useTheme } from "@/components/theme-provider";
import { PageHeader } from "@/components/layout/PageHeader";
import { SetupPage } from "@/pages/SetupPage";
import { ValetPage } from "@/pages/ValetPage";
import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import type {
    HookInstallResult,
    SetupDiagnostics,
    UpdateCheckResult,
    UpdateDownloadResult,
    UpdateInstallResult,
    ValetLinuxRemediationResult,
    ValetLinuxVerification,
} from "@/types";

export function SettingsPage({
    diagnostics,
    hookResult,
    installingHook,
    onRefreshDiagnostics,
    onEnableCLIHook,
    licenseKey,
    onLicenseKeyChange,
    onSaveLicense,
    appVersion,
    updateStatus,
    checkingForUpdates,
    downloadingUpdate,
    installingUpdate,
    onCheckForUpdates,
    onDownloadUpdate,
    onInstallUpdate,
    valetVerification,
    refreshingValet,
    onRefreshValet,
    confirmValetRemediation,
    onConfirmValetRemediation,
    applyingValetRemediation,
    onApplyValetRemediation,
    valetRemediationResult,
    updateDownloadResult,
    updateInstallResult,
}: {
    diagnostics: SetupDiagnostics | null;
    hookResult: HookInstallResult | null;
    installingHook: boolean;
    onRefreshDiagnostics: () => void;
    onEnableCLIHook: () => void;
    licenseKey: string;
    onLicenseKeyChange: (value: string) => void;
    onSaveLicense: () => void;
    appVersion: string;
    updateStatus: UpdateCheckResult | null;
    checkingForUpdates: boolean;
    downloadingUpdate: boolean;
    installingUpdate: boolean;
    onCheckForUpdates: () => void;
    onDownloadUpdate: () => void;
    onInstallUpdate: () => void;
    valetVerification: ValetLinuxVerification | null;
    refreshingValet: boolean;
    onRefreshValet: () => void;
    confirmValetRemediation: boolean;
    onConfirmValetRemediation: (checked: boolean) => void;
    applyingValetRemediation: boolean;
    onApplyValetRemediation: () => void;
    valetRemediationResult: ValetLinuxRemediationResult | null;
    updateDownloadResult: UpdateDownloadResult | null;
    updateInstallResult: UpdateInstallResult | null;
    }) {
    const { theme, setTheme } = useTheme();
    const [showLicense, setShowLicense] = useState(false);

    const themeButtonClass = (isActive: boolean) => (
        isActive
            ? "min-w-24 border-primary bg-primary text-primary-foreground"
            : "min-w-24 border-border/60 bg-transparent text-muted-foreground hover:border-primary/60 hover:text-primary"
    );

    const currentVersion = updateStatus?.currentVersion || appVersion || "unknown";
    const latestVersion = updateStatus?.latestVersion || updateDownloadResult?.latestVersion || "unknown";
    const hasDownloadedUpdate = Boolean(updateDownloadResult?.downloaded && updateDownloadResult.filePath);
    const hasUpdateAvailable = Boolean(updateStatus?.updateAvailable);

    const updateState = (() => {
        if (updateInstallResult?.error || updateDownloadResult?.error || updateStatus?.error) {
            return { label: "Error", variant: "destructive" as const };
        }

        if (installingUpdate) {
            return { label: "Installing", variant: "secondary" as const };
        }

        if (updateInstallResult?.installed) {
            return { label: "Installing", variant: "default" as const };
        }

        if (downloadingUpdate) {
            return { label: "Downloading", variant: "secondary" as const };
        }

        if (hasDownloadedUpdate) {
            return { label: "Ready to install", variant: "default" as const };
        }

        if (checkingForUpdates) {
            return { label: "Checking", variant: "secondary" as const };
        }

        if (hasUpdateAvailable) {
            return { label: "Update available", variant: "default" as const };
        }

        if (updateStatus && !updateStatus.error) {
            return { label: "Up to date", variant: "outline" as const };
        }

        return { label: "Not checked", variant: "outline" as const };
    })();

    const releaseNotes = updateDownloadResult?.notes || updateStatus?.notes || "";

    return (
        <div className="space-y-6">
            <PageHeader
                title="Settings"
                watermark="CFG"
                description="Manage appearance, licensing, updates, diagnostics, and recovery tools."
            />

            <Card>
                <CardHeader>
                    <CardTitle>Appearance</CardTitle>
                    <CardDescription>Customize the look and feel of Phant.</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex flex-wrap gap-4">
                        <Button
                            variant="outline"
                            className={themeButtonClass(theme === "light")}
                            onClick={() => setTheme("light")}
                        >
                            Light
                        </Button>
                        <Button
                            variant="outline"
                            className={themeButtonClass(theme === "dark")}
                            onClick={() => setTheme("dark")}
                        >
                            Dark
                        </Button>
                        <Button
                            variant="outline"
                            className={themeButtonClass(theme === "system")}
                            onClick={() => setTheme("system")}
                        >
                            System
                        </Button>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>License</CardTitle>
                    <CardDescription>Activate Phant and keep your updates eligible.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="license-key">License key</Label>
                        <div className="relative">
                            <Input
                                id="license-key"
                                type={showLicense ? "text" : "password"}
                                value={licenseKey}
                                onChange={(event) => onLicenseKeyChange(event.target.value)}
                                placeholder="PHANT-XXXX-XXXX-XXXX"
                                className="pr-24"
                            />
                            <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                onClick={() => setShowLicense((value) => !value)}
                                className="absolute inset-y-0 right-0 h-full px-3 text-muted-foreground hover:text-foreground"
                            >
                                {showLicense ? <EyeOff /> : <Eye />}
                                <span className="ml-2">{showLicense ? "Hide" : "Show"}</span>
                            </Button>
                        </div>
                        <div className="flex items-center justify-between gap-2">
                            <p className="text-xs text-muted-foreground">
                                Used for auto-update eligibility and to support Phant development.
                            </p>
                            <Button onClick={onSaveLicense}>Save</Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <CardTitle>Updates</CardTitle>
                        <Badge variant={updateState.variant}>{updateState.label}</Badge>
                    </div>
                    <CardDescription>Check for new releases and fetch the latest Linux AppImage.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid gap-3 rounded-lg border border-border/60 bg-muted/20 p-4 sm:grid-cols-2">
                        <div className="space-y-1">
                            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">Current</p>
                            <p className="text-lg font-semibold">{currentVersion}</p>
                        </div>
                        <div className="space-y-1">
                            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">Latest</p>
                            <p className="text-lg font-semibold">{latestVersion}</p>
                        </div>
                    </div>

                    <div className="flex flex-wrap gap-2">
                        <Button variant={hasUpdateAvailable || hasDownloadedUpdate ? "outline" : "default"} onClick={onCheckForUpdates} disabled={checkingForUpdates}>
                            {checkingForUpdates ? "Checking..." : "Check for updates"}
                        </Button>
                        <Button onClick={onDownloadUpdate} disabled={downloadingUpdate || !hasUpdateAvailable}>
                            {downloadingUpdate ? "Downloading..." : `Download ${latestVersion}`}
                        </Button>
                        {hasDownloadedUpdate ? (
                            <Button variant="secondary" onClick={onInstallUpdate} disabled={installingUpdate}>
                                {installingUpdate ? "Installing..." : "Install & restart"}
                            </Button>
                        ) : null}
                    </div>

                    {updateStatus?.error ? (
                        <p className="text-sm text-destructive">{updateStatus.error}</p>
                    ) : null}

                    {updateDownloadResult?.error ? (
                        <p className="text-sm text-destructive">{updateDownloadResult.error}</p>
                    ) : null}

                    {updateDownloadResult?.downloaded ? (
                        <div className="space-y-1 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4 text-sm text-emerald-500">
                            <p className="font-medium">Update downloaded successfully.</p>
                            <p className="text-muted-foreground">File: {updateDownloadResult.filePath}</p>
                            <p className="text-muted-foreground">Bytes: {updateDownloadResult.bytesWritten}</p>
                        </div>
                    ) : null}

                    {updateInstallResult?.error ? (
                        <p className="text-sm text-destructive">{updateInstallResult.error}</p>
                    ) : null}

                    {updateInstallResult?.installed ? (
                        <div className="space-y-1 rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-4 text-sm text-emerald-500">
                            <p className="font-medium">{updateInstallResult.message || "Update installation started."}</p>
                            {updateInstallResult.targetPath ? (
                                <p className="text-muted-foreground">Target: {updateInstallResult.targetPath}</p>
                            ) : null}
                        </div>
                    ) : null}

                    {releaseNotes ? (
                        <div className="space-y-2 rounded-lg border border-border/60 bg-muted/20 p-4">
                            <p className="text-xs font-medium uppercase tracking-[0.18em] text-muted-foreground">What&apos;s new</p>
                            <p className="text-sm text-muted-foreground">{releaseNotes}</p>
                        </div>
                    ) : null}
                </CardContent>
            </Card>

            <div className="space-y-2">
                <h2 className="text-xl font-semibold">Diagnostics</h2>
                <p className="text-sm text-muted-foreground">Inspect PHP CLI health and manage hook installation.</p>
                <SetupPage
                    embedded
                    diagnostics={diagnostics}
                    hookResult={hookResult}
                    installingHook={installingHook}
                    onRefresh={onRefreshDiagnostics}
                    onEnable={onEnableCLIHook}
                />
            </div>

            <div className="space-y-2">
                <h2 className="text-xl font-semibold">Valet</h2>
                <p className="text-sm text-muted-foreground">Diagnose Valet Linux and apply remediation safely.</p>
                <ValetPage
                    embedded
                    valetVerification={valetVerification}
                    refreshingValet={refreshingValet}
                    onRefresh={onRefreshValet}
                    confirmValetRemediation={confirmValetRemediation}
                    onConfirm={onConfirmValetRemediation}
                    applyingValetRemediation={applyingValetRemediation}
                    onApply={onApplyValetRemediation}
                    valetRemediationResult={valetRemediationResult}
                />
            </div>
        </div>
    );
}
