import React from 'react';
import { ActionButton } from '@/components/ui/action-button';
import { ValueRow } from '@/components/ui/value-row';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import type { CollectorStatus, DumpEvent } from '@/types';

const getCallsiteLabel = (event: DumpEvent): string | null => {
    const frame = event.trace?.[0];
    if (!frame?.file || !frame?.line) {
        return null;
    }

    if (frame.func) {
        return `${frame.file}:${frame.line} (${frame.func})`;
    }

    return `${frame.file}:${frame.line}`;
};

const isPlainObject = (value: unknown): value is Record<string, unknown> => (
    typeof value === 'object' && value !== null && !Array.isArray(value)
);

const renderScalar = (value: unknown): React.ReactNode => {
    if (value === null) {
        return <span className="text-zinc-500 dark:text-zinc-400">null</span>;
    }

    if (value === undefined) {
        return <span className="text-zinc-500 dark:text-zinc-400">undefined</span>;
    }

    if (typeof value === 'string') {
        return <span className="text-yellow-700 dark:text-yellow-300">&quot;{value}&quot;</span>;
    }

    if (typeof value === 'number') {
        return <span className="font-bold text-purple-700 dark:text-purple-300">{value}</span>;
    }

    if (typeof value === 'boolean') {
        return <span className="font-bold text-pink-600 dark:text-pink-400">{value ? 'true' : 'false'}</span>;
    }

    return <span className="text-zinc-600 dark:text-zinc-400">{String(value)}</span>;
};

const renderPhpDumpValue = (value: unknown, depth = 0): React.ReactNode => {
    const indentClass = depth > 0 ? 'ml-4' : '';

    if (Array.isArray(value)) {
        return (
            <>
                <span className="font-bold text-cyan-700 dark:text-cyan-400">array:{value.length}</span>
                <span className="text-zinc-500">[</span>
                <span className="text-zinc-500">▼</span>
                {value.map((item, index) => (
                    <div key={`arr-${depth}-${index}`} className="ml-4">
                        <span className="text-emerald-700 dark:text-emerald-400">&quot;{index}&quot;</span>
                        <span className="text-zinc-500"> =&gt; </span>
                        {isPlainObject(item) || Array.isArray(item)
                            ? renderPhpDumpValue(item, depth + 1)
                            : renderScalar(item)}
                    </div>
                ))}
                <div className={indentClass}>
                    <span className="text-zinc-500">]</span>
                </div>
            </>
        );
    }

    if (isPlainObject(value)) {
        const entries = Object.entries(value);

        return (
            <>
                <span className="font-bold text-cyan-700 dark:text-cyan-400">array:{entries.length}</span>
                <span className="text-zinc-500">[</span>
                <span className="text-zinc-500">▼</span>
                {entries.map(([key, nested]) => (
                    <div key={`${depth}-${key}`} className="ml-4">
                        <span className="text-emerald-700 dark:text-emerald-400">&quot;{key}&quot;</span>
                        <span className="text-zinc-500"> =&gt; </span>
                        {isPlainObject(nested) || Array.isArray(nested)
                            ? renderPhpDumpValue(nested, depth + 1)
                            : renderScalar(nested)}
                    </div>
                ))}
                <div className={indentClass}>
                    <span className="text-zinc-500">]</span>
                </div>
            </>
        );
    }

    return renderScalar(value);
};

const DumpPayloadView = React.memo(({ event }: { event: DumpEvent }) => (
    <div className="border border-zinc-200 bg-white p-3 text-sm leading-relaxed dark:border-zinc-800 dark:bg-black">
        <div className="overflow-x-auto font-mono text-[13px]">
            {renderPhpDumpValue(event.payload)}
        </div>
    </div>
));

const DumpRow = React.memo(({ event }: { event: DumpEvent }) => {
    const callsiteLabel = getCallsiteLabel(event);
    const occurredAt = new Date(event.timestamp).toLocaleString();

    return (
        <article className="space-y-3 border border-zinc-300 bg-white p-4 cut-corner dark:border-zinc-800 dark:bg-black/80">
            <div className="flex items-center justify-between border-b border-zinc-200 pb-2 font-mono text-[10px] tracking-[0.12em] text-zinc-500 uppercase dark:border-zinc-800 dark:text-zinc-500">
                <span>{occurredAt}</span>
                <span className="text-primary">DUMP PAYLOAD INTERCEPTED</span>
            </div>
            {callsiteLabel ? (
                <div className="border-l-4 border-primary/80 bg-primary/5 px-3 py-2 font-mono text-xs text-zinc-600 dark:bg-primary/10 dark:text-zinc-300">
                    <span className="text-zinc-500 dark:text-zinc-500">Callsite:</span> {callsiteLabel}
                </div>
            ) : null}
            <DumpPayloadView event={event} />
        </article>
    );
});

