import {isRecord} from '@/utils/typeGuards';

import {isUUID} from './selectionTypes';
import {isPreviewGPS, isWritePreview} from './writePreviewTypes';

import type {TPreviewGPS, TWritePlan} from './writePreviewTypes';

const states = ['queued', 'writing', 'verifying', 'retryable', 'succeeded', 'conflict', 'failed', 'canceled'];
export type TWriteOperation = {id: string; plan: TWritePlan; digest: string; status: string; code: string; approvedAt: string; attempts: number; generation: number; observed: TPreviewGPS | null; verified: boolean; refreshed: boolean; noop: boolean; settled: boolean; events: {code: string; at: string; attempt: number}[]};
export type TWriteConfirmation = {previewId: string; digest: string; idempotencyKey: string};
export type TWriteHistory = {items: {id: string; status: string; draftRevision: number; approvedAt: string}[]; nextCursor: string};

const date = (value: unknown): boolean => typeof value === 'string' && Number.isFinite(Date.parse(value));
const count = (value: unknown): boolean => Number.isSafeInteger(value) && Number(value) >= 0;
const code = (value: unknown): boolean => typeof value === 'string' && /^[a-zA-Z0-9_]{0,80}$/.test(value);

export function isWriteOperation(value: unknown): value is TWriteOperation {
 return isRecord(value) && isUUID(value.id) && isWritePreview({plan: value.plan, digest: value.digest, status: 'usable', diff: 'changed'}) && states.includes(String(value.status)) && code(value.code) && date(value.approvedAt) && count(value.attempts) && Number(value.attempts) <= 2 && count(value.generation) && (value.observed === null || isPreviewGPS(value.observed)) && ['verified', 'refreshed', 'noop', 'settled'].every(key => typeof value[key] === 'boolean') && Array.isArray(value.events) && value.events.length <= 100 && value.events.every(item => isRecord(item) && code(item.code) && date(item.at) && count(item.attempt) && Number(item.attempt) <= 2);
}

export function isWriteHistory(value: unknown): value is TWriteHistory {
 return isRecord(value) && (value.nextCursor === '' || isUUID(value.nextCursor)) && Array.isArray(value.items) && value.items.length <= 100 && value.items.every(item => isRecord(item) && isUUID(item.id) && states.includes(String(item.status)) && count(item.draftRevision) && Number(item.draftRevision) > 0 && date(item.approvedAt));
}

export function isWriteConfirmation(value: unknown): value is TWriteConfirmation {
 return isRecord(value) && isUUID(value.previewId) && typeof value.digest === 'string' && /^[a-f0-9]{64}$/.test(value.digest) && typeof value.idempotencyKey === 'string' && /^[a-f0-9]{32}$/.test(value.idempotencyKey);
}
