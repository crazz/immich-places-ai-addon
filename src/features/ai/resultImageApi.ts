import {backendFetch} from '@/shared/services/backendApi.fetch';
import {getBackendBaseURL} from '@/utils/backendUrls';

import {raceAbort, withCompleteOperationDeadline} from './completeOperation';
import {isUUID} from './selectionTypes';

export async function fetchResultThumbnail(job: string, item: string, signal?: AbortSignal): Promise<Blob> {
 if (!isUUID(job) || !isUUID(item)) {throw new Error('Image unavailable');}
 return withCompleteOperationDeadline(15_000, signal, async operationSignal => raceAbort(operationSignal, (async () => {
  const response = await backendFetch(`${getBackendBaseURL()}/ai/jobs/${job}/items/${item}/thumbnail`, {method: 'GET', cache: 'no-store'}, {signal: operationSignal});
  if (!response.ok || response.headers.get('content-type') !== 'image/jpeg') {throw new Error('Image unavailable');}
  const blob = await response.blob();
  if (!blob.size || blob.size > 10 * 1024 * 1024) {throw new Error('Image unavailable');}
  return blob;
 })()));
}
