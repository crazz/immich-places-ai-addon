import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {expect, test, vi} from 'vitest';

import {PaginationFooter} from './PaginationFooter';

test('requests the exact numbered page selected by the user', async () => {
	const user = userEvent.setup();
	const onPageChange = vi.fn();
	render(<PaginationFooter currentPage={3} totalPages={7} isLoading={false} onPageChangeAction={onPageChange} />);

	await user.click(screen.getByRole('button', {name: '4'}));

	expect(onPageChange).toHaveBeenCalledExactlyOnceWith(4);
});

test('lets the user reach and activate a numbered page with the keyboard', async () => {
	const user = userEvent.setup();
	const onPageChange = vi.fn();
	render(<PaginationFooter currentPage={1} totalPages={5} isLoading={false} onPageChangeAction={onPageChange} />);

	await user.tab();
	await user.tab();
	await user.tab();
	expect(screen.getByRole('button', {name: '3'})).toHaveFocus();
	await user.keyboard('{Enter}');

	expect(onPageChange).toHaveBeenCalledExactlyOnceWith(3);
});

test('prevents another page request while the collection is loading', async () => {
	const user = userEvent.setup();
	const onPageChange = vi.fn();
	render(<PaginationFooter currentPage={3} totalPages={7} isLoading={true} onPageChangeAction={onPageChange} />);

	for (const button of screen.getAllByRole('button')) {
		expect(button).toBeDisabled();
	}
	await user.click(screen.getByRole('button', {name: '4'}));

	expect(onPageChange).not.toHaveBeenCalled();
});
