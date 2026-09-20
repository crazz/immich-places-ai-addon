import {fireEvent, render, screen} from '@testing-library/react';
import {expect, it, vi} from 'vitest';

import {ResultsFilters} from './ResultsFilters';

it('emits the supported unknown-date filter without a contradictory date range', () => {
 const apply = vi.fn();
 render(<ResultsFilters query={{startDate: '2026-09-01'}} onApplyAction={apply} />);
 fireEvent.click(screen.getByText('Filter results'));
 fireEvent.change(screen.getByLabelText('Capture date'), {target: {value: 'true'}});
 fireEvent.click(screen.getByRole('button', {name: 'Apply filters'}));
 expect(apply).toHaveBeenCalledWith({undated: 'true'});
});
