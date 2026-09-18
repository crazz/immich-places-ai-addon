import path from 'node:path';

const outputDirectories = new Set(['node_modules', '.next', 'out', 'coverage', '.git', '.gitnexus']);
const manifests = new Set(['package.json', 'go.mod', 'go.sum', 'bun.lock', 'bun.lockb', 'package-lock.json', 'yarn.lock', 'pnpm-lock.yaml']);
const documentExtensions = new Set(['.md', '.mdx', '.rst', '.txt', '.docx', '.pdf']);
const mediaExtensions = new Set(['.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg', '.ico', '.mp4', '.mp3', '.woff', '.woff2']);
const sourceExtensions = new Set(['.go', '.ts', '.tsx', '.mts', '.cts', '.js', '.jsx', '.mjs', '.cjs', '.css', '.sql', '.sh', '.bash', '.zsh', '.fish', '.ps1']);
const fixtureDataExtensions = new Set(['.json', '.yaml', '.yml', '.csv', '.tsv', '.xml', '.gpx']);

export function excludedPath(name) {
	return name === 'backend/coverage.out' || name === 'next-env.d.ts' || name.split('/').some(part => outputDirectories.has(part));
}

export function isHandwrittenSource(name, content, mode = 0) {
	if (excludedPath(name)) {
		return false;
	}
	const extension = path.extname(name).toLowerCase();
	if (sourceExtensions.has(extension) || content.startsWith('#!')) {
		return true;
	}
	if (content.includes('\0')) {
		return false;
	}
	if ((mode & 0o111) !== 0) {
		return true;
	}
	if (manifests.has(path.basename(name)) || documentExtensions.has(extension) || mediaExtensions.has(extension)) {
		return false;
	}
	return !(name.split('/').some(part => ['fixtures', '__fixtures__', 'testdata'].includes(part)) && fixtureDataExtensions.has(extension));
}
