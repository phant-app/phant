import { useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';

export function JsonBox({ value }: { value: unknown }) {
    const [expanded, setExpanded] = useState(false);
    
    // Vercel rule rerender-memo: Extract expensive serialization work
    const jsonString = useMemo(() => {
        try {
            return JSON.stringify(value, null, 2) || '';
        } catch (e) {
            return String(value);
        }
    }, [value]);

    const isLarge = jsonString.length > 2000;
    const displayString = !expanded && isLarge 
        ? jsonString.slice(0, 2000) + '\n\n... [Payload truncated for scrolling performance. Click Expand to view full JSON]' 
        : jsonString;

    return (
        <div className="relative group">
            <pre className={`overflow-x-auto border border-zinc-300 bg-zinc-50 p-3 font-mono text-xs leading-relaxed text-zinc-700 cut-corner dark:border-zinc-800 dark:bg-black dark:text-zinc-300 ${expanded ? 'max-h-none' : 'max-h-[300px] overflow-y-auto'}`}>
                {displayString}
            </pre>
            {isLarge && (
                <Button 
                    type="button"
                    variant="outline" 
                    size="sm" 
                    className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity"
                    onClick={() => setExpanded(!expanded)}
                >
                    {expanded ? 'Collapse' : 'Expand full JSON'}
                </Button>
            )}
        </div>
    );
}
