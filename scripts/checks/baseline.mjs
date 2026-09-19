import path from 'node:path';

export const policyPath = 'docs/engineering/coding-standards.md';

export function parseBaseline(markdown) {
	const invalid = reason => new Error(`Invalid inherited size baseline: ${reason}`);
	const headings = [...markdown.matchAll(/^## Inherited size baseline\r?$/gm)];
	if (headings.length !== 1) {
		throw invalid('expected one Inherited size baseline section');
	}
	const section = markdown.slice(headings[0].index + headings[0][0].length).split(/^## /m)[0];
	const lines = section.trim().split(/\r?\n/);
	const header = '| File | Adoption ceiling (physical lines) |';
	const start = lines.indexOf(header);
	if (start < 0 || !/^\|\s*:?-+\s*\|\s*:?-+:?\s*\|$/.test(lines[start + 1] ?? '')) {
		throw invalid('expected the File / Adoption ceiling table and separator');
	}
	const entries = new Map();
	let end = start + 2;
	while (end < lines.length && lines[end].trim() !== '') {
		const match = /^\|\s*`([^`]+)`\s*\|\s*([1-9]\d*)\s*\|$/.exec(lines[end]);
		if (!match) {
			throw invalid(`malformed record: ${lines[end]}`);
		}
		const [, name, count] = match;
		const ceiling = Number(count);
		if (path.posix.isAbsolute(name) || path.posix.normalize(name) !== name || name.startsWith('../') || /[\\:*?[\]{}\0]/.test(name)) {
			throw invalid(`expected an exact repository-relative path: ${name}`);
		}
		if (!Number.isSafeInteger(ceiling) || ceiling <= 500 || entries.has(name)) {
			throw invalid(`duplicate path or invalid ceiling: ${name}`);
		}
		entries.set(name, ceiling);
		end += 1;
	}
	if ([...lines.slice(0, start), ...lines.slice(end)].some(line => line.startsWith('|'))) {
		throw invalid('unexpected or duplicate table');
	}
	return entries;
}
