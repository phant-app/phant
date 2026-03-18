import { toast } from 'sonner';
import { cn } from '@/lib/utils';

export function ValueRow({ label, value, copyable = false }: { label: string; value: string; copyable?: boolean }) {
    const handleCopy = () => {
        if (!copyable || !value || value === 'n/a') return;
        navigator.clipboard.writeText(value);
        toast.success('Copied to clipboard', {
            description: value,
        });
    };

    return (
        <div className="flex items-center justify-between gap-4 overflow-hidden border border-zinc-300 bg-zinc-50 px-3 py-2 text-sm cut-corner dark:border-zinc-800 dark:bg-black/80">
            <span className="shrink-0 whitespace-nowrap font-mono text-[10px] tracking-[0.12em] text-zinc-500 uppercase">{label}</span>
            <span 
                className={cn(
                    "truncate text-right font-mono text-xs font-bold text-foreground",
                    copyable && value !== 'n/a' && "cursor-pointer hover:text-primary hover:underline transition-colors"
                )} 
                title={value}
                onClick={handleCopy}
            >
                {value}
            </span>
        </div>
    );
}
