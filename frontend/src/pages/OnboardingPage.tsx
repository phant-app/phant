import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { HookInstallResult, SetupDiagnostics, ValetLinuxRemediationResult, ValetLinuxVerification } from "@/types";

type OnboardingStep = "welcome" | "hooks" | "license";

const steps: Array<{ id: OnboardingStep; title: string; subtitle: string }> = [
    {
        id: "welcome",
        title: "Welcome",
        subtitle: "What Phant does and what we will set up in under a minute.",
    },
    {
        id: "hooks",
        title: "PHP Hooks",
        subtitle: "Install CLI and PHP-FPM prepend hooks so dump capture works in both CLI and Valet HTTP requests.",
    },
    {
        id: "license",
        title: "License",
        subtitle: "Save your key for future auto-update and project support.",
    },
];

const stepIndexById: Record<OnboardingStep, number> = {
    welcome: 0,
    hooks: 1,
    license: 2,
};

export function OnboardingPage({
    diagnostics,
    valetVerification,
    hookResult,
    fpmHookResult,
    installingHook,
    installingFPMHook,
    licenseKey,
    onLicenseKeyChange,
    onSetupHook,
    onSetupFPMHook,
    onSaveLicense,
    onComplete,
}: {
    diagnostics: SetupDiagnostics | null;
    valetVerification: ValetLinuxVerification | null;
    hookResult: HookInstallResult | null;
    fpmHookResult: ValetLinuxRemediationResult | null;
    installingHook: boolean;
    installingFPMHook: boolean;
    licenseKey: string;
    onLicenseKeyChange: (value: string) => void;
    onSetupHook: () => Promise<void>;
    onSetupFPMHook: () => Promise<void>;
    onSaveLicense: () => void;
    onComplete: () => void;
}) {
    const [step, setStep] = useState<OnboardingStep>("welcome");

    const cliHookEnabled = Boolean(
        hookResult?.success
        || (valetVerification?.cliAutoPrepend
            && valetVerification.expectedPrependPath
            && valetVerification.cliAutoPrepend === valetVerification.expectedPrependPath),
    );

    const fpmHookEnabled = Boolean(
        valetVerification?.fpmServices?.length
        && valetVerification.fpmServices.every((service) => service.matchesExpected),
    );

    const cliInfoMessage = (() => {
        if (hookResult?.error) {
            return { tone: "error", text: hookResult.error } as const;
        }

        if (hookResult?.success && !hookResult.alreadyEnabled) {
            return { tone: "success", text: "CLI hook installed successfully." } as const;
        }

        if (cliHookEnabled) {
            return { tone: "success", text: "Already enabled. No changes required." } as const;
        }

        return null;
    })();

    const fpmInfoMessage = (() => {
        if (fpmHookResult?.error) {
            return { tone: "error", text: fpmHookResult.error } as const;
        }

        if (fpmHookResult?.applied && fpmHookResult.message) {
            return { tone: "success", text: fpmHookResult.message } as const;
        }

        if (fpmHookEnabled) {
            return { tone: "success", text: "Already enabled. No changes required." } as const;
        }

        return null;
    })();

    const cliInstallDisabled = installingHook || (cliHookEnabled && !hookResult?.error);
    const fpmInstallDisabled = installingFPMHook || (fpmHookEnabled && !fpmHookResult?.error);

    const cliBadge = cliInfoMessage?.tone === "error"
        ? { text: "Error", className: "border-destructive/50 bg-destructive/10 text-destructive" }
        : cliHookEnabled
            ? { text: "Enabled", className: "border-emerald-500/50 bg-emerald-500/10 text-emerald-500" }
            : { text: "Pending", className: "border-amber-500/50 bg-amber-500/10 text-amber-500" };

    const fpmBadge = fpmInfoMessage?.tone === "error"
        ? { text: "Error", className: "border-destructive/50 bg-destructive/10 text-destructive" }
        : fpmHookEnabled
            ? { text: "Enabled", className: "border-emerald-500/50 bg-emerald-500/10 text-emerald-500" }
            : { text: "Pending", className: "border-amber-500/50 bg-amber-500/10 text-amber-500" };

    const progressPercent = useMemo(() => {
        const current = stepIndexById[step] + 1;
        return Math.round((current / steps.length) * 100);
    }, [step]);

    return (
        <div className="min-h-screen bg-background text-foreground">
            <div className="relative mx-auto flex min-h-screen w-full max-w-7xl flex-col gap-6 p-5 md:p-8">
                <div className="pointer-events-none absolute inset-0 opacity-35">
                    <div className="absolute -left-20 top-10 h-80 w-80 rounded-full bg-primary/20 blur-3xl" />
                    <div className="absolute bottom-0 right-0 h-96 w-96 rounded-full bg-emerald-500/10 blur-3xl" />
                </div>

                <header className="relative z-10 flex flex-col gap-4 border-2 border-border bg-card/80 p-5 backdrop-blur md:flex-row md:items-end md:justify-between">
                    <div>
                        <p className="font-mono text-xs uppercase tracking-[0.2em] text-primary">Quick Setup</p>
                        <h1 className="font-rock text-4xl uppercase leading-none md:text-5xl">Phant Onboarding</h1>
                        <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
                            One guided pass to get your local setup ready. You can still edit everything later in Settings.
                        </p>
                    </div>
                    <div className="w-full max-w-sm space-y-2">
                        <div className="flex items-center justify-between font-mono text-xs uppercase tracking-[0.16em] text-muted-foreground">
                            <span>Progress</span>
                            <span>{progressPercent}%</span>
                        </div>
                        <div className="h-2 border border-border bg-background">
                            <div className="h-full bg-primary transition-all duration-300" style={{ width: `${progressPercent}%` }} />
                        </div>
                    </div>
                </header>

                <div className="relative z-10 grid gap-6 lg:grid-cols-[1.1fr_1.6fr]">
                    <Card className="scanlines cut-corner-lg bg-card/85">
                        <CardHeader>
                            <CardTitle className="font-mono text-sm uppercase tracking-[0.18em]">Setup Status</CardTitle>
                            <CardDescription>Live status while you complete onboarding.</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-4 text-sm">
                            <div className="space-y-2">
                                {steps.map((item) => {
                                    const index = stepIndexById[item.id];
                                    const active = item.id === step;
                                    const done = stepIndexById[step] > index;

                                    return (
                                        <div
                                            key={item.id}
                                            className={`border px-3 py-2 transition-colors ${
                                                active
                                                    ? "border-primary bg-primary/10"
                                                    : done
                                                        ? "border-emerald-500/40 bg-emerald-500/10"
                                                        : "border-border bg-background/60"
                                            }`}
                                        >
                                            <p className="font-mono text-xs uppercase tracking-[0.14em]">
                                                Step {index + 1}: {item.title}
                                            </p>
                                            <p className="mt-1 text-xs text-muted-foreground">{item.subtitle}</p>
                                        </div>
                                    );
                                })}
                            </div>

                            <div className="border border-border bg-background/60 p-3 space-y-1">
                                <p className="font-mono text-xs uppercase tracking-[0.14em]">Environment Snapshot</p>
                                <p className="text-xs text-muted-foreground">PHP found: {diagnostics?.phpFound ? "yes" : "no"}</p>
                                <p className="text-xs text-muted-foreground">PHP version: {diagnostics?.phpVersion || "n/a"}</p>
                                <p className="text-xs text-muted-foreground">Hook: {cliHookEnabled ? "enabled" : "pending"}</p>
                                <p className="text-xs text-muted-foreground">
                                    FPM hook: {fpmHookEnabled ? "enabled" : "pending"}
                                </p>
                                <p className="text-xs text-muted-foreground">License key: {licenseKey.trim() ? "provided" : "missing"}</p>
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="cut-corner-lg bg-card/90">
                        <CardHeader>
                            <CardTitle className="font-rock text-3xl uppercase leading-none">
                                {step === "welcome" ? "Get Started" : null}
                                {step === "hooks" ? "Install PHP Hooks" : null}
                                {step === "license" ? "Save License Key" : null}
                            </CardTitle>
                            <CardDescription>
                                {step === "welcome" ? "Phant helps you manage PHP, services, sites, and diagnostics in one place." : null}
                                {step === "hooks" ? "This enables reliable dump capture and setup diagnostics for CLI commands." : null}
                                {step === "license" ? "You can edit this later in Settings > Integrations." : null}
                            </CardDescription>
                        </CardHeader>

                        <CardContent className="space-y-6">
                            {step === "welcome" ? (
                                <div className="grid gap-3 md:grid-cols-3">
                                    <div className="border border-border bg-background/60 p-3">
                                        <p className="font-mono text-xs uppercase tracking-[0.14em] text-primary">Services</p>
                                        <p className="mt-2 text-sm text-muted-foreground">Inspect running service state and detected ports from your machine.</p>
                                    </div>
                                    <div className="border border-border bg-background/60 p-3">
                                        <p className="font-mono text-xs uppercase tracking-[0.14em] text-primary">Sites</p>
                                        <p className="mt-2 text-sm text-muted-foreground">Discover local Valet sites and verify linked paths quickly.</p>
                                    </div>
                                    <div className="border border-border bg-background/60 p-3">
                                        <p className="font-mono text-xs uppercase tracking-[0.14em] text-primary">Dumps</p>
                                        <p className="mt-2 text-sm text-muted-foreground">Capture and inspect dump events while developing.</p>
                                    </div>
                                </div>
                            ) : null}

                            {step === "hooks" ? (
                                <div className="space-y-3">
                                    <div className="flex flex-col gap-2 border border-border bg-background/60 p-3 md:flex-row md:items-start md:justify-between">
                                        <div className="space-y-1 pr-2">
                                            <div className="flex items-center gap-2">
                                                <p className="font-mono text-xs uppercase tracking-[0.14em] text-primary">PHP CLI Hook</p>
                                                <Badge variant="outline" className={cliBadge.className}>{cliBadge.text}</Badge>
                                            </div>
                                            <p className="text-sm text-muted-foreground">
                                                Adds an auto-prepend file for CLI commands so Phant can capture dump()/dd() and diagnostics.
                                            </p>
                                            <p className="text-sm text-muted-foreground">
                                                Current status: {cliHookEnabled ? "Hook is enabled." : "Hook not installed yet."}
                                            </p>
                                            {cliInfoMessage ? (
                                                <p className={`text-sm ${cliInfoMessage.tone === "error" ? "text-destructive" : "text-emerald-500"}`}>
                                                    {cliInfoMessage.text}
                                                </p>
                                            ) : null}
                                        </div>
                                        <Button className="shrink-0" onClick={() => { void onSetupHook(); }} disabled={cliInstallDisabled}>
                                            {installingHook ? "Installing..." : (cliInstallDisabled ? "Installed" : "Install")}
                                        </Button>
                                    </div>

                                    <div className="flex flex-col gap-2 border border-border bg-background/60 p-3 md:flex-row md:items-start md:justify-between">
                                        <div className="space-y-1 pr-2">
                                            <div className="flex items-center gap-2">
                                                <p className="font-mono text-xs uppercase tracking-[0.14em] text-primary">PHP-FPM Hook (Valet Linux)</p>
                                                <Badge variant="outline" className={fpmBadge.className}>{fpmBadge.text}</Badge>
                                            </div>
                                            <p className="text-sm text-muted-foreground">
                                                Writes/validates the same auto-prepend file in detected PHP-FPM conf.d targets so HTTP requests via Valet are captured.
                                            </p>
                                            <p className="text-sm text-muted-foreground">
                                                Current status: {fpmHookEnabled ? "Hook is enabled." : "Hook not installed yet."}
                                            </p>
                                            {fpmInfoMessage ? (
                                                <p className={`text-sm ${fpmInfoMessage.tone === "error" ? "text-destructive" : "text-emerald-500"}`}>
                                                    {fpmInfoMessage.text}
                                                </p>
                                            ) : null}
                                        </div>
                                        <Button className="shrink-0" onClick={() => { void onSetupFPMHook(); }} disabled={fpmInstallDisabled}>
                                            {installingFPMHook ? "Installing..." : (fpmInstallDisabled ? "Installed" : "Install")}
                                        </Button>
                                    </div>
                                </div>
                            ) : null}

                            {step === "license" ? (
                                <div className="space-y-3 max-w-xl">
                                    <div className="space-y-2">
                                        <Label htmlFor="onboarding-license">License key</Label>
                                        <Input
                                            id="onboarding-license"
                                            placeholder="PHANT-XXXX-XXXX-XXXX"
                                            value={licenseKey}
                                            onChange={(event) => onLicenseKeyChange(event.target.value)}
                                        />
                                    </div>
                                    <p className="text-xs text-muted-foreground">
                                        Phant is free to use. A valid license key enables auto-update and supports project development.
                                    </p>
                                    <Button variant="outline" onClick={onSaveLicense}>Save Key</Button>
                                </div>
                            ) : null}

                            <div className="flex flex-wrap items-center gap-2 border-t border-border pt-4">
                                {step !== "welcome" ? (
                                    <Button variant="outline" onClick={() => setStep(steps[stepIndexById[step] - 1].id)}>
                                        Back
                                    </Button>
                                ) : null}

                                {step === "welcome" ? (
                                    <Button onClick={() => setStep("hooks")}>Continue</Button>
                                ) : null}

                                {step === "hooks" ? (
                                    <Button onClick={() => setStep("license")}>Continue</Button>
                                ) : null}

                                {step === "license" ? (
                                    <Button
                                        onClick={() => {
                                            onSaveLicense();
                                            onComplete();
                                        }}
                                    >
                                        Finish Setup
                                    </Button>
                                ) : null}
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}
