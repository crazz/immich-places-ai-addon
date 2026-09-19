import {spawnSync} from 'node:child_process';
import {mkdtempSync, rmSync} from 'node:fs';
import {fileURLToPath} from 'node:url';

import {fixtureEnvironment, git, write} from './git-fixture.mjs';

export {git, write};

const cli = fileURLToPath(new URL('./dependencies.mjs', import.meta.url));

export function createRepository(t) {
	const root = mkdtempSync('/tmp/engineering-dependencies-');
	t.after(() => rmSync(root, {recursive: true, force: true}));
	git(root, 'init', '-b', 'main');
	git(root, 'config', 'user.email', 'fixture@example.test');
	git(root, 'config', 'user.name', 'Fixture');
	write(root, 'tsconfig.json', JSON.stringify({compilerOptions: {module: 'esnext', moduleResolution: 'bundler', paths: {'@/*': ['./src/*']}}}));
	write(root, 'src/base.ts', 'export {};\n');
	git(root, 'add', '.');
	git(root, 'commit', '-m', 'Fixture base');
	return root;
}

export function check(root, ...args) {
	return spawnSync(process.execPath, [cli, ...args], {cwd: root, encoding: 'utf8', env: fixtureEnvironment()});
}
