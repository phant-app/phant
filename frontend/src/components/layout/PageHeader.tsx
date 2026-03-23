import { type ReactNode } from 'react';
import { cn } from '@/lib/utils';

export function PageHeader({
    title,
    watermark,
    description,
    meta,
    actions,
    className,
}: {
    title: string;
    watermark?: string;
    description?: ReactNode;
    meta?: ReactNode;
    actions?: ReactNode;
    className?: string;
}) {
    return (
        <div className={cn('relative isolate border-b-2 border-border pb-4', className)}>
            {watermark ? (
                <div
                    aria-hidden="true"
                    role="presentation"
                    className="pointer-events-none absolute -bottom-5 right-0 z-0 select-none font-rock text-[86px] text-zinc-200/80 dark:text-zinc-900/40 md:text-[150px]"
                >
                    {watermark}
                </div>
            ) : null}

            <div className="relative z-10 flex items-end justify-between gap-4">
                <h1 className="font-rock text-4xl tracking-wide text-foreground uppercase md:text-5xl">{title}</h1>
                {actions ? <div className="relative z-10 flex items-center gap-2">{actions}</div> : null}
            </div>

            {description ? (
                <p className="relative z-10 mt-2 font-mono text-[11px] tracking-[0.14em] text-muted-foreground uppercase">{description}</p>
            ) : null}

            {meta ? (
                <p className="relative z-10 mt-1 font-mono text-[10px] tracking-[0.12em] text-muted-foreground uppercase">{meta}</p>
            ) : null}
        </div>
    );
}
