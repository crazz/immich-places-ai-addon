import type {TJobConfiguration, TJobProgress} from '@/features/ai/jobTypes';
import type {TLaunchFields} from '@/features/ai/launchAdmission';
import type {TProviderProfile} from '@/features/ai/providerApi';
import type {TSelectionPreview} from '@/features/ai/selectionTypes';

export const aiProfile: TProviderProfile = {
 id: 'p',
name: 'Private',
baseURL: 'https://provider.example/v1',
model: 'bound-model',
enabled: true,
revision: 1,
hasSecret: true,
 executionReadiness: {status: 'ready', source: 'operator-attested', policyID: 'a'.repeat(64), maxInputTokens: 100, maxOutputTokens: 40, maxRequestBytes: 10000, maxImageBytes: 5000, costStatus: 'unknown', contextAllowed: true},
 capabilityReport: {attemptID: 'attempt', profileID: 'p', revision: 1, protocolVersion: 'capability-v1', policyFingerprint: 'fp', lifecycle: 'completed', startedAt: '2026-09-20T12:00:00Z', deadlineAt: '2026-09-20T12:02:00Z', completedAt: '2026-09-20T12:00:05Z', requestedModel: 'bound-model', observations: {image: {status: 'supported'}, strict: {status: 'supported'}, json: {status: 'supported'}}, compatibility: 'strict-schema sample compatible', applicable: true}
};
export function aiPreview(): TSelectionPreview {
 return {mode: 'explicit', scope: {view: 'all', gpsFilter: 'no-gps', hiddenFilter: 'visible'}, snapshotID: '11111111-1111-4111-8111-111111111111', policyVersion: 'selection-v1', assetIDs: ['aaaaaaaa-0000-4000-8000-000000000001'], requestedCount: 1, uniqueCount: 1, duplicateCount: 0, eligibleCount: 1, excludedCount: 0, exclusions: [], createdAt: new Date().toISOString(), expiresAt: new Date(Date.now() + 60000).toISOString()};
}
export const aiFields: TLaunchFields = {profileId: 'p', mode: 'visual', format: 'strict', languages: 'en', primaryLanguage: 'en', maxCalls: '1', maxTokens: '140', outputTokens: '40', costCap: '', classes: [], hint: ''};
export const aiConfiguration: TJobConfiguration = {selectionToken: '11111111-1111-4111-8111-111111111111', profileId: 'p', revision: 1, mode: 'visual', format: 'strict', allowJson: false, languages: ['en'], primaryLanguage: 'en', policyId: 'a'.repeat(64), limits: {maxCalls: 1, maxTokens: 140, outputTokens: 40}};
export function aiProgress(): TJobProgress {
 return {id: '22222222-2222-4222-8222-222222222222', configuration: aiConfiguration, createdAt: 1789919280000000000, canceled: false, blocked: false, counts: {queued: 1, total: 1}, items: [{id: '33333333-3333-4333-8333-333333333333', assetId: 'aaaaaaaa-0000-4000-8000-000000000001', state: 'queued', attempts: 0, calls: 0, resultId: null}], usage: {calls: 0, reservedTokens: 0, reportedStatus: 'unknown', costStatus: 'unknown', inputReported: null, outputReported: null, totalReported: null, estimatedMicros: null}};
}
