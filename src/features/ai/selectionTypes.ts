import {isRecord} from '@/utils/typeGuards';

export type TSelectionScope = {
	view: 'all' | 'album' | 'folder';
	gpsFilter: string;
	hiddenFilter: string;
	albumID?: string;
	folderPath?: string;
	tagID?: string;
	startDate?: string;
	endDate?: string;
};
export type TSelectionInput = {mode: 'explicit' | 'all-matching'; scope: TSelectionScope; assetIDs?: string[]};
export type TSelectionPreview = {
	mode: 'explicit' | 'all-matching';
	scope: TSelectionScope;
	snapshotID: string | null;
	policyVersion: string;
	assetIDs: string[];
	requestedCount: number;
	uniqueCount: number;
	duplicateCount: number;
	eligibleCount: number;
	excludedCount: number;
	matchedCount?: number;
	exclusionCounts?: Record<string, number>;
	exclusions?: {assetID: string; reason: string}[];
	createdAt: string;
	expiresAt: string;
};

export function isCount(value: unknown): value is number {
	return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0;
}
export function isUUID(value: unknown): value is string {
	return typeof value === 'string' && /^[a-f0-9]{8}-(?:[a-f0-9]{4}-){3}[a-f0-9]{12}$/i.test(value);
}
export function isSelectionPreview(value: unknown): value is TSelectionPreview {
	if (!isRecord(value) || !isRecord(value.scope)) {return false;}
	const scope = value.scope;
	if (!['all', 'album', 'folder'].includes(String(scope.view)) ||
		typeof scope.gpsFilter !== 'string' || typeof scope.hiddenFilter !== 'string' ||
		!['albumID', 'folderPath', 'tagID', 'startDate', 'endDate'].every(key => scope[key] === undefined || typeof scope[key] === 'string') ||
		!['explicit', 'all-matching'].includes(String(value.mode)) || value.policyVersion !== 'selection-v1' ||
		(value.snapshotID !== null && !isUUID(value.snapshotID)) || !Array.isArray(value.assetIDs) || value.assetIDs.length > 500 || !value.assetIDs.every(isUUID) ||
		!['requestedCount', 'uniqueCount', 'duplicateCount', 'eligibleCount', 'excludedCount'].every(key => isCount(value[key])) ||
		value.eligibleCount !== value.assetIDs.length || new Set(value.assetIDs).size !== value.assetIDs.length ||
		typeof value.createdAt !== 'string' || !Number.isFinite(Date.parse(value.createdAt)) || typeof value.expiresAt !== 'string' || !Number.isFinite(Date.parse(value.expiresAt))) {
		return false;
	}
	if (value.exclusions !== undefined && (!Array.isArray(value.exclusions) || !value.exclusions.every(item => isRecord(item) && typeof item.assetID === 'string' && typeof item.reason === 'string'))) {
		return false;
	}
	return value.mode !== 'all-matching' || (isCount(value.matchedCount) && isRecord(value.exclusionCounts) && Object.values(value.exclusionCounts).every(isCount));
}
