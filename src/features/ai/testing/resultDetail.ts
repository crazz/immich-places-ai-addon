import {resultEntry} from './results';

import type {TResultDetail} from '../resultDetailTypes';

export function resultDetail(): TResultDetail {
const entry = resultEntry();
const schemaVersionKey = 'schema_version';
const selectionKey = 'selected_candidate_id';
const candidateKey = 'candidate_id';
const reasonKey = 'unavailable_reason';
return {
	entry,
	proposal: {[schemaVersionKey]: '1.0', [selectionKey]: null, outcome: 'unknown', observations: [{id: 'obs-1', kind: 'visual', text: '<script>Untrusted scene</script>'}], warnings: ['Uncertain'], candidates: [], descriptions: [{language: 'en', status: 'complete', text: 'An uncertain scene.', basis: 'scene_only', [candidateKey]: null, [reasonKey]: null}]},
	provenance: {Mode: 'visual', Installation: entry.jobId, Profile: entry.jobId, Model: entry.model, Revision: 1, Languages: ['en'], PrimaryLanguage: 'en', SelectionDigest: 'a'.repeat(64), SourceDigest: 'b'.repeat(64), ImageDigest: 'c'.repeat(64), PromptVersion: 'visual-v1', SchemaVersion: '1.0', ValidationVersion: 'analysis-result-v1'}
};

}
