import {isUUID} from './selectionTypes';
import {writeOperationRequest} from './writeOperationApi';

import type {TDraft} from './draftTypes';
import type {TWriteOperation} from './writeOperationTypes';

export type TMirrorRetry = {operationId: string; assetId: string; generation: number};
export async function mirrorOperationRequest(owner: string, draft: TDraft, identity: TMirrorRetry, action: 'retry' | 'reconcile', signal?: AbortSignal): Promise<TWriteOperation> {
 if (!isUUID(identity.operationId) || identity.assetId !== draft.assetId || !isUUID(identity.assetId) || !Number.isSafeInteger(identity.generation) || identity.generation < 1) {throw new Error('Invalid metadata recovery identity');}
 const value = await writeOperationRequest(owner, draft, `/${identity.operationId}/targets/${identity.assetId}/metadata/${action}`, {generation: identity.generation}, signal);
 if (value.id !== identity.operationId || value.plan.version !== 'mirror-preview-v4') {throw new Error('Metadata operation changed');}
 return value;
}
