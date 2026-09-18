import {execFileSync} from 'node:child_process';

export function git(root, args) {
	return execFileSync('git', args, {cwd: root, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe']});
}

export function currentPaths(root) {
	const deleted = new Set(git(root, ['ls-files', '--deleted', '-z']).split('\0'));
	return [...new Set(git(root, ['ls-files', '--cached', '--others', '--exclude-standard', '-z']).split('\0').filter(Boolean))].filter(name => !deleted.has(name));
}

export function resolveBase(root, revision) {
	if (revision === undefined) {
		for (const branch of ['refs/remotes/origin/main', 'refs/heads/main']) {
			try {
				return git(root, ['merge-base', 'HEAD', branch]).trim();
			} catch {
				continue;
			}
		}
		throw new Error('Cannot resolve main merge base; supply --base <revision>');
	}
	try {
		return git(root, ['rev-parse', '--verify', '--end-of-options', `${revision}^{commit}`]).trim();
	} catch {
		throw new Error(`Invalid change base: ${revision}`);
	}
}

export function revisionPaths(root, revision) {
	return git(root, ['ls-tree', '-r', '-z', '--name-only', revision]).split('\0').filter(Boolean);
}

export function readRevisionFile(root, revision, name) {
	return git(root, ['show', `${revision}:${name}`]);
}

export function renamedPaths(root, base) {
	const fields = git(root, ['diff', '--no-ext-diff', '--name-status', '--find-renames', '-z', base, '--']).split('\0').filter(Boolean);
	const renames = new Map();
	for (let index = 0; index < fields.length;) {
		const status = fields[index++];
		const original = fields[index++];
		if (status.startsWith('R')) {
			renames.set(fields[index++], original);
		}
	}
	return renames;
}

export function hasPathHistory(root, revision, name) {
	return git(root, ['log', '--format=%H', '-1', '--full-history', revision, '--', name]).trim() !== '';
}

export function readRevisionMode(root, revision, name) {
	const entries = git(root, ['--literal-pathspecs', 'ls-tree', '-z', revision, '--', name]).split('\0');
	const entry = entries.find(value => value.slice(value.indexOf('\t') + 1) === name);
	if (!entry) {
		throw new Error(`Cannot establish historical file mode ${revision}:${name}`);
	}
	return Number.parseInt(entry.split(' ')[0], 8);
}
