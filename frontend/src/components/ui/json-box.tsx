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
            <pre className={`overflow-x-auto rounded-md border border-border bg-muted/50 p-3 text-xs text-muted-foreground font-mono ${expanded ? 'max-h-none' : 'max-h-[300px] overflow-y-auto'}`}>
                {displayString}
            </pre>
            {isLarge && (
                <Button 
                    type="button"
                    variant="outline" 
                    size="sm" 
                    className="absolute top-2 right-4 opacity-0 group-hover:opacity-100 transition-opacity bg-background/80 backdrop-blur-sm"
                    onClick={() => setExpanded(!expanded)}
                >
                    {expanded ? 'Collapse' : 'Expand full JSON'}
                </Button>
            )}
        </div>
    );
}
