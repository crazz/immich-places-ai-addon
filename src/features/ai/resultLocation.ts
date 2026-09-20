import {isUUID} from './selectionTypes';

import type {TResultReference} from './resultDetailTypes';
import type {TResultQuery} from './resultTypes';

export type TResultLocation = {open: boolean; query: TResultQuery; reference: TResultReference | null};
const fields = {state: 'State', outcome: 'Outcome', assetId: 'Asset', jobId: 'Job', startDate: 'Start', endDate: 'End', albumId: 'Album', undated: 'Undated', cursor: 'Cursor'} as const;
export function resultLocation(url: URL): TResultLocation {
 const query: TResultQuery = {};
 for (const [key, suffix] of Object.entries(fields)) {
  const value = url.searchParams.get(`aiHistory${suffix}`);
  if (value && value.length <= (key === 'cursor' ? 2048 : 64)) {Object.assign(query, {[key]: value});}
 }
 const analysisId = url.searchParams.get('aiResult');
 const jobId = url.searchParams.get('aiResultJob'); const itemId = url.searchParams.get('aiResultItem');
 const reference = isUUID(analysisId) ? {analysisId} : isUUID(jobId) && isUUID(itemId) ? {jobId, itemId} : null;
 return {open: url.searchParams.get('aiResults') === '1' || !!reference, query, reference};
}
export function resultURL(url: URL, state: TResultLocation): URL {
 const next = new URL(url);
 for (const key of ['aiResults', 'aiResult', 'aiResultJob', 'aiResultItem', ...Object.values(fields).map(suffix => `aiHistory${suffix}`)]) {next.searchParams.delete(key);}
 if (state.open) {
  next.searchParams.delete('aiJob'); next.searchParams.set('aiResults', '1');
  for (const [key, suffix] of Object.entries(fields)) {
   const value = state.query[key as keyof typeof fields];
   if (value) {next.searchParams.set(`aiHistory${suffix}`, value);}
  }
  if (state.reference) {
   if ('analysisId' in state.reference) {next.searchParams.set('aiResult', state.reference.analysisId);}
   else {next.searchParams.set('aiResultJob', state.reference.jobId); next.searchParams.set('aiResultItem', state.reference.itemId);}
  }
 }
 return next;
}
