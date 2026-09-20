'use client';

import * as Dialog from '@radix-ui/react-dialog';

import type {ReactElement, ReactNode} from 'react';

type TDialogShellProps = {
	isOpen: boolean;
	onClose: () => void;
	title: string;
	subtitle: string;
	maxWidth?: 'sm' | 'md' | 'review';
	headerTrailing?: ReactNode;
	children: ReactNode;
};

export function DialogShell({
	isOpen,
	onClose,
	title,
	subtitle,
	maxWidth = 'md',
	headerTrailing,
	children
}: TDialogShellProps): ReactElement {
	let maxWidthClass = 'max-w-md';
	if (maxWidth === 'review') {
		maxWidthClass = 'max-w-5xl';
	}
	if (maxWidth === 'sm') {
		maxWidthClass = 'max-w-sm';
	}

	return (
		<Dialog.Root
			open={isOpen}
			onOpenChange={open => {
				if (!open) {
					onClose();
				}
			}}>
			<Dialog.Portal>
				<Dialog.Overlay className={'fixed inset-0 z-[2000] bg-black/50'} />
				<div
					className={
						'pointer-events-none fixed inset-0 z-[2001] flex items-center justify-center md:pl-[calc(var(--app-sidebar-width)+1.5rem)]'
					}>
					<Dialog.Content
						className={`pointer-events-auto w-full ${maxWidthClass} rounded-lg border border-(--color-border) bg-(--color-surface) shadow-lg focus:outline-none`}>
						<div className={'flex items-center justify-between border-b border-(--color-border) px-4 py-3'}>
							<div>
								<Dialog.Title className={'text-sm font-semibold text-(--color-text)'}>
									{title}
								</Dialog.Title>
								<p className={'mt-0.5 text-xs text-(--color-text-secondary)'}>{subtitle}</p>
							</div>
							<div className={'flex items-center gap-2'}>
								{headerTrailing}
								<Dialog.Close
									aria-label={'Close dialog'}
									className={
										'flex h-6 w-6 cursor-pointer items-center justify-center rounded border-0 bg-transparent text-(--color-text-secondary) hover:text-(--color-text)'
									}>
									{'\u00D7'}
								</Dialog.Close>
							</div>
						</div>
						{children}
					</Dialog.Content>
				</div>
			</Dialog.Portal>
		</Dialog.Root>
	);
}
