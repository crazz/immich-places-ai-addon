import {useRef, useState} from 'react';

import {saveProvider} from './providerApi';

import type {TProviderProfile} from './providerApi';
import type {FormEvent, ReactElement} from 'react';

type TProviderFormProps = {
	profile?: TProviderProfile;
	onSavedAction: (profile: TProviderProfile) => void;
	onCancelAction: () => void;
	onBusyAction: (isBusy: boolean) => void;
};

const inputClass = 'w-full rounded border border-(--color-border) bg-(--color-bg) px-2 py-1.5 text-sm text-(--color-text)';

export function ProviderForm({profile, onSavedAction, onCancelAction, onBusyAction}: TProviderFormProps): ReactElement {
	const [name, setName] = useState(profile?.name ?? '');
	const [baseURL, setBaseURL] = useState(profile?.baseURL ?? '');
	const [model, setModel] = useState(profile?.model ?? '');
	const [isEnabled, setIsEnabled] = useState(profile?.enabled ?? true);
	const [secret, setSecret] = useState('');
	const [shouldRemoveSecret, setShouldRemoveSecret] = useState(false);
	const [isSaving, setIsSaving] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const isSubmitting = useRef(false);

	async function submit(event: FormEvent<HTMLFormElement>): Promise<void> {
		event.preventDefault();
		if (isSubmitting.current) {
			return;
		}
		isSubmitting.current = true;
		setIsSaving(true);
		onBusyAction(true);
		setError(null);
		const input = {name, baseURL, model, enabled: isEnabled, secret: shouldRemoveSecret ? '' : secret || undefined};
		setSecret('');
		try {
			onSavedAction(await saveProvider(input, profile));
		} catch (failure) {
			setError(failure instanceof Error ? failure.message : 'Could not save provider settings.');
		} finally {
			isSubmitting.current = false;
			setIsSaving(false);
			onBusyAction(false);
		}
	}

	return (
		<form onSubmit={submit} aria-label={'Provider profile'} className={'flex flex-col gap-3'}>
			<fieldset disabled={isSaving} className={'flex flex-col gap-3'}>
				<label className={'text-sm'}>
					{'Name'}
					<input className={inputClass} value={name} onChange={event => setName(event.target.value)} required maxLength={80} />
				</label>
				<label className={'text-sm'}>
					{'API base URL'}
					<input className={inputClass} type={'url'} value={baseURL} onChange={event => setBaseURL(event.target.value)} required maxLength={2048} />
				</label>
				<label className={'text-sm'}>
					{'Model'}
					<input className={inputClass} value={model} onChange={event => setModel(event.target.value)} required maxLength={200} />
				</label>
				<label className={'text-sm'}>
					{'API key'}
					<input
						className={inputClass}
						type={'password'}
						autoComplete={'off'}
						value={secret}
						onChange={event => setSecret(event.target.value)}
						disabled={shouldRemoveSecret}
						maxLength={4096}
					/>
				</label>
				{profile?.hasSecret && (
					<>
						<p className={'text-xs text-(--color-text-secondary)'}>{'A key is stored. Leave blank to keep it.'}</p>
						<label className={'text-sm'}>
							<input
								type={'checkbox'}
								checked={shouldRemoveSecret}
								onChange={event => {
									setShouldRemoveSecret(event.target.checked);
									setSecret('');
								}}
							/>
							{' Remove all stored keys for this profile'}
						</label>
					</>
				)}
				{profile?.hasSecret && baseURL !== profile.baseURL && (
					<p className={'text-xs'}>{'Changing the URL requires a replacement key or explicit removal.'}</p>
				)}
				<label className={'text-sm'}>
					<input type={'checkbox'} checked={isEnabled} onChange={event => setIsEnabled(event.target.checked)} />
					{' Enabled'}
				</label>
				<p className={'text-xs text-(--color-text-secondary)'}>
					{'Disabling keeps encrypted keys. Removing keys clears all profile revisions, but does not rewrite backups.'}
				</p>
			</fieldset>
			{error && <p role={'alert'} className={'text-sm text-red-600'}>{error}</p>}
			<div className={'flex gap-3'}>
				<button type={'submit'} disabled={isSaving} className={'rounded bg-(--color-primary) px-3 py-2 text-sm text-white disabled:opacity-50'}>
					{isSaving ? 'Saving…' : 'Save provider'}
				</button>
				<button type={'button'} disabled={isSaving} onClick={onCancelAction} className={'rounded border border-(--color-border) px-3 py-2 text-sm'}>
					{'Cancel'}
				</button>
			</div>
		</form>
	);
}
