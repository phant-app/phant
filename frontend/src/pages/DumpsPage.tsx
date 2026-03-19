import React from 'react';
import { Copy } from 'lucide-react';
import { toast } from 'sonner';
import { ActionButton } from '@/components/ui/action-button';
import { Button } from '@/components/ui/button';
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

type CallsiteDetails = {
    filePath: string;
    line: number;
    primary: string;
    closureContext?: string;
    functionName?: string;
};

const shortenPath = (value: string, keepSegments = 3): string => {
    if (!value) {
        return value;
    }

    const normalized = value.replace(/\\/g, '/');
    const parts = normalized.split('/').filter(Boolean);
    if (parts.length <= keepSegments) {
        return normalized;
    }

    return `.../${parts.slice(parts.length - keepSegments).join('/')}`;
};

const extractClosureContext = (funcName: string): string | undefined => {
    const match = funcName.match(/\{closure:([^}]+)\}/);
    if (!match || !match[1]) {
        return undefined;
    }

    return match[1];
};

const getCallsiteDetails = (event: DumpEvent): CallsiteDetails | null => {
    const frame = event.trace?.[0];
    if (!frame?.file || !frame?.line) {
        return null;
    }

    const shortFile = shortenPath(frame.file, 4);
    const functionName = frame.func || undefined;
    const closureContext = functionName ? extractClosureContext(functionName) : undefined;

    return {
        filePath: frame.file,
        line: frame.line,
        primary: `${shortFile}:${frame.line}`,
        closureContext,
        functionName,
    };
};

const isPlainObject = (value: unknown): value is Record<string, unknown> => (
    typeof value === 'object' && value !== null && !Array.isArray(value)
);

type DumpObjectPayload = {
    __phantType: 'object';
    __className: string;
    __objectId: number;
    __properties: Record<string, unknown>;
};

type DumpObjectRefPayload = {
    __phantType: 'object-ref';
    __className: string;
    __objectId: number;
};

type DumpResourcePayload = {
    __phantType: 'resource';
    __resourceType: string;
};

const isDumpObject = (value: unknown): value is DumpObjectPayload => (
    isPlainObject(value)
    && value.__phantType === 'object'
    && typeof value.__className === 'string'
    && typeof value.__objectId === 'number'
    && isPlainObject(value.__properties)
);

const isDumpObjectRef = (value: unknown): value is DumpObjectRefPayload => (
    isPlainObject(value)
    && value.__phantType === 'object-ref'
    && typeof value.__className === 'string'
    && typeof value.__objectId === 'number'
);

const isDumpResource = (value: unknown): value is DumpResourcePayload => (
    isPlainObject(value)
    && value.__phantType === 'resource'
    && typeof value.__resourceType === 'string'
);

const isMaxDepthMarker = (value: unknown): boolean => (
    isPlainObject(value) && value.__phantType === 'max-depth'
);

const shouldExpandByDefault = (depth: number): boolean => {
    return depth === 0;
};

const DumpToggle = React.memo(({
    expanded,
    onToggle,
}: {
    expanded: boolean;
    onToggle: () => void;
}) => (
    <button
        type="button"
        onClick={onToggle}
        className="ml-1 cursor-pointer text-zinc-500 hover:text-zinc-700 dark:hover:text-zinc-300"
        aria-label={expanded ? 'Collapse section' : 'Expand section'}
    >
        {expanded ? '▼' : '▶'}
    </button>
));

const stringifyForClipboard = (value: unknown): string => {
    const output = JSON.stringify(
        value,
        (_key, nestedValue) => (typeof nestedValue === 'bigint' ? nestedValue.toString() : nestedValue),
        2,
    );

    return output ?? String(value);
};

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
        return <span className="font-bold text-purple-600 dark:text-purple-400">{value ? 'true' : 'false'}</span>;
    }

    if (typeof value === 'bigint') {
        return <span className="font-bold text-purple-700 dark:text-purple-300">{value.toString()}</span>;
    }

    return <span className="text-zinc-600 dark:text-zinc-400">{String(value)}</span>;
};

