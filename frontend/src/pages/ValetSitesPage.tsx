import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";

const mockSites = [
    { name: 'example-app', path: '/home/ronald/code/example-app', url: 'http://example-app.test', php: '8.3', ssl: false },
    { name: 'my-store', path: '/home/ronald/code/my-store', url: 'https://my-store.test', php: '8.2', ssl: true },
];

export function ValetSitesPage() {
    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Valet Sites</h1>
                <p className="text-muted-foreground mt-2">
                    Manage your linked and parked Valet sites.
                </p>
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>Local Sites</CardTitle>
                    <CardDescription>All isolated domains mapped via Laravel Valet.</CardDescription>
                </CardHeader>
                <CardContent>
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Site</TableHead>
                                <TableHead>URL</TableHead>
                                <TableHead>PHP</TableHead>
                                <TableHead>SSL</TableHead>
                                <TableHead className="text-right">Action</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {mockSites.map((site) => (
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
                                        <Badge variant="outline">PHP {site.php}</Badge>
                                    </TableCell>
                                    <TableCell>
                                        {site.ssl ? (
                                            <Badge variant="default" className="bg-green-600/20 text-green-500 hover:bg-green-600/30">Secured</Badge>
                                        ) : (
                                            <Badge variant="secondary">None</Badge>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-right">
                                        <Button variant="ghost" size="sm">Logs</Button>
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
