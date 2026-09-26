import {isRecord} from '@/utils/typeGuards';

import {isUUID} from './selectionTypes';

export type TTranslationRequest = {key: string; draftId: string; revision: number; factsRevision: number; profileId: string; profileRevision: number; basis: string; basisKind: 'scene' | 'candidate'; languages: string[]; confirmed: boolean; parentId?: string; maxTokens?: number; maxEstimatedMicros?: number};
export type TTranslationItem = {language: string; state: 'queued' | 'reserved' | 'complete' | 'unavailable' | 'failed' | 'canceled' | 'interrupted'; text: string | null; failure: string};
export type TTranslationRun = {id: string; request: TTranslationRequest; items: TTranslationItem[]; policyId: string};
export type TTranslationPage = {ids: string[]; next: string};
const positive = (value: unknown): value is number => Number.isSafeInteger(value) && Number(value) > 0;
const bounded = (value: unknown, bytes: number): value is string => typeof value === 'string' && new TextEncoder().encode(value).length <= bytes;
function isRequest(value: unknown): value is TTranslationRequest {
 return isRecord(value) && isUUID(value.draftId) && bounded(value.key, 128) && value.key.length > 0 && bounded(value.profileId, 128) && value.profileId.length > 0 &&
 [value.revision, value.factsRevision, value.profileRevision].every(positive) && bounded(value.basis, 16384) && value.basis.trim().length > 0 && ['scene', 'candidate'].includes(String(value.basisKind)) &&
 value.confirmed === true && Array.isArray(value.languages) && value.languages.length > 0 && value.languages.length <= 8 && value.languages.every(tag => typeof tag === 'string' && /^[a-zA-Z]{2,8}(?:-[a-zA-Z0-9]{1,8})*$/.test(tag) && tag.length <= 64) && new Set(value.languages).size === value.languages.length &&
 (value.parentId === undefined || bounded(value.parentId, 128)) && (value.maxTokens === undefined || (Number.isSafeInteger(value.maxTokens) && Number(value.maxTokens) >= 0 && Number(value.maxTokens) <= 8_000_000_000)) && (value.maxEstimatedMicros === undefined || (Number.isSafeInteger(value.maxEstimatedMicros) && Number(value.maxEstimatedMicros) >= 0));
}
function isItem(value: unknown): value is TTranslationItem {
 return isRecord(value) && bounded(value.language, 64) && bounded(value.failure, 128) && ['queued', 'reserved', 'complete', 'unavailable', 'failed', 'canceled', 'interrupted'].includes(String(value.state)) &&
 (value.state === 'complete' ? bounded(value.text, 16384) && value.text.trim().length > 0 : value.text === null);
}
export function isTranslationRun(value: unknown): value is TTranslationRun {
 if (!isRecord(value) || !isUUID(value.id) || !isRequest(value.request) || typeof value.policyId !== 'string' || !/^[0-9a-f]{64}$/.test(value.policyId) || !Array.isArray(value.items) || !value.items.every(isItem)) {return false;}
 const languages = value.request.languages;
 return value.items.length === languages.length && new Set(value.items.map(item => item.language)).size === languages.length && value.items.every(item => languages.includes(item.language));
}
export function isTranslationPage(value: unknown): value is TTranslationPage {return isRecord(value) && Array.isArray(value.ids) && value.ids.length <= 20 && value.ids.every(isUUID) && new Set(value.ids).size === value.ids.length && (value.next === '' || isUUID(value.next));}
