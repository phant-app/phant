import { type ReactNode } from 'react';
import { Button } from '@/components/ui/button';

export function ActionButton({
    onClick,
    disabled,
    children,
}: {
    onClick: () => void;
    disabled?: boolean;
    children: ReactNode;
}) {
    return (
        <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={onClick}
            disabled={disabled}
        >
            {children}
        </Button>
    );
}
