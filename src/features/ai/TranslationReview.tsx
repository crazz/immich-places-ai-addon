import {TranslationForm} from './TranslationForm';
import {TranslationSuggestions} from './TranslationSuggestions';
import {useTranslationReview} from './useTranslationReview';

import type {TDraft} from './draftTypes';
import type {ReactElement} from 'react';

export type TTranslationReviewProps = {owner: string; draft: TDraft; disabled?: boolean; onAdopt: (value: TDraft) => void; onDirtyChange?: (value: boolean) => void};
export function TranslationReview(props: TTranslationReviewProps): ReactElement {return <TranslationPanel key={`${props.owner}:${props.draft.id}:${props.draft.revision}`} {...props} />;}
function TranslationPanel({draft, disabled, onAdopt, onDirtyChange}: TTranslationReviewProps): ReactElement {
 const {hasEdits,shouldConfirmClose,markEdited,closeReview,keepEditing,discardEdits,isOpen,setOpen,providers,error,run,isBusy,pending,retry,setRetry,history,submit,adopt,update,browse,selectRun} = useTranslationReview({draft,disabled,onAdopt,onDirtyChange});
 const isStale = Boolean(run && (run.request.revision !== draft.revision || run.request.factsRevision !== draft.factsRevision));
 return <section aria-label={'Translation review'} className={'space-y-3 rounded border p-3'}><button type={'button'} onClick={() => isOpen ? closeReview() : setOpen(true)}>{isOpen ? 'Close translation review' : 'Translate reviewed text'}</button>{isOpen && <>{error && <p role={'alert'}>{error}</p>}<TranslationForm onEdit={markEdited} key={retry ? `${retry.id}:${retry.languages.join(',')}` : 'fresh'} retry={retry} draft={draft} providers={providers} disabled={Boolean(disabled || isBusy || pending || run?.items.some(item => ['queued','reserved'].includes(item.state)))} onSubmit={input => void submit(input)} />
 {shouldConfirmClose && <div role={'alert'}><p>{'Unsaved translation text will be discarded.'}</p><button type={'button'} onClick={keepEditing}>{'Keep translation edits'}</button><button type={'button'} onClick={discardEdits}>{'Discard translation edits'}</button></div>}
 {pending && !isBusy && <button type={'button'} disabled={disabled} onClick={() => void submit(pending)}>{'Recover submitted translation'}</button>}
 <div><label>{'Translation history'}<select value={run?.id ?? ''} disabled={isBusy} onChange={event => void selectRun(event.target.value)}><option value={''} disabled>{'Choose saved run'}</option>{history.ids.map((id,index) => <option key={id} value={id}>{`Run ${index+1} · newest first`}</option>)}</select></label><button type={'button'} disabled={isBusy} onClick={() => void browse('')}>{'Newest translation runs'}</button>{history.next && <button type={'button'} disabled={isBusy} onClick={() => void browse(history.next)}>{'Older translation runs'}</button>}</div>
 {run && <div><button type={'button'} disabled={isBusy} onClick={() => void update(false)}>{'Refresh translations'}</button>{run.items.some(item => ['queued','reserved'].includes(item.state)) && <button type={'button'} disabled={isBusy} onClick={() => void update(true)}>{'Cancel unfinished translations'}</button>}</div>}
 {isStale && <p role={'status'}>{'These suggestions are stale because the saved draft or factual basis changed. Review the current text and generate again before adoption.'}</p>}
 {run && <TranslationSuggestions key={run.id} run={run} disabled={Boolean(disabled || isBusy || hasEdits || isStale)} onAdopt={languages => void adopt(languages)} onRetry={languages => setRetry({id: run.id, basis: run.request.basis, basisKind: run.request.basisKind, languages})} />}
                                                                                                                                                                                                                                                                  </>}
        </section>;
}