const DumpObjectNode = React.memo(({ value, depth }: { value: DumpObjectPayload; depth: number }) => {
    const indentation = depth > 0 ? 'ml-4' : '';
    const properties = Object.entries(value.__properties);
    const [expanded, setExpanded] = React.useState(shouldExpandByDefault(depth));

    return (
        <>
            <span className="font-bold text-cyan-700 dark:text-cyan-400">{value.__className}</span>
            <span className="text-zinc-500"> </span>
            <span className="text-purple-600 dark:text-purple-400">{`{#${value.__objectId}`}</span>
            <span className="text-zinc-500"> </span>
            <DumpToggle expanded={expanded} onToggle={() => setExpanded((previous) => !previous)} />
            {expanded ? (
                <>
                    {properties.map(([propertyName, nested]) => (
                        <div key={`${depth}-${propertyName}`} className="ml-4">
                            <span className="text-zinc-500">{propertyName}</span>
                            <span className="text-zinc-500">: </span>
                            <DumpValueNode value={nested} depth={depth + 1} />
                        </div>
                    ))}
                    <div className={indentation}>
                        <span className="text-zinc-500">{`}`}</span>
                    </div>
                </>
            ) : (
                <span className="text-zinc-500">{`}`}</span>
            )}
        </>
    );
});

const DumpArrayNode = React.memo(({ value, depth }: { value: unknown[]; depth: number }) => {
    const indentation = depth > 0 ? 'ml-4' : '';
    const [expanded, setExpanded] = React.useState(shouldExpandByDefault(depth));

    return (
        <>
            <span className="font-bold text-cyan-700 dark:text-cyan-400">array:{value.length}</span>
            <span className="text-zinc-500"> [</span>
            <DumpToggle expanded={expanded} onToggle={() => setExpanded((previous) => !previous)} />
            {expanded ? (
                <>
                    {value.map((item, index) => (
                        <div key={`arr-${depth}-${index}`} className="ml-4">
                            <span className="text-emerald-700 dark:text-emerald-400">&quot;{index}&quot;</span>
                            <span className="text-zinc-500"> =&gt; </span>
                            <DumpValueNode value={item} depth={depth + 1} />
                        </div>
                    ))}
                    <div className={indentation}>
                        <span className="text-zinc-500">]</span>
                    </div>
                </>
            ) : (
                <span className="text-zinc-500">]</span>
            )}
        </>
    );
});

const DumpMapNode = React.memo(({ value, depth }: { value: Record<string, unknown>; depth: number }) => {
    const indentation = depth > 0 ? 'ml-4' : '';
    const entries = Object.entries(value);
    const [expanded, setExpanded] = React.useState(shouldExpandByDefault(depth));

    return (
        <>
            <span className="font-bold text-cyan-700 dark:text-cyan-400">array:{entries.length}</span>
            <span className="text-zinc-500"> [</span>
            <DumpToggle expanded={expanded} onToggle={() => setExpanded((previous) => !previous)} />
            {expanded ? (
                <>
                    {entries.map(([key, nested]) => (
                        <div key={`${depth}-${key}`} className="ml-4">
                            <span className="text-emerald-700 dark:text-emerald-400">&quot;{key}&quot;</span>
                            <span className="text-zinc-500"> =&gt; </span>
                            <DumpValueNode value={nested} depth={depth + 1} />
                        </div>
                    ))}
                    <div className={indentation}>
                        <span className="text-zinc-500">]</span>
                    </div>
                </>
            ) : (
                <span className="text-zinc-500">]</span>
            )}
        </>
    );
});

