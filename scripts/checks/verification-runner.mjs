import {spawnSync} from 'node:child_process';
import {mkdirSync} from 'node:fs';
import path from 'node:path';

import {verificationGates} from './verification-gates.mjs';

export function verify({root, base, gate: selected, execute = spawnSync, report = console.log}) {
	const failures = new Set();
	const gates = verificationGates(root, base).filter(gate => !selected || gate.name === selected || selected === 'types' && gate.name === 'typegen');
	for (const gate of gates) {
		if (gate.name === 'types' && failures.has('typegen')) {
			report('BLOCKED types: typegen failed');
			continue;
		}
		report(`RUN ${gate.name}`);
		try {
			if (gate.name === 'go-build') {
				mkdirSync(path.join(root, 'out/checks'), {recursive: true});
			}
			const result = execute(gate.command, gate.args, {cwd: gate.cwd ?? root, shell: false, stdio: 'inherit'});
			if (result.error) {
				throw result.error;
			}
			if (result.signal) {
				throw new Error(`terminated by ${result.signal}`);
			}
			if (result.status !== 0) {
				throw new Error(`exit ${result.status}`);
			}
			report(`PASS ${gate.name}`);
		} catch (error) {
			failures.add(gate.name);
			report(`FAIL ${gate.name}: ${error.message}`);
		}
	}
	return failures.size ? 1 : 0;
}
