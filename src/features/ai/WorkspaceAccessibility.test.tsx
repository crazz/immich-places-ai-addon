import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {expect, it} from 'vitest';

import {AIWorkspace} from './AIWorkspace';

it('describes the dialog and restores keyboard focus after closing', async () => {
	const user = userEvent.setup();
	render(<AIWorkspace owner={'owner'} input={{scope: {view: 'all', gpsFilter: 'all', hiddenFilter: 'visible'}, selected: [], page: [], blockedReason: ''}} />);
	const entry = screen.getByRole('button', {name: 'AI Locate'});
	await user.click(entry);
	expect(screen.getByRole('dialog')).toHaveAccessibleDescription('Preview and authorize image analysis, observe saved jobs, or cancel unfinished work. Closing this window does not stop a job.');
	expect(screen.getByRole('button', {name: 'Close dialog'})).toHaveFocus();
	await user.keyboard('{Escape}');
	expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
	// The entry is outside Radix Trigger; verify focus explicitly when dismissing.
	expect(entry).toHaveFocus();
});
