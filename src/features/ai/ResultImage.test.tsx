import {act, render, screen, waitFor} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import {ResultImage} from './ResultImage';
import {fetchResultThumbnail} from './resultImageApi';
import {resultEntry} from './testing/results';

vi.mock('./resultImageApi', () => ({fetchResultThumbnail: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); vi.unstubAllGlobals();});
it('releases private image URLs and discards late images after owner changes', async () => {
 const create = vi.fn(() => 'blob:private'); const revoke = vi.fn();
 vi.stubGlobal('URL', class extends URL {static createObjectURL = create; static revokeObjectURL = revoke;});
 let resolve: (value: Blob) => void = () => {};
 vi.mocked(fetchResultThumbnail).mockResolvedValueOnce(new Blob(['one'])).mockImplementationOnce(async () => new Promise(done => {resolve = done;}));
 const entry = resultEntry({sourceAvailable: true});
 const view = render(<ResultImage owner={'one'} entry={entry} />);
 expect(await screen.findByRole('img', {name: 'Current source photo'})).toHaveAttribute('src', 'blob:private');
 view.rerender(<ResultImage owner={'two'} entry={entry} />);
 expect(screen.queryByRole('img')).not.toBeInTheDocument();
 await waitFor(() => expect(revoke).toHaveBeenCalledWith('blob:private'));
 view.rerender(<ResultImage owner={'three'} entry={{...entry, sourceAvailable: false}} />);
 await act(async () => resolve(new Blob(['late'])));
 expect(screen.queryByRole('img')).not.toBeInTheDocument();
 expect(screen.getByText('Current image unavailable')).toBeVisible();
 expect(create).toHaveBeenCalledTimes(1);
 expect(fetchResultThumbnail).toHaveBeenCalledTimes(2);
});
