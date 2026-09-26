import {act, render, screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {afterEach, expect, it, vi} from 'vitest';

import {stackOperation, stagedPreviewDraft} from './testing/writePreview';
import {WriteOperation} from './WriteOperation';

afterEach(() => {vi.unstubAllGlobals(); sessionStorage.clear();});

it('keeps an explicitly inspected stack operation when the initial latest-history detail arrives late', async () => {
 const user = userEvent.setup(); const draft = stagedPreviewDraft(); const latest = stackOperation();
 const inspected = {...latest, id: '55555555-5555-4555-8555-555555555555', approvedAt: new Date(Date.parse(latest.approvedAt) - 1000).toISOString()};
 const history = {items: [latest, inspected].map(value => ({id: value.id, status: value.status, draftRevision: draft.revision, approvedAt: value.approvedAt})), nextCursor: ''};
 let release: (value: Response) => void = () => undefined;
 const request = vi.fn(async (path: string) => {
  if (path.endsWith(`/${latest.id}`)) {return new Promise<Response>(resolve => {release = resolve;});}
  return new Response(JSON.stringify(path.includes('?draftId=') ? history : inspected));
 }); vi.stubGlobal('fetch', request);
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 const selector = await screen.findByRole('combobox', {name: 'Saved GPS operations'});
 await waitFor(() => expect(request.mock.calls.some(([path]) => path.endsWith(`/${latest.id}`))).toBe(true));
 await user.selectOptions(selector, inspected.id);
 await waitFor(() => expect(selector).toHaveValue(inspected.id));
 await act(async () => release(new Response(JSON.stringify(latest))));
 expect(selector).toHaveValue(inspected.id);
 expect(screen.getByText('Stack operation: partial')).toBeVisible();
});

it('reports stack mutation budgets only for their individual targets', async () => {
 const draft = stagedPreviewDraft(); const operation = stackOperation();
 const history = {items: [{id: operation.id, status: operation.status, draftRevision: draft.revision, approvedAt: operation.approvedAt}], nextCursor: ''};
 vi.stubGlobal('fetch', vi.fn(async (path: string) => new Response(JSON.stringify(path.includes('?draftId=') ? history : operation))));
 render(<WriteOperation owner={'owner'} draft={draft} preview={null} />);
 expect((await screen.findByText(/^Approved at /)).textContent).toBe(`Approved at ${operation.approvedAt}`);
 for (const target of operation.targets) {
  expect(within(screen.getByRole('region', {name: `Outcome for ${target.assetId}`})).getByText(`Attempts ${target.attempts} of 2 · Generation ${target.generation}`)).toBeVisible();
 }
});
