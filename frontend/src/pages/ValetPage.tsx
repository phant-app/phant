import { ActionButton } from '@/components/ui/action-button';
import { JsonBox } from '@/components/ui/json-box';
import { PageCard } from '@/components/ui/page-card';
import { ValueRow } from '@/components/ui/value-row';
import type { ValetLinuxRemediationResult, ValetLinuxVerification } from '@/types';

export function ValetPage({
    valetVerification,
    refreshingValet,
    onRefresh,
    confirmValetRemediation,
    onConfirm,
    applyingValetRemediation,
    onApply,
    valetRemediationResult,
    embedded = false,
}: {
    valetVerification: ValetLinuxVerification | null;
    refreshingValet: boolean;
    onRefresh: () => void;
    confirmValetRemediation: boolean;
    onConfirm: (checked: boolean) => void;
    applyingValetRemediation: boolean;
    onApply: () => void;
    valetRemediationResult: ValetLinuxRemediationResult | null;
    embedded?: boolean;
}) {
    return (
        <div className="space-y-6">
            {!embedded ? (
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Valet Linux</h1>
                    <p className="text-muted-foreground mt-2">
                        Diagnose and manage your Valet configuration safely. Review FPM hook targets and apply remediation.
                    </p>
                </div>
            ) : null}

            <PageCard
                title="Verification"
                actions={
                    <>
                        <ActionButton onClick={onRefresh} disabled={refreshingValet}>
                            {refreshingValet ? 'Checking...' : 'Refresh'}
                        </ActionButton>
                        <ActionButton
                            onClick={onApply}
                            disabled={applyingValetRemediation || !confirmValetRemediation}
                        >
                            {applyingValetRemediation ? 'Applying...' : 'Apply Remediation'}
                        </ActionButton>
                    </>
                }
            >
                <label className="flex items-center gap-2 text-sm text-slate-300">
                    <input
                        type="checkbox"
                        checked={confirmValetRemediation}
                        onChange={(event) => onConfirm(event.target.checked)}
                    />
                    Confirm I want to modify FPM hook INI files and attempt service restarts.
                </label>

                <div className="grid gap-2 md:grid-cols-2">
                    <ValueRow label="Generated" value={valetVerification?.generatedAt || 'n/a'} />
                    <ValueRow label="Supported OS" value={valetVerification?.supported ? 'yes' : 'no'} />
                    <ValueRow label="Valet detected" value={valetVerification?.valetDetected ? 'yes' : 'no'} />
                    <ValueRow label="Service manager" value={valetVerification?.serviceManager || 'n/a'} />
                    <ValueRow label="CLI conf.d" value={valetVerification?.cliConfDPath || 'n/a'} copyable />
                    <ValueRow label="CLI prepend" value={valetVerification?.cliAutoPrepend || 'n/a'} copyable />
                    <ValueRow label="Expected prepend" value={valetVerification?.expectedPrependPath || 'n/a'} copyable />
                    {valetVerification?.lastError ? (
                        <ValueRow label="Verification error" value={valetVerification.lastError} />
                    ) : null}
                </div>
            </PageCard>

            {valetVerification?.fpmServices?.length ? (
                <PageCard title="PHP-FPM services">
                    <div className="space-y-3">
                        {valetVerification.fpmServices.map((service) => (
                            <JsonBox key={service.serviceName} value={service} />
                        ))}
                    </div>
                </PageCard>
            ) : null}

            {valetVerification?.recommendations?.length ? (
                <PageCard title="Recommendations">
                    <ul className="list-inside list-disc space-y-1 text-sm text-slate-300">
                        {valetVerification.recommendations.map((item, index) => (
                            <li key={`${index}-${item}`}>{item}</li>
                        ))}
                    </ul>
                </PageCard>
            ) : null}

            {valetRemediationResult ? (
                <PageCard title="Last remediation run">
                    <div className="grid gap-2 md:grid-cols-2">
                        <ValueRow label="Status" value={valetRemediationResult.applied ? 'applied' : 'not applied'} />
                        <ValueRow
                            label="Message"
                            value={valetRemediationResult.message || valetRemediationResult.error || 'n/a'}
                        />
                    </div>
                    {valetRemediationResult.targets?.length ? (
                        <div className="space-y-3">
                            {valetRemediationResult.targets.map((target) => (
                                <JsonBox key={target.serviceName} value={target} />
                            ))}
                        </div>
                    ) : null}
                    {valetRemediationResult.suggestedCommands?.length ? (
                        <div className="space-y-2">
                            {valetRemediationResult.suggestedCommands.map((command, index) => (
                                <JsonBox key={`${index}-${command}`} value={command} />
                            ))}
                        </div>
                    ) : null}
                </PageCard>
            ) : null}
        </div>
    );
}
