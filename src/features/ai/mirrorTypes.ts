import {isRecord} from '@/utils/typeGuards';

import {isUUID} from './selectionTypes';

export type TMirrorSelection = {direction: boolean; precision: boolean; place: boolean; languages: string[] | null; provenance: boolean; model: boolean};
export const emptyMirrorSelection: TMirrorSelection = {direction: false, precision: false, place: false, languages: [], provenance: false, model: false};

export function isMirrorLanguage(value: unknown): value is string {
 if (typeof value !== 'string' || !value.length || value.length > 100) {return false;}
 try {return Intl.getCanonicalLocales(value)[0] === value;} catch {return false;}
}

export function isMirrorSelection(value: unknown): value is TMirrorSelection {
 if (!isRecord(value) || !Object.keys(value).every(key => ['direction', 'precision', 'place', 'languages', 'provenance', 'model'].includes(key)) || !['direction', 'precision', 'place', 'provenance', 'model'].every(key => typeof value[key] === 'boolean')) {return false;}
 const languages = value.languages ?? [];
 return Array.isArray(languages) && languages.length <= 8 && languages.every(isMirrorLanguage) && new Set(languages).size === languages.length && (value.direction || value.precision || value.place || value.provenance || languages.length > 0) && (!value.model || value.provenance === true);
}

export type TMirrorObservation = {present: boolean; value?: Record<string, unknown>};
export type TMirrorExport = {schemaVersion: 1; application: 'immich-places-ai-addon'; recordId: string; review: {draftRevision: number; factsRevision: number}; direction?: {heading: number | null; method: string; uncertainty: number | null}; precision?: {radius: number | null; basis: string}; place?: string; descriptions?: Record<string, string>; provenance?: {mode: 'visual' | 'context-assisted' | 'research'; model?: string; reviewedAt: string; fields: Record<string, 'user' | 'model'>}};
export type TMirrorPlan = {key: 'immich-places-ai-addon'; recordId: string; before: TMirrorObservation; value: TMirrorExport; selection: TMirrorSelection; disclosure: 'asset-readers-v1'};

const hasKeys = (value: Record<string, unknown>, keys: string[]): boolean => Object.keys(value).every(key => keys.includes(key));
const bytes = (value: unknown): number => new TextEncoder().encode(JSON.stringify(value)).length;
const text = (value: unknown, limit: number): value is string => typeof value === 'string' && value.trim().length > 0 && new TextEncoder().encode(value).length <= limit;
const count = (value: unknown): boolean => Number.isSafeInteger(value) && Number(value) > 0;
const finite = (value: unknown, max: number): value is number => typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= max;

export function isMirrorObservation(value: unknown): value is TMirrorObservation {
 return isRecord(value) && hasKeys(value, ['present', 'value']) && typeof value.present === 'boolean' && (value.present ? isRecord(value.value) && bytes(value.value) <= 65536 : value.value === undefined);
}

export function isMirrorPlan(value: unknown): value is TMirrorPlan {
 if (!isRecord(value) || !hasKeys(value, ['key', 'recordId', 'before', 'value', 'selection', 'disclosure']) || value.key !== 'immich-places-ai-addon' || !isUUID(value.recordId) || !isMirrorObservation(value.before) || value.disclosure !== 'asset-readers-v1' || !isMirrorSelection(value.selection) || !isRecord(value.value)) {return false;}
 const exported = value.value; const choice = value.selection; const languages = choice.languages ?? [];
 if (!hasKeys(exported, ['schemaVersion', 'application', 'recordId', 'review', 'direction', 'precision', 'place', 'descriptions', 'provenance']) || bytes(exported) > 65536 || exported.schemaVersion !== 1 || exported.application !== value.key || exported.recordId !== value.recordId || !isRecord(exported.review) || !hasKeys(exported.review, ['draftRevision', 'factsRevision']) || !count(exported.review.draftRevision) || !count(exported.review.factsRevision) || Number(exported.review.factsRevision) > Number(exported.review.draftRevision)) {return false;}
 if (['direction', 'precision', 'place', 'provenance'].some(key => choice[key as 'direction' | 'precision' | 'place' | 'provenance'] !== (exported[key] !== undefined))) {return false;}
 if (languages.length ? !isRecord(exported.descriptions) || Object.keys(exported.descriptions).length !== languages.length || !languages.every(language => isRecord(exported.descriptions) && text(exported.descriptions[language], 16384)) : exported.descriptions !== undefined) {return false;}
 if (choice.place && !text(exported.place, 8192)) {return false;}
 return validDirection(exported.direction) && validPrecision(exported.precision) && validProvenance(exported.provenance, choice);
}

