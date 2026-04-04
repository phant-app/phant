import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { CheckCircle2 } from "lucide-react";
import { PageHeader } from "@/components/layout/PageHeader";
import {
    GetPHPManagerSnapshot,
    InstallPHPVersion,
    SetPHPExtensionState,
    SwitchPHPVersion,
    UpdatePHPIniSettings,
} from "../../bindings/phant/internal/services/phpservice";
import type { PHPIniSettings, PHPManagerSnapshot } from "@/types";

const initialSettings: PHPIniSettings = {
    uploadMaxFilesize: "",
    postMaxSize: "",
    memoryLimit: "",
    maxExecutionTime: "",
};

export function PhpManagerPage() {
    const [snapshot, setSnapshot] = useState<PHPManagerSnapshot | null>(null);
    const [loading, setLoading] = useState(true);
    const [workingVersion, setWorkingVersion] = useState<string | null>(null);
    const [savingSettings, setSavingSettings] = useState(false);
    const [settings, setSettings] = useState<PHPIniSettings>(initialSettings);
    const [workingExtension, setWorkingExtension] = useState<string | null>(null);

    const loadSnapshot = async () => {
        setLoading(true);
        try {
            const value = await GetPHPManagerSnapshot();
            setSnapshot(value as unknown as PHPManagerSnapshot);
            setSettings((value as unknown as PHPManagerSnapshot).settings ?? initialSettings);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        void loadSnapshot();
    }, []);

    const reportAction = (message: string, error: string, requiresPrivilege: boolean, suggestedCommands: string[]) => {
        if (error) {
            toast.error(error);
            if (requiresPrivilege && suggestedCommands.length > 0) {
                toast.info(`Try: ${suggestedCommands[0]}`);
            }
            return;
        }

        toast.success(message || "Action completed");
    };

    const installVersion = async (version: string) => {
        setWorkingVersion(version);
        try {
            const result = await InstallPHPVersion(version);
            reportAction(result.message, result.error, result.requiresPrivilege, result.suggestedCommands || []);
            await loadSnapshot();
        } finally {
            setWorkingVersion(null);
        }
    };

    const switchVersion = async (version: string) => {
        setWorkingVersion(version);
        try {
            const result = await SwitchPHPVersion(version);
            reportAction(result.message, result.error, result.requiresPrivilege, result.suggestedCommands || []);
            await loadSnapshot();
        } finally {
            setWorkingVersion(null);
        }
    };

    const saveSettings = async () => {
        setSavingSettings(true);
        try {
            const result = await UpdatePHPIniSettings(settings);
            reportAction(result.message, result.error, result.requiresPrivilege, result.suggestedCommands || []);
            await loadSnapshot();
        } finally {
            setSavingSettings(false);
        }
    };

    const toggleExtension = async (name: string, enabled: boolean) => {
        setWorkingExtension(name);
        try {
            const result = await SetPHPExtensionState({ name, enabled });
            reportAction(result.message, result.error, result.requiresPrivilege, result.suggestedCommands || []);
            await loadSnapshot();
        } finally {
            setWorkingExtension(null);
        }
    };

    const versions = snapshot?.versions ?? [];
    const extensions = snapshot?.extensions ?? [];
    const activeVersion = snapshot?.activeVersion || "none";

    const updateSettings = (key: keyof PHPIniSettings, value: string) => {
        setSettings((previous) => ({
            ...previous,
            [key]: value,
        }));
    };

    return (
        <div className="space-y-6">
            <PageHeader
                title="PHP"
                watermark="PHP"
                description="Manage installed PHP versions and switch the active globally linked version used by Valet."
                meta={`Active version: PHP ${activeVersion}`}
            />

            {loading ? <p className="text-sm text-muted-foreground">Loading PHP manager...</p> : null}
            {snapshot?.lastError ? <p className="text-sm text-destructive">{snapshot.lastError}</p> : null}
            {(snapshot?.warnings || []).map((warning) => (
                <p key={warning} className="text-xs text-muted-foreground">{warning}</p>
            ))}

            <Card>
                <CardHeader>
                    <CardTitle>Available Versions</CardTitle>
                    <CardDescription>
                        Install or switch PHP versions from your Linux package manager. The active CLI version is highlighted.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Version</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Actions</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {versions.map((v) => (
                                <TableRow
                                    key={v.version}
                                    className={v.active ? "bg-primary/5 hover:bg-primary/10" : undefined}
                                >
                                    <TableCell className={v.active ? "font-semibold text-primary" : "font-medium"}>
                                        <span className="inline-flex items-center gap-2">
                                            {v.active ? <CheckCircle2 className="h-4 w-4" aria-hidden="true" /> : null}
                                            <span>{v.version}</span>
                                        </span>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex gap-2">
                                            {v.installed ? (
                                                <Badge variant="secondary">Installed</Badge>
                                            ) : (
                                                <Badge variant="outline">Not Installed</Badge>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-right">
                                        {!v.installed ? (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                className="w-28 justify-center"
                                                disabled={workingVersion !== null}
                                                onClick={() => installVersion(v.version)}
                                            >
                                                {workingVersion === v.version ? "Installing..." : "Install"}
                                            </Button>
                                        ) : v.active ? (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                disabled
                                                className="w-28 justify-center border-border/40 bg-muted/40 text-muted-foreground"
                                            >
                                                Current
                                            </Button>
                                        ) : (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                className="w-28 justify-center"
                                                disabled={workingVersion !== null}
                                                onClick={() => switchVersion(v.version)}
                                            >
                                                {workingVersion === v.version ? "Switching..." : "Switch"}
                                            </Button>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>PHP Settings</CardTitle>
                    <CardDescription>Apply values to CLI and detected PHP-FPM targets.</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="grid gap-4 md:grid-cols-2">
                        <div className="space-y-2">
                            <Label htmlFor="upload_max_filesize">upload_max_filesize</Label>
                            <Input
                                id="upload_max_filesize"
                                value={settings.uploadMaxFilesize}
                                onChange={(event) => updateSettings("uploadMaxFilesize", event.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="post_max_size">post_max_size</Label>
                            <Input
                                id="post_max_size"
                                value={settings.postMaxSize}
                                onChange={(event) => updateSettings("postMaxSize", event.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="memory_limit">memory_limit</Label>
                            <Input
                                id="memory_limit"
                                value={settings.memoryLimit}
                                onChange={(event) => updateSettings("memoryLimit", event.target.value)}
                            />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="max_execution_time">max_execution_time</Label>
                            <Input
                                id="max_execution_time"
                                value={settings.maxExecutionTime}
                                onChange={(event) => updateSettings("maxExecutionTime", event.target.value)}
                            />
                        </div>
                    </div>
                    <Button onClick={saveSettings} disabled={savingSettings || loading}>
                        {savingSettings ? "Saving..." : "Save Settings"}
                    </Button>
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Extensions</CardTitle>
                    <CardDescription>Enable or disable PHP extensions for installed versions.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Extension</TableHead>
                                <TableHead>Status</TableHead>
                                <TableHead className="text-right">Action</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {extensions.map((extension) => (
                                <TableRow key={extension.name}>
                                    <TableCell className="font-medium">{extension.name}</TableCell>
                                    <TableCell>
                                        {extension.enabled
                                            ? <Badge variant="secondary">Enabled</Badge>
                                            : <Badge variant="outline">Disabled</Badge>}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <Button
                                            variant={extension.enabled ? "outline" : "secondary"}
                                            size="sm"
                                            className="w-28 justify-center"
                                            disabled={workingExtension !== null}
                                            onClick={() => toggleExtension(extension.name, !extension.enabled)}
                                        >
                                            {workingExtension === extension.name
                                                ? "Applying..."
                                                : extension.enabled ? "Disable" : "Enable"}
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
}
