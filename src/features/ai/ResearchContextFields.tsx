import type {TContextClass} from './jobTypes';
import type {TLaunchFields} from './launchAdmission';
import type {TSelectionPreview} from './selectionTypes';
import type {ReactElement} from 'react';

export function ResearchContextFields({
	fields,
	preview,
	onChangeAction
}: {
	fields: TLaunchFields;
	preview: TSelectionPreview;
	onChangeAction: (change: Partial<TLaunchFields>) => void;
}): ReactElement | null {
	const context = preview.contextPreview;
	if (!context) {
		return null;
	}
	const toggle = (kind: TContextClass, checked: boolean): void =>
		onChangeAction({classes: checked ? [...fields.classes, kind] : fields.classes.filter(item => item !== kind)});
	const captures = preview.assetIDs.flatMap((id, index) =>
		context.captureTimes[id] ? [{index, value: context.captureTimes[id]}] : []
	);
	return (
		<fieldset className={'space-y-2 sm:col-span-2'}>
			<legend>{'Optional context'}</legend>
			{context.albumLabel !== null && (
				<div>
					<label>
						<input
							type={'checkbox'}
							checked={fields.classes.includes('selected_album')}
							onChange={event => toggle('selected_album', event.target.checked)}
						/>
						{'Include selected album label'}
					</label>
					<p>{context.albumLabel}</p>
				</div>
			)}
			{captures.length > 0 && (
				<div>
					<label>
						<input
							type={'checkbox'}
							checked={fields.classes.includes('capture_time')}
							onChange={event => toggle('capture_time', event.target.checked)}
						/>
						{'Include recorded capture times'}
					</label>
					<ul className={'max-h-40 overflow-auto'}>
						{captures.map(capture => (
							<li key={capture.index}>
								{`Image ${capture.index + 1}: `}
								<span>{capture.value}</span>
							</li>
						))}
					</ul>
					<p className={'text-xs text-muted-foreground'}>
						{'Recorded values may be inaccurate. Only the displayed values are included.'}
					</p>
				</div>
			)}
		</fieldset>
	);
}
