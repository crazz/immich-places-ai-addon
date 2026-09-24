import {resultDetail} from './resultDetail';

import type {TDraft} from '../draftTypes';

export function savedDraft(): TDraft {
 const detail = resultDetail();
 return {
id: '99999999-9999-4999-8999-999999999999',
analysisId: detail.entry.analysisId!,
assetId: detail.entry.assetId,
 revision: 1,
state: 'draft',
camera: null,
fields: [],
candidateId: null,
radius: null,
radiusBasis: '',
radiusStale: false,
 factsRevision: 1,
heading: null,
headingStale: false,
headingUserSupplied: false,
headingFactsRevision: 1,
descriptions: [],
 originalSourceDigest: 'a'.repeat(64),
baseline: {status: 'unavailable', latitude: null, longitude: null, imageIdentity: '', sourceDigest: '', assetId: '', ownerId: '', checksum: '', type: '', observedAt: ''},
updatedAt: '2026-09-24T12:00:00Z'
};
}
