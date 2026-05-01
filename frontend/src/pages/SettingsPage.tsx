import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useTheme } from "@/components/theme-provider";
import { PageHeader } from "@/components/layout/PageHeader";
import { SetupPage } from "@/pages/SetupPage";
import { ValetPage } from "@/pages/ValetPage";
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

    const themeButtonClass = (isActive: boolean) => (
        isActive
            ? "min-w-24 border-primary bg-primary text-primary-foreground"
            : "min-w-24 border-border/60 bg-transparent text-muted-foreground hover:border-primary/60 hover:text-primary"
    );

    return (
        <div className="space-y-6">
            <PageHeader
                title="Settings"
                watermark="CFG"
                description="Manage diagnostics, Valet behavior, integrations, and appearance."
            />

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

            <Card>
                <CardHeader>
                    <CardTitle>Integrations</CardTitle>
                    <CardDescription>Connections with external tools and services.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="space-y-2">
                        <Label htmlFor="license-key">License key</Label>
                        <div className="flex flex-col gap-2 sm:flex-row">
                            <Input
                                id="license-key"
                                value={licenseKey}
                                onChange={(event) => onLicenseKeyChange(event.target.value)}
                                placeholder="PHANT-XXXX-XXXX-XXXX"
                            />
                            <Button onClick={onSaveLicense}>Save</Button>
                        </div>
                        <p className="text-xs text-muted-foreground">
                            Used for auto-update eligibility and to support Phant development.
                        </p>
                    </div>
                    <p className="text-sm text-muted-foreground">
                        Integrations are optional connections (for example editors, notifications, or tunnel/share tools).
                        This section is reserved for upcoming integration toggles.
                    </p>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Updates</CardTitle>
                    <CardDescription>Check for new releases and fetch the latest Linux AppImage.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <p className="text-sm text-muted-foreground">Current version: {appVersion || "unknown"}</p>

                    <div className="flex flex-wrap gap-2">
                        <Button variant="outline" onClick={onCheckForUpdates} disabled={checkingForUpdates}>
                            {checkingForUpdates ? "Checking..." : "Check for updates"}
                        </Button>
                        <Button onClick={onDownloadUpdate} disabled={downloadingUpdate || !updateStatus?.updateAvailable}>
                            {downloadingUpdate ? "Downloading..." : "Download latest"}
                        </Button>
                        <Button
                            variant="secondary"
                            onClick={onInstallUpdate}
                            disabled={installingUpdate || !updateDownloadResult?.downloaded || !updateDownloadResult.filePath}
                        >
                            {installingUpdate ? "Installing..." : "Install & restart"}
                        </Button>
                    </div>

                    {updateStatus?.error ? (
                        <p className="text-sm text-destructive">{updateStatus.error}</p>
                    ) : null}

                    {updateStatus && !updateStatus.error ? (
                        <div className="space-y-1 text-sm text-muted-foreground">
                            <p>Latest version: {updateStatus.latestVersion || "unknown"}</p>
                            <p>Status: {updateStatus.updateAvailable ? "Update available" : "Up to date"}</p>
                            {updateStatus.notes ? <p>Notes: {updateStatus.notes}</p> : null}
                        </div>
                    ) : null}

                    {updateDownloadResult?.error ? (
                        <p className="text-sm text-destructive">{updateDownloadResult.error}</p>
                    ) : null}

                    {updateDownloadResult?.downloaded ? (
                        <div className="space-y-1 text-sm text-emerald-500">
                            <p>Update downloaded successfully.</p>
                            <p className="text-muted-foreground">File: {updateDownloadResult.filePath}</p>
                            <p className="text-muted-foreground">Bytes: {updateDownloadResult.bytesWritten}</p>
                        </div>
                    ) : null}

                    {updateInstallResult?.error ? (
                        <p className="text-sm text-destructive">{updateInstallResult.error}</p>
                    ) : null}

                    {updateInstallResult?.installed ? (
                        <div className="space-y-1 text-sm text-emerald-500">
                            <p>{updateInstallResult.message || "Update installation started."}</p>
                            {updateInstallResult.targetPath ? (
                                <p className="text-muted-foreground">Target: {updateInstallResult.targetPath}</p>
                            ) : null}
                        </div>
                    ) : null}
                </CardContent>
            </Card>

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
        </div>
    );
}
