import assert from 'node:assert/strict';
import test from 'node:test';

import {check, createRepository, write} from './dependency-fixture.mjs';

test('resolves static import types and enforces their AI public boundary', t => {
	const root = createRepository(t);
	write(root, 'src/consumer.ts', "export type Missing = import('@/features/ai/missing').Missing;\n");
	const missing = check(root);
	assert.equal(missing.status, 1);
	assert.match(missing.stderr, /src\/consumer\.ts: unresolved local import @\/features\/ai\/missing/);
	write(root, 'src/features/ai/private.ts', 'export type Private = number; export const value = 1;\n');
	for (const reference of ["import('@/features/ai/private').Private", "typeof import('@/features/ai/private')"]) {
		write(root, 'src/consumer.ts', `export type Stolen = ${reference};\n`);
		const privateImport = check(root);
		assert.equal(privateImport.status, 1);
		assert.match(privateImport.stderr, /src\/consumer\.ts -> src\/features\/ai\/private\.ts.*public integration surface/);
	}
});

test('keeps resolved import types outside runtime cycles', t => {
	const root = createRepository(t);
	write(root, 'src/a.ts', "export type A = import('./b').B; export type Module = typeof import('./b');\n");
	write(root, 'src/b.ts', "import './a'; export type B = number;\n");
	const result = check(root);
	assert.equal(result.status, 0, result.stderr);
	assert.match(result.stdout, /1 runtime edges/);
});

test('follows configured TypeScript import elision for runtime cycles', t => {
	const root = createRepository(t);
	write(root, 'src/a.ts', "import {B} from './b'; export type A = B;\n");
	write(root, 'src/b.ts', "import './a'; export type B = number;\n");
	const erased = check(root);
	assert.equal(erased.status, 0, erased.stderr);
	assert.match(erased.stdout, /1 runtime edges/);
	write(root, 'src/b.ts', "import './a'; export class B {}\n");
	write(root, 'tsconfig.json', '{"compilerOptions":{"module":"esnext","moduleResolution":"bundler","verbatimModuleSyntax":true}}');
	const preserved = check(root);
	assert.equal(preserved.status, 1);
	assert.match(preserved.stderr, /new runtime cycle: src\/a\.ts -> src\/b\.ts -> src\/a\.ts/);
	write(root, 'src/a.ts', "import type {B} from './b'; export type A = B;\n");
	assert.equal(check(root).status, 0);
});