const DumpList = React.memo(({ events }: { events: DumpEvent[] }) => {
    if (events.length === 0) {
        return (
            <div className="flex h-full items-center justify-center p-8">
                <p className="font-mono text-xs tracking-[0.14em] text-zinc-500 uppercase">
                    No dumps yet. Trigger dump() or dd() in Laravel.
                </p>
            </div>
        );
    }

    return (
        <div className="space-y-4 p-4 relative z-20">
            {events.map((event) => (
                <DumpRow key={event.id} event={event} />
            ))}
        </div>
    );
});

export function DumpsPage({
    channelName,
    status,
    events,
    onClear,
}: {
    channelName: string;
    status: CollectorStatus | null;
    events: DumpEvent[];
    onClear: () => void;
}) {
    const latestEvent = events[events.length - 1];
    const latestCallsiteLabel = latestEvent ? getCallsiteLabel(latestEvent) : null;
    const [runtimeOpen, setRuntimeOpen] = React.useState(false);

    return (
        <div className="relative flex min-h-full flex-col gap-6">
            <div className="relative border-b-2 border-border pb-4">
                <div className="pointer-events-none absolute -bottom-5 right-0 select-none font-rock text-[86px] text-zinc-200/80 dark:text-zinc-900/40 md:text-[150px]">
                    DD()
                </div>
                <div className="flex items-end justify-between gap-4">
                    <h1 className="font-rock text-4xl tracking-wide text-foreground uppercase md:text-5xl">Live Dumps</h1>
                    <div className="relative z-10 flex items-center gap-2">
                        <ActionButton onClick={() => setRuntimeOpen(true)}>Runtime</ActionButton>
                        <ActionButton onClick={onClear}>Clear Events</ActionButton>
                    </div>
                </div>
                <p className="mt-2 font-mono text-[11px] tracking-[0.14em] text-muted-foreground uppercase">
                    Monitor <span className="border border-primary/40 bg-primary/10 px-1 text-primary">dump()</span> and{' '}
                    <span className="border border-primary/40 bg-primary/10 px-1 text-primary">dd()</span> calls in real-time.
                </p>
            </div>

            <section className="flex min-h-[420px] flex-1 flex-col overflow-hidden border-2 border-border bg-muted/40 cut-corner scanlines dark:border-zinc-700 dark:bg-[#050505]">
                <header className="relative z-20 flex items-center justify-between gap-3 border-b-2 border-border bg-muted px-4 py-2 dark:border-zinc-700 dark:bg-zinc-900">
                    <span className="font-mono bg-primary px-2 py-0.5 text-[10px] font-bold tracking-[0.14em] text-primary-foreground uppercase">
                        DUMPS_RCVD: {String(events.length).padStart(2, '0')}
                    </span>
                    <span className="truncate font-mono text-[10px] text-muted-foreground">
                        {latestCallsiteLabel || 'Waiting for first dump payload...'}
                    </span>
                </header>

                <div className="relative min-h-0 flex-1 overflow-y-auto">
                    <DumpList events={events} />
                </div>
            </section>

            <Dialog open={runtimeOpen} onOpenChange={setRuntimeOpen}>
                <DialogContent className="max-w-xl cut-corner">
                    <DialogHeader>
                        <DialogTitle className="font-mono text-base tracking-[0.14em] text-primary uppercase">Runtime Configuration</DialogTitle>
                        <DialogDescription>
                            Collector runtime details for this dump channel.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="grid gap-2 md:grid-cols-2">
                        <ValueRow label="Channel" value={channelName || 'loading...'} />
                        <ValueRow label="Collector" value={status?.running ? 'running' : 'stopped'} />
                        <ValueRow label="Dropped Packets" value={String(status?.dropped ?? 0)} />
                        <ValueRow label="Socket" value={status?.socketPath || 'n/a'} copyable />
                        {status?.lastError ? <ValueRow label="Last Error" value={status.lastError} /> : null}
                    </div>

                    <DialogFooter showCloseButton />
                </DialogContent>
            </Dialog>
        </div>
    );
}
