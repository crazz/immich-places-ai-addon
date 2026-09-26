import {StandardWriteComparison} from './StandardWriteComparison';

import type {TMirrorWritePlan, TStackWritePlan} from './writePreviewTypes';
import type {ReactElement} from 'react';

export function StackWriteComparison({plan}: {plan: TMirrorWritePlan | TStackWritePlan}): ReactElement {
 return <section aria-label={'Approved stack target matrix'} className={'min-w-0 space-y-3'}>
  <p>{`Exact approved targets: ${plan.manifest.targets.length} · Stack ${plan.manifest.stackId}`}</p>
  {plan.manifest.targets.map(target => <section aria-label={`Approved photo ${target.assetId}`} key={target.assetId} className={'min-w-0 rounded border p-2'}>
   <p className={'break-all'}>{`Photo ${target.assetId}`}</p>
   <p className={'break-all'}>{`Reviewed source: ${target.imageIdentity}`}</p>
   <StandardWriteComparison plan={target} />
                                       </section>)}
        </section>;
}
