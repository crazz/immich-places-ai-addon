import {useCallback} from 'react';

import {ResultImage} from './ResultImage';
import {fetchResults} from './resultsApi';
import {useResultRead} from './useResultRead';

import type {TResultEntry, TResultQuery} from './resultTypes';
import type {ReactElement} from 'react';

export function ResultList({owner, query, onOpenAction, onPageAction}: {owner: string; query: TResultQuery; onOpenAction: (entry: TResultEntry) => void; onPageAction: (cursor?: string) => void}): ReactElement {
 const load = useCallback(async (signal: AbortSignal) => fetchResults(query, signal), [query]);
 const {data, loading: isLoading, error, refresh} = useResultRead(owner, JSON.stringify(query), load);
 return <section aria-label={'Saved results'} className={'space-y-3'}>
  <button type={'button'} onClick={() => {if (query.cursor) {onPageAction();} else {refresh();}}}>{'Refresh results'}</button>
  {isLoading && <p role={'status'}>{'Loading results…'}</p>}
  {error && <p role={'alert'}>{error}</p>}
  {data && !data.items.length && <p>{Object.values(query).some(Boolean) ? 'No results match these filters.' : 'No completed AI runs yet.'}</p>}
  {data?.items.map(entry => <article key={entry.id} className={'space-y-2 rounded border border-(--color-border) p-3'}>
   <ResultImage owner={owner} entry={entry} />
   <h3 className={'font-semibold break-words'}>{entry.label || 'Saved AI run'}</h3>
   <p>{`Execution: ${entry.executionState} · Proposal: ${entry.proposalOutcome || 'none'}`}</p>
   <p className={'text-xs text-(--color-text-secondary)'}>{'Review: unreviewed · Write: not requested'}</p>
   <p>{`${entry.mode === 'visual' ? 'Visual' : 'Context-assisted'} · ${entry.model}`}</p>
   <p>{entry.captureDay ? `Captured ${entry.captureDay}` : 'Capture date unknown'}</p>
   <p>{entry.albumId ? `Album at launch: ${entry.albumLabel || entry.albumId}` : 'No retained album selection'}</p>
   <p className={'text-xs'}>{`Completed ${new Date(entry.terminalAt).toLocaleString()}`}</p>
   <button id={`ai-result-${entry.id}`} type={'button'} aria-label={`Open result ${entry.id}`} onClick={() => onOpenAction(entry)}>{'Open result'}</button>
                            </article>)}
  {data?.nextCursor && <button type={'button'} disabled={isLoading} onClick={() => onPageAction(data.nextCursor)}>{'Older results'}</button>}
        </section>;
}
