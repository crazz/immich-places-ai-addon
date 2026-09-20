import {ResearchContextFields} from './ResearchContextFields';

import type {TContextClass} from './jobTypes';
import type {TLaunchFields} from './launchAdmission';
import type {TProviderProfile} from './providerApi';
import type {TSelectionPreview} from './selectionTypes';
import type {ReactElement} from 'react';


const classes: {id: TContextClass; label: string}[] = [
	{id: 'capture_time', label: 'Capture time'}, {id: 'selected_album', label: 'Selected album label'},
	{id: 'user_hint', label: 'User hint'}, {id: 'nearby_locations', label: 'Nearby known locations (metadata only)'}
];

export function LaunchFields({fields, profiles, preview, disabled, onChangeAction}: {fields: TLaunchFields; profiles: TProviderProfile[]; preview: TSelectionPreview; disabled: boolean; onChangeAction: (change: Partial<TLaunchFields>) => void}): ReactElement {
	const profile = profiles.find(item => item.id === fields.profileId);
	return <fieldset disabled={disabled} className={'grid min-w-0 gap-4 sm:grid-cols-2 [&>label]:flex [&>label]:min-w-0 [&>label]:flex-col [&>label]:gap-1.5'}>
		<label className={'sm:col-span-2'}>{'Provider'}<select aria-label={'Provider'} value={fields.profileId} onChange={event => onChangeAction({profileId: event.target.value})}>
			{profiles.map(item => <option key={item.id} value={item.id}>{`${item.name} · ${item.model} · revision ${item.revision}`}</option>)}
                                                 </select>
  </label>
		<label>{'Analysis mode'}<select aria-label={'Analysis mode'} value={fields.mode} onChange={event => onChangeAction({mode: event.target.value as TLaunchFields['mode']})}>
			<option value={'research'}>{'Research'}</option>
			<option value={'visual'}>{'Visual'}</option><option value={'context-assisted'}>{'Context-assisted'}</option>
                          </select>
  </label>
		<label>{'Requested languages'}<input aria-label={'Requested languages'} value={fields.languages} onChange={event => onChangeAction({languages: event.target.value})} placeholder={'en, uk'} /></label>
		<label>{'Primary language'}<input aria-label={'Primary language'} value={fields.primaryLanguage} onChange={event => onChangeAction({primaryLanguage: event.target.value})} /></label>
		{fields.mode === 'research' && <label className={'sm:col-span-2'}>{'Hint'}<textarea aria-label={'Hint'} value={fields.hint} onChange={event => onChangeAction({hint: event.target.value})} placeholder={'Places, dates or details you remember (optional)'} /><span>{`${new TextEncoder().encode(fields.hint).length}/2,000 UTF-8 bytes`}</span></label>}
		{fields.mode === 'research' && <ResearchContextFields fields={fields} preview={preview} onChangeAction={onChangeAction} />}
		{fields.mode === 'context-assisted' && <fieldset className={'space-y-2 sm:col-span-2'}>
			<legend>{'Context disclosure — choose each class explicitly'}</legend>
			{classes.map(choice => <label key={choice.id} className={'block'}><input type={'checkbox'} checked={fields.classes.includes(choice.id)} disabled={choice.id === 'selected_album' && (preview.scope.view !== 'album' || !preview.scope.albumID)} onChange={event => onChangeAction({classes: event.target.checked ? [...fields.classes, choice.id] : fields.classes.filter(item => item !== choice.id)})} />{choice.label}</label>)}
			{fields.classes.includes('user_hint') && <label className={'block'}>{'Hint'}<textarea aria-label={'Hint'} value={fields.hint} onChange={event => onChangeAction({hint: event.target.value})} /><span>{`${new TextEncoder().encode(fields.hint).length}/2,000 UTF-8 bytes`}</span></label>}
			<p>{'Nearby context includes up to six metadata records within six hours, without neighbor images. Source origin may be unknown; no usable context may be available.'}</p>
			{fields.classes.length === 0 && <p>{'No context classes selected. Context-assisted mode remains selected.'}</p>}
                                         </fieldset>}
		<details className={'sm:col-span-2'}>
			<summary className={'cursor-pointer font-medium'}>{'Advanced settings'}</summary>
			<div className={'mt-3 grid min-w-0 gap-4 sm:grid-cols-2 [&>label]:flex [&>label]:min-w-0 [&>label]:flex-col [&>label]:gap-1.5'}>
		<label>{'Output format'}<select aria-label={'Output format'} value={fields.format} onChange={event => onChangeAction({format: event.target.value as TLaunchFields['format']})}>
			<option value={'strict'}>{'Strict schema'}</option><option value={'json'}>{'JSON (explicit fallback permission)'}</option>
                          </select>
  </label>
		<label>{'Maximum calls'}<input aria-label={'Maximum calls'} type={'number'} min={1} max={preview.eligibleCount * 3} value={fields.maxCalls} onChange={event => onChangeAction({maxCalls: event.target.value})} /></label>
		<label>{'Total token allowance'}<input aria-label={'Total token allowance'} type={'number'} min={1} max={1e12} value={fields.maxTokens} onChange={event => onChangeAction({maxTokens: event.target.value})} /></label>
		<label>{'Output tokens per call'}<input aria-label={'Output tokens per call'} type={'number'} min={1} max={profile?.executionReadiness?.maxOutputTokens} value={fields.outputTokens} onChange={event => onChangeAction({outputTokens: event.target.value})} /></label>
		<label>{`Estimated cost cap${profile?.executionReadiness?.currency ? ` (${profile.executionReadiness.currency})` : ''}`}<input aria-label={'Estimated cost cap'} type={'number'} min={0} step={'0.000001'} value={fields.costCap} disabled={profile?.executionReadiness?.costStatus !== 'estimated'} onChange={event => onChangeAction({costCap: event.target.value})} /></label>
			</div>
			<p className={'mt-3 text-xs text-muted-foreground'}>{profile?.executionReadiness?.source === 'application-defaults' ? 'Token allowances are planning estimates. The provider may ignore the requested output limit. Cost is unknown.' : 'Reservations remain consumed after uncertain delivery. Reported usage and charges may differ from estimates.'}</p>
		</details>
        </fieldset>;
}
