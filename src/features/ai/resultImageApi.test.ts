import {afterEach, expect, it, vi} from 'vitest';

import {fetchResultThumbnail} from './resultImageApi';
import {resultEntry} from './testing/results';

afterEach(() => vi.unstubAllGlobals());
it('reads only the exact authorized image and rejects invalid content or identities without retrying', async () => {
 const entry = resultEntry();
 const request = vi.fn<typeof fetch>().mockResolvedValueOnce(new Response('jpeg', {headers: new Headers([['Content-Type', 'image/jpeg']])})).mockResolvedValue(new Response('<html>', {headers: new Headers([['Content-Type', 'text/html']])}));
 vi.stubGlobal('fetch', request);
 expect((await fetchResultThumbnail(entry.jobId, entry.id)).type).toBe('image/jpeg');
 expect(String(request.mock.calls[0][0])).toContain(`/ai/jobs/${entry.jobId}/items/${entry.id}/thumbnail`);
 expect(request.mock.calls[0][1]).toMatchObject({method: 'GET', cache: 'no-store'});
 await expect(fetchResultThumbnail(entry.jobId, entry.id)).rejects.toThrow();
 await expect(fetchResultThumbnail('../secret', entry.id)).rejects.toThrow();
 expect(request).toHaveBeenCalledTimes(2);
});
