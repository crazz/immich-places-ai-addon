import {readdirSync} from 'node:fs';
import path from 'node:path';

export function verificationGates(root, base) {
	const node = process.execPath;
	const backend = path.join(root, 'backend');
	const tests = readdirSync(path.join(root, 'scripts/checks')).filter(name => name.endsWith('.test.mjs')).sort().map(name => `scripts/checks/${name}`);
	if (!tests.length) {
		throw new Error('No checker tests found in scripts/checks');
	}
	return [
		{name: 'checker-tests', command: node, args: ['--test', ...tests]},
		{name: 'size', command: node, args: ['scripts/checks/size.mjs', '--base', base]},
		{name: 'dependencies', command: node, args: ['scripts/checks/dependencies.mjs', '--base', base]},
		{name: 'gofmt', command: node, args: ['scripts/checks/format-go.mjs']},
		{name: 'lint', command: node, args: ['node_modules/eslint/bin/eslint.js', '.']},
		{name: 'typegen', command: node, args: ['node_modules/next/dist/bin/next', 'typegen']},
		{name: 'types', command: node, args: ['node_modules/typescript/bin/tsc', '--noEmit']},
		{name: 'go-vet', command: 'go', args: ['vet', '-mod=readonly', './...'], cwd: backend},
		{name: 'go-tests', command: node, args: ['scripts/checks/go-tests.mjs']},
		{name: 'go-build', command: 'go', args: ['build', '-mod=readonly', '-o', path.join(root, 'out/checks/immich-places-backend'), '.'], cwd: backend},
		{name: 'frontend-tests', command: node, args: ['node_modules/vitest/vitest.mjs', 'run', '--coverage']},
		{name: 'frontend-build', command: node, args: ['node_modules/next/dist/bin/next', 'build']},
		{name: 'smoke', command: node, args: ['scripts/checks/smoke.mjs']}
	];
}
