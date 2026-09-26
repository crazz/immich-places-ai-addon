import {isCurrentDescription} from './descriptionWriteTypes';
import {emptyMirrorSelection} from './mirrorTypes';

import type {TDraft} from './draftTypes';
import type {TMirrorSelection} from './mirrorTypes';
import type {ReactElement} from 'react';

export function MirrorSelection({draft, selection, onChange, placeName}: {draft: TDraft; placeName?: string; selection: TMirrorSelection | null; onChange: (value: TMirrorSelection | null) => void}): ReactElement {
 return <section aria-label={'Optional metadata selection'} className={'space-y-2'}>
  <label><input type={'checkbox'} checked={!!selection} onChange={event => onChange(event.target.checked ? {...emptyMirrorSelection, languages: []} : null)} />{'Mirror optional metadata for this photo'}</label>
  <p>{`Metadata target: ${draft.assetId}`}</p>
  <p>{'Optional custom metadata does not promise native Immich display or file/EXIF updates. Local records remain available. Support is verified separately by the server.'}</p>
  {selection && <>
   <label className={'block'}><input type={'checkbox'} checked={selection.place} disabled={!draft.candidateId} onChange={event => onChange({...selection, place: event.target.checked})} />{'Mirror chosen place'}</label>
   {draft.candidateId && <p>{`Chosen place for mirror: ${placeName || draft.candidateId}`}</p>}
   <label className={'block'}><input type={'checkbox'} checked={selection.direction} disabled={!selection.direction && draft.headingStale} onChange={event => onChange({...selection, direction: event.target.checked})} />{'Mirror direction'}</label>
   <label className={'block'}><input type={'checkbox'} checked={selection.precision} disabled={!selection.precision && draft.radiusStale} onChange={event => onChange({...selection, precision: event.target.checked})} />{'Mirror precision'}</label>
   <label className={'block'}><input type={'checkbox'} checked={selection.provenance} onChange={event => onChange({...selection, provenance: event.target.checked, model: event.target.checked && selection.model})} />{'Mirror minimal provenance'}</label>
   <label className={'block'}><input type={'checkbox'} checked={selection.model} disabled={!selection.provenance} onChange={event => onChange({...selection, model: event.target.checked})} />{'Include model name'}</label>
                </>}
  {selection && draft.descriptions.map(item => <label key={item.language} className={'block'}><input type={'checkbox'} checked={selection.languages?.includes(item.language) ?? false} disabled={!selection.languages?.includes(item.language) && (!isCurrentDescription(item, draft.factsRevision) || (selection.languages?.length ?? 0) >= 8)} onChange={event => onChange({...selection, languages: event.target.checked ? [...(selection.languages ?? []), item.language].sort() : (selection.languages ?? []).filter(language => language !== item.language)})} />{`Mirror description ${item.language}`}</label>)}
        </section>;
}