function validDirection(value: unknown): boolean {
 if (value === undefined) {return true;}
 if (!isRecord(value) || !hasKeys(value, ['heading', 'method', 'uncertainty']) || !['unknown', 'user_supplied', 'visual_estimate', 'known_viewpoint_alignment'].includes(String(value.method))) {return false;}
 if (value.heading === null) {return value.method === 'unknown' && value.uncertainty === null;}
 return finite(value.heading, 360) && value.heading < 360 && (value.uncertainty === null || (!['unknown', 'user_supplied'].includes(String(value.method)) && finite(value.uncertainty, 180)));
}
function validPrecision(value: unknown): boolean {
 return value === undefined || (isRecord(value) && hasKeys(value, ['radius', 'basis']) && ['unknown', 'visual_estimate', 'context_extent', 'source_reported', 'model_estimate'].includes(String(value.basis)) && (value.radius === null ? value.basis === 'unknown' : finite(value.radius, Number.MAX_VALUE) && value.basis !== 'unknown'));
}
function validProvenance(value: unknown, choice: TMirrorSelection): boolean {
 if (value === undefined) {return true;}
 const fields = [...(choice.languages ?? []).map(language => `description:${language}`), ...(['direction', 'precision', 'place'] as const).filter(key => choice[key])];
 return isRecord(value) && hasKeys(value, ['mode', 'model', 'reviewedAt', 'fields']) && ['visual', 'context-assisted', 'research'].includes(String(value.mode)) && (choice.model ? text(value.model, 256) : value.model === undefined) && typeof value.reviewedAt === 'string' && Number.isFinite(Date.parse(value.reviewedAt)) && isRecord(value.fields) && Object.keys(value.fields).length === fields.length && fields.every(field => isRecord(value.fields) && ['user', 'model'].includes(String(value.fields[field])));
}

export type TMirrorOutcome = {assetId: string; step: 'metadata'; status: string; code: string; attempts: number; generation: number; observed: TMirrorObservation | null; verified: boolean; noop: boolean; settled: boolean; events: {code: string; at: string; attempt: number; generation: number}[]};
export function isMirrorOutcome(value: unknown, intended: unknown): value is TMirrorOutcome {
 if (!isRecord(value) || !hasKeys(value, ['assetId', 'step', 'status', 'code', 'attempts', 'generation', 'observed', 'verified', 'noop', 'settled', 'events'])) {return false;}
 if (!isUUID(value.assetId) || value.step !== 'metadata' || !['blocked', 'queued', 'writing', 'verifying', 'retryable', 'succeeded', 'conflict', 'failed', 'canceled', 'expired'].includes(String(value.status))) {return false;}
 if (!mirrorCode(value.code) || !mirrorCounter(value.attempts, 2) || !mirrorCounter(value.generation) || (value.observed !== null && !isMirrorObservation(value.observed))) {return false;}
 if (!['verified', 'noop', 'settled'].every(key => typeof value[key] === 'boolean') || !Array.isArray(value.events) || value.events.length > 100 || !value.events.every(isMirrorEvent)) {return false;}
 if (value.status !== 'succeeded' && value.code !== 'METADATA_OBSERVED_UNRESOLVED') {return true;}
 return value.verified === true && isMirrorObservation(value.observed) && value.observed.present && mirrorJSONEqual(value.observed.value, intended);
}

function mirrorCode(value: unknown): boolean {return typeof value === 'string' && /^[a-zA-Z0-9_]{0,80}$/.test(value);}
function mirrorCounter(value: unknown, maximum = Number.MAX_SAFE_INTEGER): boolean {return Number.isSafeInteger(value) && Number(value) >= 0 && Number(value) <= maximum;}
function isMirrorEvent(event: unknown): boolean {
 return isRecord(event) && hasKeys(event, ['code', 'at', 'attempt', 'generation']) && mirrorCode(event.code) && typeof event.at === 'string' && Number.isFinite(Date.parse(event.at)) && mirrorCounter(event.attempt, 2) && mirrorCounter(event.generation);
}

export function mirrorJSONEqual(left: unknown, right: unknown): boolean {
 if (left === right) {return true;}
 if (Array.isArray(left) || Array.isArray(right)) {return Array.isArray(left) && Array.isArray(right) && left.length === right.length && left.every((value, index) => mirrorJSONEqual(value, right[index]));}
 return isRecord(left) && isRecord(right) && Object.keys(left).length === Object.keys(right).length && Object.keys(left).every(key => Object.hasOwn(right, key) && mirrorJSONEqual(left[key], right[key]));
}
