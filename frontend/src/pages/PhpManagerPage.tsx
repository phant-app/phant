import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const mockVersions = [
    { version: '8.3', installed: true, active: true },
    { version: '8.2', installed: true, active: false },
    { version: '8.1', installed: false, active: false },
    { version: '8.0', installed: false, active: false },
];

export function PhpManagerPage() {
    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">PHP</h1>
                <p className="text-muted-foreground mt-2">
                    Manage installed PHP versions. Switch the active globally linked PHP version used by Valet.
                </p>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Available Versions</CardTitle>
                    <CardDescription>Install or switch PHP versions from the ondrej/php PPA.</CardDescription>
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
                            {mockVersions.map((v) => (
                                <TableRow key={v.version}>
                                    <TableCell className="font-medium">PHP {v.version}</TableCell>
                                    <TableCell>
                                        <div className="flex gap-2">
                                            {v.installed ? (
                                                <Badge variant="secondary">Installed</Badge>
                                            ) : (
                                                <Badge variant="outline">Not Installed</Badge>
                                            )}
                                            {v.active && (
                                                <Badge className="bg-primary/20 text-primary hover:bg-primary/30">Active</Badge>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell className="text-right">
                                        {!v.installed ? (
                                            <Button variant="outline" size="sm">Install</Button>
                                        ) : v.active ? (
                                            <Button variant="ghost" size="sm" disabled>Current</Button>
                                        ) : (
                                            <Button variant="secondary" size="sm">Switch</Button>
                                        )}
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
