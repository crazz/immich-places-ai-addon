import {fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {ResultDetail} from './ResultDetail';
import {fetchResult} from './resultsApi';
import {resultDetail} from './testing/resultDetail';

vi.mock('./resultsApi', () => ({fetchResult: vi.fn()}));
afterEach(() => vi.resetAllMocks());
it('opens an exact saved unknown outcome with original provenance and inert observations', async () => {
 const detail = resultDetail(); vi.mocked(fetchResult).mockResolvedValue(detail);
 render(<ResultDetail owner={'owner'} reference={{analysisId: detail.entry.analysisId!}} />);
 expect(await screen.findByText('Proposal: unknown')).toBeVisible();
 expect(screen.getByText('<script>Untrusted scene</script>')).toBeVisible();
 fireEvent.click(screen.getByText('Original run provenance'));
 expect(screen.getByText('Model')).toBeVisible();
 expect(screen.getByText(detail.provenance.Model)).toBeVisible();
 expect(screen.getByText('visual-v1')).toBeVisible();
 expect(screen.getByText('Review: unreviewed · Write: not requested')).toBeVisible();
 expect(screen.getByText('No location was established.')).toBeVisible();
 expect(document.querySelector('script')).toBeNull();
 expect(screen.queryByRole('button', {name: /accept|save|apply/i})).not.toBeInTheDocument();
});
