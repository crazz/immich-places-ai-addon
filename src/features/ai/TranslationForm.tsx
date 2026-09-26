import {useState} from 'react';

import type {TDraft} from './draftTypes';
import type {TProviderList} from './providerApi';
import type {TTranslationRequest} from './translationTypes';
import type {ReactElement} from 'react';

export function TranslationForm({draft, providers, disabled, onSubmit, retry, onEdit}: {draft: TDraft; providers: TProviderList; disabled: boolean; onSubmit: (value: TTranslationRequest) => void; onEdit: () => void; retry?: {id: string; basis: string; basisKind: 'scene' | 'candidate'; languages: string[]}}): ReactElement {
 const current = draft.descriptions.find(item => !item.stale && item.factsRevision === draft.factsRevision && item.status === 'complete');
 const [basis, setBasis] = useState(retry?.basis ?? current?.text ?? '');
 const [kind, setKind] = useState<'scene' | 'candidate'>(retry?.basisKind ?? (current?.basis === 'candidate' ? 'candidate' : 'scene'));
 const [tags, setTags] = useState(retry?.languages.join(', ') ?? 'en, uk');
 const [profileId, setProfileId] = useState('');const [isConfirmed, setConfirmed] = useState(false);
 const profile = providers.items.find(item => item.id === profileId && item.enabled);
 const languages = tags.split(',').map(tag => tag.trim());
 const isValid = languages.length > 0 && languages.length <= 8 && languages.every(tag => /^[a-zA-Z]{2,8}(?:-[a-zA-Z0-9]{1,8})*$/.test(tag)) && new Set(languages.map(tag => tag.toLowerCase())).size === languages.length && basis.trim().length > 0 && new TextEncoder().encode(basis).length <= 16384;
 return <fieldset disabled={disabled || !providers.enabled} className={'space-y-2'} onChange={onEdit}>
  <legend>{retry ? 'Review unsuccessful languages for a new run' : 'Review text to translate'}</legend>
  <label className={'block'}>{'Reviewed text basis'}<textarea value={basis} onChange={event => {setBasis(event.target.value);setConfirmed(false);}} className={'block w-full rounded border p-2'} /></label>
  <label className={'block'}>{'Basis kind'}<select value={kind} onChange={event => {setKind(event.target.value === 'candidate' ? 'candidate' : 'scene');setConfirmed(false);}}><option value={'scene'}>{'Scene only'}</option><option value={'candidate'}>{'Location dependent'}</option></select></label>
  <label className={'block'}>{'Translation languages'}<input value={tags} readOnly={Boolean(retry)} onChange={event => {setTags(event.target.value);setConfirmed(false);}} /></label>
  <label className={'block'}>{'Translation provider'}<select value={profileId} onChange={event => {setProfileId(event.target.value);setConfirmed(false);}}><option value={''}>{'Choose provider and model'}</option>{providers.items.filter(item => item.enabled).map(item => <option key={item.id} value={item.id}>{`${item.name} · ${item.model}`}</option>)}</select></label>
  <p>{'Only the reviewed text and translation instructions are sent. Up to eight languages, one bounded provider call each. Review each suggestion for factual and linguistic accuracy before adoption.'}</p>
  <label className={'block'}><input type={'checkbox'} checked={isConfirmed} onChange={event => setConfirmed(event.target.checked)} />{'I approve sending this text to the selected provider and model.'}</label>
  <button type={'button'} disabled={!isValid || !profile || !isConfirmed} onClick={() => {if (profile) {onSubmit({key: crypto.randomUUID(), draftId: draft.id, revision: draft.revision, factsRevision: draft.factsRevision, profileId: profile.id, profileRevision: profile.revision, basis, basisKind: kind, languages, confirmed: true, ...(retry ? {parentId: retry.id} : {})});}}}>{'Generate translations'}</button>
        </fieldset>;
}
