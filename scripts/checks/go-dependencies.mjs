import {spawnSync} from 'node:child_process';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

const parser = fileURLToPath(new URL('./go-imports/main.go', import.meta.url));
const adapterSegment = /(?:^|\/)(?:adapters?|transports?|persistence|clients?|handlers?|providers?|immich|sqlite|database)(?:\/|$)/;
const externalIO = [
	'net/http', 'database/sql', 'modernc.org/sqlite', 'github.com/mattn/go-sqlite3',
	'github.com/hashicorp/go-retryablehttp', 'github.com/jackc/pgx', 'github.com/lib/pq',
	'github.com/go-sql-driver/mysql', 'github.com/jmoiron/sqlx', 'gorm.io', 'github.com/go-resty/resty'
];

function packagePath(file, module) {
	const directory = path.posix.dirname(file);
	return directory === 'backend' ? module : module + '/' + directory.slice('backend/'.length);
}

function modulePath(root) {
	const result = spawnSync('go', ['mod', 'edit', '-json', 'backend/go.mod'], {cwd: root, encoding: 'utf8'});
	if (result.status !== 0) {
		throw new Error(`backend/go.mod: configuration error: ${result.error?.message ?? result.stderr.trim()}`);
	}
	const name = JSON.parse(result.stdout).Module?.Path;
	if (!name) {
		throw new Error('backend/go.mod: configuration error: missing module path');
	}
	return name;
}

function writerViolations(imports, module) {
	const packages = new Map();
	for (const {file, imports: dependencies} of imports) {
		const name = packagePath(file, module);
		packages.set(name, [...(packages.get(name) ?? []), ...dependencies]);
	}
	const violations = [];
	for (const {file, imports: dependencies} of imports) {
		const readOnlyScope = /^backend\/internal\/ai\/(analysis|selection)\//.exec(file)?.[1];
		if (!readOnlyScope) {
			continue;
		}
		const queue = dependencies.map(dependency => [dependency]);
		const seen = new Set();
		while (queue.length) {
			const route = queue.shift();
			const dependency = route.at(-1);
			if (seen.has(dependency)) {
				continue;
			}
			seen.add(dependency);
			if (/(?:^|\/)(?:writeback|writers?|mutations?)(?:\/|$)/.test(dependency)) {
				violations.push(`${file} -> ${route.join(' -> ')}: ${readOnlyScope} cannot reach writer`);
			}
			for (const next of packages.get(dependency) ?? []) {
				queue.push([...route, next]);
			}
		}
	}
	return violations;
}

export function goDependencies(input) {
	const files = input.paths.filter(name => name.startsWith('backend/') && name.endsWith('.go'));
	if (!files.length && !input.paths.includes('backend/go.mod')) {
		return {files, violations: []};
	}
	const module = modulePath(input.root);
	const parsed = spawnSync('go', ['run', parser], {cwd: input.root, encoding: 'utf8', input: JSON.stringify(files)});
	if (parsed.status !== 0) {
		throw new Error(`Go import parser failed: ${parsed.error?.message ?? parsed.stderr.trim()}`);
	}
	const imports = JSON.parse(parsed.stdout);
	const executables = new Set(imports.filter(file => file.package === 'main').map(file => packagePath(file.file, module)));
	const violations = [];
	for (const {file, imports: dependencies} of imports) {
		if (!file.startsWith('backend/internal/ai/') || adapterSegment.test(file)) {
			continue;
		}
		for (const dependency of dependencies) {
			if (dependency === module || executables.has(dependency) || adapterSegment.test(dependency) || externalIO.some(prefix => dependency === prefix || dependency.startsWith(prefix + '/'))) {
				violations.push(`${file} -> ${dependency}: prohibited AI core dependency`);
			}
		}
	}
	return {files, violations: [...violations, ...writerViolations(imports, module)]};
}
