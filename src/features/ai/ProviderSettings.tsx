'use client';

import {Description} from '@radix-ui/react-dialog';
import {useState} from 'react';

import {DialogShell} from '@/shared/components/DialogShell';

import {ProviderForm} from './ProviderForm';
import {useProviderList} from './useProviderList';

import type {TProviderProfile} from './providerApi';
import type {ReactElement} from 'react';

function ProviderDialog({onCloseAction}: {onCloseAction: () => void}): ReactElement {
	const {data, error, isLoading, reload} = useProviderList();
	const [selected, setSelected] = useState<TProviderProfile | 'new' | null>(null);
	const [isSaving, setIsSaving] = useState(false);
	const [savedRevision, setSavedRevision] = useState<number | null>(null);
	return (
		<DialogShell
			isOpen={true}
			onClose={() => {
				if (!isSaving) {
					onCloseAction();
				}
			}}
			title={'AI providers'}
			subtitle={'Private settings. Saving sends no images or provider requests.'}>
			<Description className={'sr-only'}>
				{'Manage private provider profiles. Destinations and capabilities have not been verified.'}
			</Description>
			<div className={'max-h-[75vh] overflow-y-auto p-4 text-(--color-text)'}>
				{isLoading && <p role={'status'}>{'Loading provider settings…'}</p>}
				{error && <p role={'alert'}>{error}</p>}
				{data && !data.enabled && <p>{'AI is disabled by the administrator.'}</p>}
				{savedRevision !== null && <p role={'status'} className={'mb-3 text-sm'}>{`Saved revision ${savedRevision}`}</p>}
				{selected && data?.enabled && (
					<ProviderForm
						key={selected === 'new' ? 'new' : `${selected.id}:${selected.revision}`}
						profile={selected === 'new' ? undefined : selected}
						onBusyAction={setIsSaving}
						onCancelAction={() => setSelected(null)}
						onSavedAction={profile => {
							setSavedRevision(profile.revision);
							setSelected(null);
							reload();
						}}
					/>
				)}
				{!selected && data?.enabled && (
					<>
						{data.items.length === 0 && <p className={'mb-3 text-sm'}>{'No private providers yet.'}</p>}
						<ul className={'mb-3 flex flex-col gap-3'}>
							{data.items.map(profile => (
								<li key={profile.id} className={'rounded border border-(--color-border) p-3'}>
									<p className={'font-medium'}>{profile.name}</p>
									<p className={'break-all text-xs'}>{profile.baseURL}</p>
									<p className={'text-xs'}>{`${profile.model} · ${profile.enabled ? 'Enabled' : 'Disabled'} · Revision ${profile.revision}`}</p>
									<button className={'mt-2 text-sm underline'} onClick={() => setSelected(profile)}>
										{`Edit ${profile.name}`}
									</button>
								</li>
							))}
						</ul>
						<button
							className={'rounded border border-(--color-border) px-3 py-2 text-sm'}
							onClick={() => {
								setSelected('new');
								setSavedRevision(null);
							}}>
							{'Create provider'}
						</button>
					</>
				)}
				{!isLoading && (
					<button
						className={'mt-3 block text-xs underline'}
						disabled={isSaving}
						onClick={() => {
							setSelected(null);
							reload();
						}}>
						{'Reload profiles'}
					</button>
				)}
			</div>
		</DialogShell>
	);
}

export function ProviderSettings(): ReactElement {
	const [isOpen, setIsOpen] = useState(false);
	return (
		<>
			<button
				className={'rounded border border-(--color-border) px-3 py-2 text-xs text-(--color-text)'}
				onClick={() => setIsOpen(true)}>
				{'AI providers'}
			</button>
			{isOpen && <ProviderDialog onCloseAction={() => setIsOpen(false)} />}
		</>
	);
}
