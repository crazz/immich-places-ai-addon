import {expect, it} from 'vitest';

import {isJobProgress} from './jobValidation';
import {aiProgress} from './testing/fixtures';

it('rejects inconsistent progress counts so an active job cannot appear complete', () => {
 const job = aiProgress();
 expect(isJobProgress(job)).toBe(true);
 expect(isJobProgress({...job, counts: {total: 1, succeeded: 0}})).toBe(false);
 expect(isJobProgress({...job, counts: {total: 1, succeeded: 1}})).toBe(false);
 expect(isJobProgress({...job, items: [job.items[0], job.items[0]], counts: {total: 2, queued: 2}})).toBe(false);
});
