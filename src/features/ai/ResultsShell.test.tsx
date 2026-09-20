import {fireEvent, render, screen} from '@testing-library/react';
import {afterEach, expect, it, vi} from 'vitest';

import Home from '@/app/page';

import {fetchResults} from './resultsApi';

import type {ReactNode} from 'react';

vi.mock('@/features/auth/AuthContext', () => ({AuthProvider: ({children}: {children: ReactNode}) => <>{children}</>, useAuth: () => ({user: {ID: 'owner'}, hasImmichAPIKey: false, isLoading: false})}));
vi.mock('@/features/auth/AuthMapDynamic', () => ({AuthMapDynamic: () => null}));
vi.mock('@/features/auth/AuthSidebar', () => ({AuthSidebar: () => <p>{'Reconnect Immich'}</p>}));
vi.mock('@/features/map/components/MapViewDynamic', () => ({MapViewDynamic: () => null}));
vi.mock('@/shared/components/PhotoListContainer', () => ({PhotoListContainer: () => null}));
vi.mock('./resultsApi', () => ({fetchResults: vi.fn()}));
afterEach(() => {vi.resetAllMocks(); window.history.replaceState(null, '', '/');});
it('opens retained history from the authenticated shell before catalog or Immich-key readiness', async () => {
 vi.mocked(fetchResults).mockResolvedValue({items: []}); render(<Home />);
 expect(screen.getByText('Reconnect Immich')).toBeVisible();
 fireEvent.click(screen.getByRole('button', {name: 'AI Results'}));
 expect(await screen.findByText('No completed AI runs yet.')).toBeVisible();
 expect(fetchResults).toHaveBeenCalledTimes(1);
});
