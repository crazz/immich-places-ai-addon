import {expect, it} from 'vitest';

import {resultLocation, resultURL} from './resultLocation';
import {resultEntry} from './testing/results';

it('round-trips only bounded history filters and exact references while preserving unrelated navigation', () => {
 const entry = resultEntry();
 const state = {open: true, query: {state: 'succeeded', startDate: '2026-09-01', cursor: 'opaque/+='}, reference: {analysisId: entry.analysisId!}};
 const url = resultURL(new URL('http://localhost/?album=kept&aiJob=old'), state);
 expect(resultLocation(url)).toEqual(state);
 expect(url.searchParams.get('album')).toBe('kept');
 expect(url.searchParams.has('aiJob')).toBe(false);
 expect(resultLocation(new URL('http://localhost/?aiResults=1&aiResult=../bad&aiHistoryCursor=' + 'x'.repeat(2049)))).toEqual({open: true, query: {}, reference: null});
 expect(resultLocation(resultURL(url, {...state, reference: {jobId: entry.jobId, itemId: entry.id}})).reference).toEqual({jobId: entry.jobId, itemId: entry.id});
 expect(resultLocation(resultURL(url, {open: false, query: {}, reference: null})).open).toBe(false);
});
