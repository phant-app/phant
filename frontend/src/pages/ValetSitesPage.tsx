import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { ActionButton } from "@/components/ui/action-button";
import { Github } from "lucide-react";
import type { ValetSitesResult } from "@/types";

export function ValetSitesPage({
    valetSites,
    loadingValetSites,
    onRefresh,
}: {
    valetSites: ValetSitesResult | null;
    loadingValetSites: boolean;
    onRefresh: () => void;
}) {
    const sites = valetSites?.sites ?? [];
    const parkedDirectories = valetSites?.parkedDirectories ?? [];
    const supported = valetSites?.supported ?? true;

    return (
        <div className="space-y-6">
            <div>
                <div className="flex items-center justify-between gap-3">
                    <h1 className="text-3xl font-bold tracking-tight">Sites</h1>
                    <a
                        href="https://github.com/cpriego/valet-linux"
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-2 rounded-md border border-slate-700 px-3 py-2 text-xs text-slate-300 transition-colors hover:bg-slate-800 hover:text-slate-100"
                    >
                        <Github className="h-4 w-4" />
                        cpriego/valet-linux
                    </a>
                </div>
                <p className="text-muted-foreground mt-2">Discover your current sites through your Valet</p>
            </div>

            {valetSites?.warnings?.length ? (
                <div className="rounded-md border border-amber-600/40 bg-amber-500/10 p-3 text-sm text-amber-200">
                    <ul className="list-inside list-disc space-y-1">
                        {valetSites.warnings.map((warning, index) => (
                            <li key={`${index}-${warning}`}>{warning}</li>
                        ))}
                    </ul>
                </div>
            ) : null}

            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between gap-3">
                        <div className="space-y-1">
                            <CardTitle>Linked Sites</CardTitle>
                            <CardDescription>Source: `valet links`.</CardDescription>
                        </div>
                        <ActionButton onClick={onRefresh} disabled={loadingValetSites}>
                            {loadingValetSites ? 'Refreshing...' : 'Refresh'}
                        </ActionButton>
                    </div>
                </CardHeader>
                <CardContent>
                    {!supported ? (
                        <p className="text-sm text-muted-foreground">
                            Valet sites discovery is not implemented for `{valetSites?.os || 'this OS'}` yet.
                        </p>
                    ) : null}

                    {valetSites?.error ? (
                        <p className="text-sm text-red-400">
                            {valetSites.error}
                        </p>
                    ) : null}

                    {supported && !valetSites?.error && !sites.length ? (
                        <p className="text-sm text-muted-foreground">
                            {loadingValetSites ? 'Loading linked sites...' : 'No linked sites found.'}
                        </p>
                    ) : null}

                    {sites.length ? (
                        <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Site</TableHead>
                                <TableHead>URL</TableHead>
                                <TableHead>PHP</TableHead>
                                <TableHead>SSL</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {sites.map((site) => (
                                <TableRow key={site.name}>
                                    <TableCell>
                                        <p className="font-medium">{site.name}</p>
                                        <p className="text-xs text-muted-foreground">{site.path}</p>
                                    </TableCell>
                                    <TableCell>
                                        <a href={site.url} target="_blank" rel="noreferrer" className="text-primary hover:underline">
                                            {site.url}
                                        </a>
                                    </TableCell>
                                    <TableCell>
                                        {site.phpVersion ? (
                                            <Badge variant="outline">{site.phpVersion}</Badge>
                                        ) : (
                                            <Badge variant="secondary">Unknown</Badge>
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        {site.isSecure ? (
                                            <Badge variant="default" className="bg-green-600/20 text-green-500 hover:bg-green-600/30">Secured</Badge>
                                        ) : (
                                            <Badge variant="secondary">None</Badge>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                    ) : null}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>Directories Parked</CardTitle>
                    <CardDescription>Source: `valet paths`.</CardDescription>
                </CardHeader>
                <CardContent>
                    {!supported ? (
                        <p className="text-sm text-muted-foreground">
                            Valet paths discovery is not implemented for `{valetSites?.os || 'this OS'}` yet.
                        </p>
                    ) : null}

                    {supported && !valetSites?.error && !parkedDirectories.length ? (
                        <p className="text-sm text-muted-foreground">
                            {loadingValetSites ? 'Loading parked directories...' : 'No parked directories found.'}
                        </p>
                    ) : null}

                    {parkedDirectories.length ? (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Directory</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {parkedDirectories.map((directory) => (
                                    <TableRow key={directory}>
                                        <TableCell>
                                            <p className="font-mono text-xs text-muted-foreground">{directory}</p>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    ) : null}
                </CardContent>
            </Card>
        </div>
    );
}
