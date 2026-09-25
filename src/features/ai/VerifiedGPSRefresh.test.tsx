import {render, waitFor} from '@testing-library/react';
import {expect, it, vi} from 'vitest';

import {useBackend} from '@/shared/context/AppContext';

import {VerifiedGPSRefresh} from './VerifiedGPSRefresh';

vi.mock('@/shared/context/AppContext', () => ({useBackend: vi.fn()}));
it('refreshes catalog pages and markers only after a verified notification', async () => {
 const refresh = vi.fn().mockResolvedValue(undefined);
 vi.mocked(useBackend).mockReturnValue({isReady: true, health: null, backendError: null, isSyncing: false, syncError: null, refreshDataAction: refresh, retryBackendAction: async () => undefined, resyncAction: async () => undefined, fullResyncAction: async () => undefined, clearCatalogAction: () => undefined});
 const view = render(<VerifiedGPSRefresh revision={0} />);
 expect(refresh).not.toHaveBeenCalled();
 view.rerender(<VerifiedGPSRefresh revision={1} />);
 await waitFor(() => expect(refresh).toHaveBeenCalledTimes(1));
 view.rerender(<VerifiedGPSRefresh revision={1} />);
 expect(refresh).toHaveBeenCalledTimes(1);
});
