import {savedDraft} from './draft';

import type {TTranslationRun} from '../translationTypes';

export function translationRun(): TTranslationRun {
 const draft = savedDraft();
 return {id: '88888888-8888-4888-8888-888888888888', policyId: 'a'.repeat(64), request: {key: 'key', draftId: draft.id, revision: 1, factsRevision: 1, profileId: 'profile', profileRevision: 1, basis: 'An uncertain bridge.', basisKind: 'scene', languages: ['en', 'uk'], confirmed: true}, items: [{language: 'en', state: 'complete', text: 'An uncertain bridge.', failure: ''}, {language: 'uk', state: 'failed', text: null, failure: 'provider_failed'}]};
}
