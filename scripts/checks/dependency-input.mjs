import {readFileSync} from 'node:fs';
import path from 'node:path';

import {currentPaths, readRevisionFile, revisionPaths} from './inventory.mjs';
import {excludedPath} from './source-scope.mjs';

export function dependencyInput(root, revision) {
	const paths = (revision ? revisionPaths(root, revision) : currentPaths(root)).filter(name => !excludedPath(name));
	const names = new Map(paths.map(name => [path.join(root, name), name]));
	const directories = new Set([root]);
	for (const absolute of names.keys()) {
		let directory = path.dirname(absolute);
		while (directory.startsWith(root)) {
			directories.add(directory);
			directory = path.dirname(directory);
		}
	}
	const cache = new Map();
	function readFile(absolute) {
		const name = names.get(path.normalize(absolute));
		if (!name) {
			return undefined;
		}
		if (!cache.has(name)) {
			try {
				cache.set(name, revision ? readRevisionFile(root, revision, name) : readFileSync(absolute, 'utf8'));
			} catch (error) {
				throw new Error(`Cannot read source ${revision ? `${revision}:` : ''}${name}: ${error.message}`);
			}
		}
		return cache.get(name);
	}
	return {
		root,
paths,
		host: {
			useCaseSensitiveFileNames: true,
			fileExists: absolute => names.has(path.normalize(absolute)),
			directoryExists: absolute => directories.has(path.normalize(absolute)),
			readFile,
			readDirectory: () => [...names.keys()].filter(name => /\.[cm]?[jt]sx?$/.test(name)),
			getCurrentDirectory: () => root
		}
	};
}
