import { ActionButton } from '@/components/ui/action-button';
import { JsonBox } from '@/components/ui/json-box';
import { PageCard } from '@/components/ui/page-card';
import { ValueRow } from '@/components/ui/value-row';
import type { HookInstallResult, SetupDiagnostics } from '@/types';

export function SetupPage({
    diagnostics,
    hookResult,
    installingHook,
    onRefresh,
    onEnable,
    embedded = false,
}: {
    diagnostics: SetupDiagnostics | null;
    hookResult: HookInstallResult | null;
    installingHook: boolean;
    onRefresh: () => void;
    onEnable: () => void;
    embedded?: boolean;
}) {
    return (
        <div className="space-y-6">
            {!embedded ? (
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">CLI Diagnostics</h1>
                    <p className="text-muted-foreground mt-2">
                        Verify that your PHP CLI setup includes the correct prepended hooks to capture dumps.
                    </p>
                </div>
            ) : null}

            <PageCard
                title="CLI hook"
                actions={
                    <>
                        <ActionButton onClick={onRefresh}>Refresh</ActionButton>
                        <ActionButton onClick={onEnable} disabled={installingHook}>
                            {installingHook ? 'Enabling...' : 'Enable CLI Hook'}
                        </ActionButton>
                    </>
                }
            >
                <div className="grid gap-2 md:grid-cols-2">
                    <ValueRow label="Generated" value={diagnostics?.generatedAt || 'n/a'} />
                    <ValueRow label="PHP found" value={diagnostics?.phpFound ? 'yes' : 'no'} />
                    <ValueRow label="PHP version" value={diagnostics?.phpVersion || 'n/a'} />
                    <ValueRow label="Service manager" value={diagnostics?.serviceManager || 'unknown'} />
                    {diagnostics?.lastError ? <ValueRow label="Diagnostics error" value={diagnostics.lastError} /> : null}
                </div>
            </PageCard>

            {hookResult ? (
                <PageCard title="Last install attempt">
                    <div className="grid gap-2 md:grid-cols-2">
                        <ValueRow label="Status" value={hookResult.success ? 'enabled' : 'failed'} />
                        <ValueRow label="Message" value={hookResult.message || hookResult.error || 'n/a'} />
                        <ValueRow label="php.ini" value={hookResult.phpIniPath || 'n/a'} copyable />
                        <ValueRow label="prepend file" value={hookResult.prependPath || 'n/a'} copyable />
                        <ValueRow label="backup" value={hookResult.backupPath || 'n/a'} copyable />
                    </div>
                    {hookResult.requiresSudo && hookResult.suggestedCmd ? <JsonBox value={hookResult.suggestedCmd} /> : null}
                </PageCard>
            ) : null}
        </div>
    );
}
