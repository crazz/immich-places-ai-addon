import {isRecord} from '@/utils/typeGuards';

import {isUUID} from './selectionTypes';

import type {TDraftDescription} from './draftTypes';

export type TDescriptionObservation = {presence: 'absent' | 'null' | 'value'; value: string};
export type TDescriptionPlan = {before: TDescriptionObservation; intended: string; language: string; policy: 'replace' | 'managed_append'; lineage?: {id: string; block: string; hash: string}};

export function isDescriptionObservation(value: unknown): value is TDescriptionObservation {
 return isRecord(value) && Object.keys(value).every(key => ['presence', 'value'].includes(key)) && ['absent', 'null', 'value'].includes(String(value.presence)) && typeof value.value === 'string' && new TextEncoder().encode(value.value).length <= 64 * 1024 && (value.presence === 'value' || value.value === '');
}

export function isDescriptionPlan(value: unknown): value is TDescriptionPlan {
 if (!isRecord(value) || !Object.keys(value).every(key => ['before', 'intended', 'language', 'policy', 'lineage'].includes(key)) || !isDescriptionObservation(value.before) || typeof value.language !== 'string' || !value.language.length || value.language.length > 100 || typeof value.intended !== 'string' || !value.intended.trim().length || new TextEncoder().encode(value.intended).length > 64 * 1024) {return false;}
 if (value.policy === 'replace') {return value.lineage === undefined && new TextEncoder().encode(value.intended).length <= 16 * 1024;}
 const lineage = value.lineage;
 if (value.policy !== 'managed_append' || !isRecord(lineage) || !isUUID(lineage.id) || typeof lineage.block !== 'string' || new TextEncoder().encode(lineage.block).length > 20 * 1024 || typeof lineage.hash !== 'string' || !/^[a-f0-9]{64}$/.test(lineage.hash) || !Object.keys(lineage).every(key => ['id', 'block', 'hash'].includes(key))) {return false;}
 const prefix = `[[Immich Places AI v1:${lineage.id}]]\nLanguage: ${value.language}\n`;
 const suffix = `\n[[/Immich Places AI v1:${lineage.id}]]`;
 const text = lineage.block.slice(prefix.length, -suffix.length);
 return lineage.block.startsWith(prefix) && lineage.block.endsWith(suffix) && text.trim().length > 0 && new TextEncoder().encode(text).length <= 16 * 1024 && value.intended.split(lineage.block).length === 2;
}

export function isCurrentDescription(description: TDraftDescription | undefined, factsRevision: number): description is TDraftDescription & {text: string} {
 return !!description && description.status === 'complete' && typeof description.text === 'string' && description.text.trim().length > 0 && new TextEncoder().encode(description.text).length <= 16 * 1024 && !description.stale && (description.basis === 'scene_only' || description.factsRevision === factsRevision);
}
