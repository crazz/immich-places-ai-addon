import {draftRequest} from './draftApi';
import {isDraft} from './draftTypes';
import {isTranslationPage, isTranslationRun} from './translationTypes';

import type {TDraft} from './draftTypes';
import type {TTranslationPage, TTranslationRequest, TTranslationRun} from './translationTypes';

export async function submitTranslation(input: TTranslationRequest, signal?: AbortSignal): Promise<TTranslationRun> {return draftRequest('/ai/translations', 'POST', input, isTranslationRun, signal);}
export async function fetchTranslation(id: string, signal?: AbortSignal): Promise<TTranslationRun> {return draftRequest(`/ai/translations/${encodeURIComponent(id)}`, 'GET', undefined, isTranslationRun, signal);}
export async function fetchTranslations(draftId: string, before: string, signal?: AbortSignal): Promise<TTranslationPage> {return draftRequest(`/ai/translations?draftId=${encodeURIComponent(draftId)}&before=${encodeURIComponent(before)}`, 'GET', undefined, isTranslationPage, signal);}
export async function adoptTranslation(id: string, revision: number, languages: string[], signal?: AbortSignal): Promise<TDraft> {return draftRequest(`/ai/translations/${encodeURIComponent(id)}/adopt`, 'POST', {revision,languages}, isDraft, signal);}
export async function cancelTranslation(id: string, signal?: AbortSignal): Promise<TTranslationRun> {return draftRequest(`/ai/translations/${encodeURIComponent(id)}/cancel`, 'POST', {}, isTranslationRun, signal);}
