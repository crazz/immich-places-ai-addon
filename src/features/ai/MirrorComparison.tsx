import type {TMirrorPlan} from './mirrorTypes';
import type {ReactElement} from 'react';

export function MirrorComparison({mirror, targetId}: {mirror: TMirrorPlan; targetId: string}): ReactElement {
 return <section aria-label={'Optional metadata comparison'} className={'min-w-0 space-y-2'}>
  <p className={'break-all'}>{`Mirror target: ${targetId}`}</p>
  <p>{'Step order: verify this photo’s standard fields, then its optional metadata.'}</p>
  <p>{'Authorized asset readers may see this exported information. Custom metadata does not promise native Immich display or file/EXIF updates.'}</p>
  <p>{`Selected metadata: ${[...(mirror.selection.direction ? ['direction'] : []), ...(mirror.selection.precision ? ['precision'] : []), ...(mirror.selection.place ? ['place'] : []), ...(mirror.selection.languages ?? []), ...(mirror.selection.provenance ? ['provenance'] : []), ...(mirror.selection.model ? ['model name'] : [])].join(', ')}`}</p>
  {mirror.value.direction && <p>{`Mirrored direction: ${mirror.value.direction.heading ?? 'unknown'}`}</p>}
  <details><summary>{'Exact metadata before and after'}</summary><pre aria-label={'Before metadata'} tabIndex={0} className={'whitespace-pre-wrap break-words'}>{JSON.stringify(mirror.before, null, 2)}</pre><pre aria-label={'Proposed metadata'} tabIndex={0} className={'whitespace-pre-wrap break-words'}>{JSON.stringify(mirror.value, null, 2)}</pre></details>
        </section>;
}
