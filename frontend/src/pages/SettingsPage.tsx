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
    valetVerification,
    refreshingValet,
    onRefreshValet,
    confirmValetRemediation,
    onConfirmValetRemediation,
    applyingValetRemediation,
    onApplyValetRemediation,
    valetRemediationResult,
}: {
    diagnostics: SetupDiagnostics | null;
    hookResult: HookInstallResult | null;
    installingHook: boolean;
    onRefreshDiagnostics: () => void;
    onEnableCLIHook: () => void;
    licenseKey: string;
    onLicenseKeyChange: (value: string) => void;
    onSaveLicense: () => void;
    valetVerification: ValetLinuxVerification | null;
    refreshingValet: boolean;
    onRefreshValet: () => void;
    confirmValetRemediation: boolean;
    onConfirmValetRemediation: (checked: boolean) => void;
    applyingValetRemediation: boolean;
    onApplyValetRemediation: () => void;
    valetRemediationResult: ValetLinuxRemediationResult | null;
}) {
    const { theme, setTheme } = useTheme();

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
                    <CardTitle>Appearance</CardTitle>
                    <CardDescription>Customize the look and feel of Phant.</CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="flex flex-wrap gap-4">
                        <Button
                            variant={theme === "light" ? "default" : "outline"}
                            onClick={() => setTheme("light")}
                        >
                            Light
                        </Button>
                        <Button
                            variant={theme === "dark" ? "default" : "outline"}
                            onClick={() => setTheme("dark")}
                        >
                            Dark
                        </Button>
                        <Button
                            variant={theme === "system" ? "default" : "outline"}
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
