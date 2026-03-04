import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

const mockServices = [
    { name: 'MySQL', status: 'running', port: 3306 },
    { name: 'Redis', status: 'running', port: 6379 },
    { name: 'Mailpit', status: 'stopped', port: null },
];

export function ServicesPage() {
    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Services</h1>
                <p className="text-muted-foreground mt-2">
                    Manage background services on your machine commonly used for Laravel development.
                </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {mockServices.map((service) => (
                    <Card key={service.name}>
                        <CardHeader className="pb-3">
                            <div className="flex justify-between items-start">
                                <CardTitle>{service.name}</CardTitle>
                                {service.status === 'running' ? (
                                    <Badge className="bg-green-600/20 text-green-500 hover:bg-green-600/30">Running</Badge>
                                ) : (
                                    <Badge variant="secondary">Stopped</Badge>
                                )}
                            </div>
                            <CardDescription>
                                {service.port ? `Port ${service.port}` : 'Not bound'}
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="flex gap-2 w-full mt-2">
                                {service.status === 'running' ? (
                                    <>
                                        <Button variant="outline" className="w-full">Restart</Button>
                                        <Button variant="destructive" className="w-full">Stop</Button>
                                    </>
                                ) : (
                                    <Button className="w-full">Start</Button>
                                )}
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>
        </div>
    );
}
