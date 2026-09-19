'use client';

import {ProviderSettings} from '@/features/ai';
import {optionButtonClass} from '@/features/filterBar/constant';
import {HiddenFilterGroup} from '@/features/filterBar/HiddenFilterGroup';
import {NumericOptionGroup} from '@/features/filterBar/NumericOptionGroup';
import {cn} from '@/utils/cn';
import {GRID_COLUMN_OPTIONS, PAGE_SIZE_ALL, PAGE_SIZE_OPTIONS, formatMarkerLimitOption} from '@/utils/view';

import type {THiddenFilter} from '@/shared/types/map';
import type {ReactElement} from 'react';

function formatPageSizeLabel(option: number): string {
	if (option === PAGE_SIZE_ALL) {
		return 'All';
	}
	return String(option);
}

type TSettingsPanelProps = {
	pageSize: number;
	onPageSizeAction: (size: number) => void;
	gridColumns: number;
	onGridColumnsAction: (cols: number) => void;
	hiddenFilter: THiddenFilter;
	onHiddenFilterAction: (filter: THiddenFilter) => void;
	activeVisibleMarkerLimit: number;
	hasVisibleMarkerLimitOptions: boolean;
	canDecreaseVisibleMarkerLimit: boolean;
	canIncreaseVisibleMarkerLimit: boolean;
	decreaseStepLabel: string;
	increaseStepLabel: string;
	markerMaxText: string;
	onDecreaseVisibleMarkerLimitAction: () => void;
	onIncreaseVisibleMarkerLimitAction: () => void;
};

export function SettingsPanel({
	pageSize,
	onPageSizeAction,
	gridColumns,
	onGridColumnsAction,
	hiddenFilter,
	onHiddenFilterAction,
	activeVisibleMarkerLimit,
	hasVisibleMarkerLimitOptions,
	canDecreaseVisibleMarkerLimit,
	canIncreaseVisibleMarkerLimit,
	decreaseStepLabel,
	increaseStepLabel,
	markerMaxText,
	onDecreaseVisibleMarkerLimitAction,
	onIncreaseVisibleMarkerLimitAction
}: TSettingsPanelProps): ReactElement {
	return (
		<div className={'flex flex-col gap-1.5 px-3 pb-2.5'}>
			<ProviderSettings />
			<div className={'flex gap-1.5'}>
				<NumericOptionGroup
					label={'Per Page'}
					value={pageSize}
					options={PAGE_SIZE_OPTIONS}
					onChangeAction={onPageSizeAction}
					formatLabel={formatPageSizeLabel}
				/>
				<NumericOptionGroup
					label={'Grid'}
					value={gridColumns}
					options={GRID_COLUMN_OPTIONS}
					onChangeAction={onGridColumnsAction}
				/>
			</div>
			<div className={'flex gap-1.5'}>
				{hasVisibleMarkerLimitOptions && (
					<div className={'min-w-0 basis-1/2 rounded-lg bg-(--color-bg) p-2.5'}>
						<div className={'mb-1 flex items-center justify-between'}>
							<div
								className={
									'text-[0.5625rem] font-semibold uppercase tracking-[0.08em] text-(--color-text-secondary)'
								}>
								{'Markers'}
							</div>
							<div
								className={
									'ml-2 shrink-0 whitespace-nowrap text-[0.5625rem] text-(--color-text-secondary)'
								}>
								{markerMaxText}
							</div>
						</div>
						<div className={'flex items-center gap-1'}>
							<button
								onClick={onDecreaseVisibleMarkerLimitAction}
								disabled={!canDecreaseVisibleMarkerLimit}
								className={cn(
									optionButtonClass,
									'px-1.5',
									'border-(--color-border) bg-transparent text-(--color-text-secondary) hover:border-(--color-text-secondary)',
									'disabled:cursor-default disabled:opacity-40'
								)}>
								{`-${decreaseStepLabel}`}
							</button>
							<div
								className={
									'min-w-[3.5rem] text-center text-[0.6875rem] font-semibold text-(--color-text)'
								}>
								{formatMarkerLimitOption(activeVisibleMarkerLimit)}
							</div>
							<button
								onClick={onIncreaseVisibleMarkerLimitAction}
								disabled={!canIncreaseVisibleMarkerLimit}
								className={cn(
									optionButtonClass,
									'px-1.5',
									'border-(--color-border) bg-transparent text-(--color-text-secondary) hover:border-(--color-text-secondary)',
									'disabled:cursor-default disabled:opacity-40'
								)}>
								{`+${increaseStepLabel}`}
							</button>
						</div>
					</div>
				)}
				<div
					className={cn(
						'min-w-0',
						hasVisibleMarkerLimitOptions && 'basis-1/2',
						!hasVisibleMarkerLimitOptions && 'flex-1'
					)}>
					<HiddenFilterGroup
						hiddenFilter={hiddenFilter}
						onHiddenFilterAction={onHiddenFilterAction}
					/>
				</div>
			</div>
		</div>
	);
}
