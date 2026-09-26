import {lstatSync, readFileSync, statSync} from 'node:fs';
import path from 'node:path';

import {parseBaseline, policyPath} from './baseline.mjs';
import {parseCheckOptions} from './check-options.mjs';
import {currentPaths, readRevisionFile, readRevisionMode, renamedPaths, resolveBase, revisionPaths} from './inventory.mjs';
import {readSizeBaseline} from './size-history.mjs';
import {physicalLines, sizeViolations} from './size-policy.mjs';
import {excludedPath, isHandwrittenSource} from './source-scope.mjs';

const usage = 'Usage: node scripts/checks/size.mjs [--base <revision>]';

function main() {
	const root = process.cwd();
	const args = process.argv.slice(2);
	if (args.length === 1 && args[0] === '--help') {
		console.log(usage);
		return;
	}
	const options = parseCheckOptions(args, usage);
	const base = resolveBase(root, options['--base']);
	const files = currentPaths(root).filter(name => !excludedPath(name)).flatMap(name => {
		let content;
		let metadata;
		try {
			metadata = statSync(path.join(root, name));
			if (metadata.isDirectory() && lstatSync(path.join(root, name)).isSymbolicLink()) {
				return [];
			}
			content = readFileSync(path.join(root, name), 'utf8');
		} catch (error) {
			throw new Error(`Cannot read source ${name}: ${error.message}`);
		}
		return isHandwrittenSource(name, content, metadata.mode) ? [{path: name, lines: physicalLines(content)}] : [];
	});
	const baseline = parseBaseline(readFileSync(path.join(root, policyPath), 'utf8'));
	const basePaths = new Set(revisionPaths(root, base));
	const baseBaseline = readSizeBaseline(root, base, basePaths);
	const renames = renamedPaths(root, base);
	const baseCounts = new Map();
	const inheritedCeilings = new Map();
	for (const name of baseline.keys()) {
		const original = basePaths.has(name) ? name : renames.get(name);
		if (original && basePaths.has(original)) {
			const content = readRevisionFile(root, base, original);
			const count = physicalLines(content);
			baseCounts.set(name, count);
			if (baseBaseline === null && count > 500 && isHandwrittenSource(original, content, readRevisionMode(root, base, original))) {
				inheritedCeilings.set(name, count);
			}
			if (baseBaseline?.has(original)) {
				inheritedCeilings.set(name, baseBaseline.get(original));
			}
		}
	}
	const violations = sizeViolations(files, baseline, baseCounts, inheritedCeilings);
	for (const violation of violations) {
		console.error(violation);
	}
	process.exitCode = violations.length ? 1 : 0;
	console.log(`size: checked ${files.length} files against ${base}`);
}

try {
	main();
} catch (error) {
	console.error(error.message);
	process.exitCode = 1;
}
