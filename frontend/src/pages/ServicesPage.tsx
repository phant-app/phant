import { useEffect, useMemo, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ActionButton } from "@/components/ui/action-button";
import { PageHeader } from "@/components/layout/PageHeader";
import { GetServicesStatus } from "../../bindings/phant/internal/services/servicesstatusservice";
import type { ServicesStatusSnapshot } from "@/types";

const sortOrder = {
    running: 0,
    stopped: 1,
    unavailable: 2,
} as const;

const badgeByState = {
    running: "border-emerald-300/70 bg-emerald-400/30 text-emerald-200",
    stopped: "border-amber-500/50 bg-amber-500/10 text-amber-500",
    unavailable: "border-zinc-500/50 bg-zinc-500/10 text-zinc-400",
} as const;

const stateLabel = {
    running: "Running",
    stopped: "Stopped",
    unavailable: "Unavailable",
} as const;

export function ServicesPage() {
    const [snapshot, setSnapshot] = useState<ServicesStatusSnapshot | null>(null);
    const [loading, setLoading] = useState(false);

    const loadServices = async () => {
        setLoading(true);
        try {
            const result = await GetServicesStatus();
            setSnapshot(result as unknown as ServicesStatusSnapshot);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        void loadServices();
    }, []);

    const orderedServices = useMemo(() => {
        const services = snapshot?.services ?? [];
        return [...services].sort((left, right) => sortOrder[left.state] - sortOrder[right.state]);
    }, [snapshot?.services]);

    return (
        <div className="space-y-6">
            <PageHeader
                title="Services"
                watermark="SRV"
                description="Check local service status for Laravel development workflows."
                meta={snapshot?.generatedAt ? `Last checked: ${new Date(snapshot.generatedAt).toLocaleString()}` : "Status not loaded yet"}
                actions={(
                    <ActionButton onClick={() => { void loadServices(); }} disabled={loading}>
                        {loading ? "Refreshing..." : "Refresh"}
                    </ActionButton>
                )}
            />

            {snapshot?.lastError ? (
                <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
                    {snapshot.lastError}
                </div>
            ) : null}

            {snapshot?.warnings?.length ? (
                <div className="rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-600 dark:text-amber-300">
                    <ul className="list-disc list-inside space-y-1">
                        {snapshot.warnings.map((warning) => (
                            <li key={warning}>{warning}</li>
                        ))}
                    </ul>
                </div>
            ) : null}

            {loading && orderedServices.length === 0 ? (
                <p className="text-sm text-muted-foreground">Loading services status...</p>
            ) : null}

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {orderedServices.map((service) => (
                    <Card key={service.id}>
                        <CardHeader className="pb-3">
                            <div className="flex justify-between items-start">
                                <CardTitle>{service.label}</CardTitle>
                                <Badge variant="outline" className={badgeByState[service.state]}>{stateLabel[service.state]}</Badge>
                            </div>
                            <CardDescription>
                                {service.description}
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-2 pt-0">
                            <p className="font-mono text-xs text-muted-foreground">Port {service.port}</p>
                            <p className="font-mono text-xs text-muted-foreground">Unit: {service.unit || "N/A"}</p>
                        </CardContent>
                    </Card>
                ))}

                {!loading && orderedServices.length === 0 ? (
                    <Card>
                        <CardHeader>
                            <CardTitle>No services detected</CardTitle>
                            <CardDescription>Try refreshing to re-run machine discovery.</CardDescription>
                        </CardHeader>
                    </Card>
                ) : null}
            </div>
        </div>
    );
}
