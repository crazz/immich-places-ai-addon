import {expect, it} from 'vitest';

import {buildAdmission} from './launchAdmission';
import {aiFields, aiPreview, aiProfile} from './testing/fixtures';


it('binds exact normalized language choices, policy limits and separately selected context disclosure', () => {
 const profile = aiProfile;
 const preview = {...aiPreview(), eligibleCount: 2};
 const fields = {...aiFields, languages: 'uk, en-us', primaryLanguage: 'uk', maxCalls: '2', maxTokens: '280', hint: 'not selected'};
 const visual = buildAdmission(fields, preview, profile, true, 'exact-key');
 expect(visual.configuration.languages).toEqual(['uk', 'en-US']);
 expect(visual.configuration).not.toHaveProperty('context');
 expect(visual.consent.configuration).toEqual(visual.configuration);
 expect(visual.idempotencyKey).toBe('exact-key');
 const context = buildAdmission({...fields, mode: 'context-assisted', classes: ['user_hint']}, preview, profile, true, 'new-key');
 expect(context.configuration.context).toEqual({version: 'context-v1', classes: ['user_hint'], hint: 'not selected'});
 expect(() => buildAdmission(fields, preview, profile, false, 'key')).toThrow('consent');
});

it('accepts an estimated cost cap in currency units and stores exact integer millionths', () => {
 const profile = {...aiProfile, executionReadiness: {...aiProfile.executionReadiness!, costStatus: 'estimated' as const, currency: 'USD'}};
 const admission = buildAdmission({...aiFields, costCap: '1.25'}, aiPreview(), profile, true, 'key');
 expect(admission.configuration.limits.maxEstimatedMicros).toBe(1250000);
 expect(() => buildAdmission({...aiFields, costCap: '0.0000001'}, aiPreview(), profile, true, 'key')).toThrow();
});