const DumpValueNode = React.memo(({ value, depth = 0 }: { value: unknown; depth?: number }) => {
    if (isMaxDepthMarker(value)) {
        return <span className="text-zinc-500">...</span>;
    }

    if (isDumpObject(value)) {
        return <DumpObjectNode value={value} depth={depth} />;
    }

    if (isDumpObjectRef(value)) {
        return (
            <>
                <span className="font-bold text-cyan-700 dark:text-cyan-400">{value.__className}</span>
                <span className="text-zinc-500"> </span>
                <span className="text-purple-600 dark:text-purple-400">{`{#${value.__objectId}`}</span>
                <span className="text-zinc-500"> *RECURSION* </span>
                <span className="text-zinc-500">{`}`}</span>
            </>
        );
    }

    if (isDumpResource(value)) {
        return <span className="text-zinc-500">resource({value.__resourceType})</span>;
    }

    if (Array.isArray(value)) {
        return <DumpArrayNode value={value} depth={depth} />;
    }

    if (isPlainObject(value)) {
        return <DumpMapNode value={value} depth={depth} />;
    }

    return renderScalar(value);
});

const DumpPayloadView = React.memo(({ event }: { event: DumpEvent }) => (
    <div className="border border-zinc-200 bg-white p-3 text-sm leading-relaxed dark:border-zinc-800 dark:bg-black">
        <div className="overflow-x-auto font-mono text-[13px]">
            <DumpValueNode value={event.payload} />
        </div>
    </div>
));

const DumpRow = React.memo(({ event }: { event: DumpEvent }) => {
    const callsite = getCallsiteDetails(event);
    const occurredAt = new Date(event.timestamp).toLocaleString();

    const handleCopyDump = async () => {
        try {
            const payloadText = stringifyForClipboard(event.payload);
            const header = callsite
                ? `${callsite.filePath}:${callsite.line}`
                : 'unknown-callsite';
            await navigator.clipboard.writeText(`${header}\n${payloadText}`);
            toast.success('Dump copied', {
                description: 'Payload and callsite copied to clipboard.',
            });
        } catch (error) {
            toast.error('Failed to copy dump', {
                description: 'Clipboard access was denied by the system.',
            });
        }
    };

    return (
        <article className="space-y-3 border border-zinc-300 bg-white p-4 cut-corner dark:border-zinc-800 dark:bg-black/80">
            <div className="flex items-center justify-between border-b border-zinc-200 pb-2 font-mono text-[10px] tracking-[0.12em] text-zinc-500 uppercase dark:border-zinc-800 dark:text-zinc-500">
                <span>{occurredAt}</span>
                <div className="flex items-center gap-1">
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon-xs"
                        className="h-6 w-6"
                        onClick={handleCopyDump}
                        title="Copy dump payload"
                        aria-label="Copy dump payload"
                    >
                        <Copy className="size-3" />
                    </Button>
                </div>
            </div>

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
    const latestCallsite = latestEvent ? getCallsiteDetails(latestEvent) : null;
    const [runtimeOpen, setRuntimeOpen] = React.useState(false);

    return (
        <div className="relative flex min-h-full flex-col gap-6">
            <div className="relative border-b-2 border-border pb-4">
                <div className="pointer-events-none absolute -bottom-5 right-0 select-none font-rock text-[86px] text-zinc-200/80 dark:text-zinc-900/40 md:text-[150px]">
                    DD()
                </div>
                <div className="flex items-end justify-between gap-4">
                    <h1 className="font-rock text-4xl tracking-wide text-foreground uppercase md:text-5xl">Dumps</h1>
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
                    {latestCallsite ? (
                        <div className="truncate text-right font-mono text-[10px] text-muted-foreground">
                            <div title={`${latestCallsite.filePath}:${latestCallsite.line}`}>{latestCallsite.primary}</div>
                            {latestCallsite.closureContext ? (
                                <div className="text-primary/80" title={latestCallsite.closureContext}>
                                    closure: {shortenPath(latestCallsite.closureContext, 3)}
                                </div>
                            ) : null}
                        </div>
                    ) : (
                        <span className="truncate font-mono text-[10px] text-muted-foreground">
                            Waiting for first dump payload...
                        </span>
                    )}
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
