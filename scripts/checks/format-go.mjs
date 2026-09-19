import {execFileSync} from 'node:child_process';
import path from 'node:path';

import {currentPaths} from './inventory.mjs';
import {excludedPath} from './source-scope.mjs';

function main() {
	const args = process.argv.slice(2);
	const usage = 'Usage: node scripts/checks/format-go.mjs';
	if (args.length === 1 && args[0] === '--help') {
		console.log(usage);
		return;
	}
	if (args.length) {
		throw new Error(usage);
	}
	const root = process.cwd();
	const files = currentPaths(root).filter(name => name.endsWith('.go') && !excludedPath(name));
	for (const name of files) {
		const result = execFileSync('gofmt', ['-l', path.join(root, name)], {encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe']});
		if (result.length) {
			console.error(`gofmt required: ${name}`);
			process.exitCode = 1;
		}
	}
	console.log(`gofmt: checked ${files.length} files`);
}

try {
	main();
} catch (error) {
	console.error(`gofmt: ${error.message}`);
	process.exitCode = 1;
}
