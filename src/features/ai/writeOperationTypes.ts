import {isRecord} from '@/utils/typeGuards';

import {isDescriptionObservation} from './descriptionWriteTypes';
import {isMirrorOutcome} from './mirrorTypes';
import {isUUID} from './selectionTypes';
import {isPreviewGPS, isWritePreview} from './writePreviewTypes';

import type {TDescriptionObservation} from './descriptionWriteTypes';
import type {TMirrorOutcome} from './mirrorTypes';
import type {TPreviewGPS, TWritePlan} from './writePreviewTypes';

const states = ['queued', 'writing', 'verifying', 'retryable', 'succeeded', 'partial', 'conflict', 'failed', 'canceled', 'expired'];
export type TFieldOutcome = {field: 'gps' | 'description'; status: 'pending' | 'verified' | 'baseline' | 'conflict' | 'unavailable'; wasVerified: boolean; gps?: TPreviewGPS; description?: TDescriptionObservation};
export type TTargetOutcome = {assetId: string; status: string; code: string; attempts: number; generation: number; observed: TPreviewGPS | null; verified: boolean; refreshed: boolean; noop: boolean; settled: boolean; fields: TFieldOutcome[]; events: {code: string; at: string; attempt: number}[]};
export type TWriteOperation = {mirror?: TMirrorOutcome; targets?: TTargetOutcome[]; fields?: TFieldOutcome[]; id: string; plan: TWritePlan; digest: string; status: string; code: string; approvedAt: string; attempts: number; generation: number; observed: TPreviewGPS | null; verified: boolean; refreshed: boolean; noop: boolean; settled: boolean; events: {code: string; at: string; attempt: number}[]};
export type TWriteConfirmation = {previewId: string; digest: string; idempotencyKey: string};
export type TWriteHistory = {items: {id: string; status: string; draftRevision: number; approvedAt: string}[]; nextCursor: string};

const date = (value: unknown): boolean => typeof value === 'string' && Number.isFinite(Date.parse(value));
const count = (value: unknown): boolean => Number.isSafeInteger(value) && Number(value) >= 0;
const code = (value: unknown): boolean => typeof value === 'string' && /^[a-zA-Z0-9_]{0,80}$/.test(value);

export function isWriteOperation(value: unknown): value is TWriteOperation {
 return isRecord(value) && isUUID(value.id) && isWritePreview({plan: value.plan, digest: value.digest, status: 'usable', diff: 'changed'}) && validOutcomes(value) && validTargets(value) && (isRecord(value.plan) && value.plan.version === 'mirror-preview-v4' ? isRecord(value.plan.mirror) && isMirrorOutcome(value.mirror, value.plan.mirror.value) && value.mirror.assetId === value.plan.targetId : value.mirror === undefined) && states.includes(String(value.status)) && code(value.code) && date(value.approvedAt) && count(value.attempts) && Number(value.attempts) <= 2 && count(value.generation) && (value.observed === null || isPreviewGPS(value.observed)) && ['verified', 'refreshed', 'noop', 'settled'].every(key => typeof value[key] === 'boolean') && Array.isArray(value.events) && value.events.length <= 100 && value.events.every(item => isRecord(item) && code(item.code) && date(item.at) && count(item.attempt) && Number(item.attempt) <= 2);
}

export function isWriteHistory(value: unknown): value is TWriteHistory {
 return isRecord(value) && (value.nextCursor === '' || isUUID(value.nextCursor)) && Array.isArray(value.items) && value.items.length <= 100 && value.items.every(item => isRecord(item) && isUUID(item.id) && states.includes(String(item.status)) && count(item.draftRevision) && Number(item.draftRevision) > 0 && date(item.approvedAt));
}

export function isWriteConfirmation(value: unknown): value is TWriteConfirmation {
 return isRecord(value) && isUUID(value.previewId) && typeof value.digest === 'string' && /^[a-f0-9]{64}$/.test(value.digest) && typeof value.idempotencyKey === 'string' && /^[a-f0-9]{32}$/.test(value.idempotencyKey);
}

function validOutcomes(value: Record<string, unknown>): boolean {
 if (!isRecord(value.plan) || value.plan.version !== 'standard-preview-v2') {return value.fields === undefined;}
 return validFieldOutcomes(value.plan.fields, value.fields);
}

function validFieldOutcomes(fields: unknown, outcomes: unknown): boolean {
 return Array.isArray(fields) && Array.isArray(outcomes) && fields.length === outcomes.length && outcomes.every((item, index) => isRecord(item) && item.field === fields[index] && ['pending', 'verified', 'baseline', 'conflict', 'unavailable'].includes(String(item.status)) && typeof item.wasVerified === 'boolean' && (item.status !== 'verified' || item.wasVerified) && (!['verified', 'baseline'].includes(String(item.status)) || (item.field === 'gps' ? isPreviewGPS(item.gps) : isDescriptionObservation(item.description))) && (item.gps === undefined || (item.field === 'gps' && isPreviewGPS(item.gps))) && (item.description === undefined || (item.field === 'description' && isDescriptionObservation(item.description))));
}

function validTargets(value: Record<string, unknown>): boolean {
 if (!isRecord(value.plan) || !['stack-preview-v3', 'mirror-preview-v4'].includes(String(value.plan.version))) {return value.targets === undefined && value.status !== 'expired';}
 if (!isRecord(value.plan.manifest) || !Array.isArray(value.plan.manifest.targets) || !Array.isArray(value.targets) || value.targets.length !== value.plan.manifest.targets.length || value.attempts !== 0 || value.generation !== 0 || value.observed !== null || !Array.isArray(value.events) || value.events.length !== 0) {return false;}
 const targets = value.plan.manifest.targets;
 return value.targets.every((target, index) => isRecord(target) && isRecord(targets[index]) && target.assetId === targets[index].assetId && states.includes(String(target.status)) && code(target.code) && count(target.attempts) && Number(target.attempts) <= 2 && count(target.generation) && (target.observed === null || isPreviewGPS(target.observed)) && ['verified', 'refreshed', 'noop', 'settled'].every(key => typeof target[key] === 'boolean') && validFieldOutcomes(targets[index].fields, target.fields) && Array.isArray(target.events) && target.events.length <= 100 && target.events.every(item => isRecord(item) && code(item.code) && date(item.at) && count(item.attempt) && Number(item.attempt) <= 2));
}
