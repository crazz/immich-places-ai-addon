import assert from 'node:assert/strict';
import {mkdirSync, rmSync, symlinkSync} from 'node:fs';
import path from 'node:path';
import test from 'node:test';

import {check, createRepository, git, write} from './dependency-fixture.mjs';

test('rejects executable, adapters and concrete I/O imports from AI core while allowing pure contracts', t => {
	const root = createRepository(t);
	write(root, 'backend/go.mod', 'module example.test/app\n\ngo 1.25\n');
	write(root, 'backend/internal/ai/model/types.go', 'package model\ntype Asset struct { ID string }\n');
	write(root, 'backend/internal/ai/analysis/analysis.go', 'package analysis\nimport "example.test/app/internal/ai/model"\ntype Reader interface { Read() model.Asset }\n');
	const accepted = check(root);
	assert.equal(accepted.status, 0, accepted.stderr);
	assert.match(accepted.stdout, /2 Go files/);
	for (const imported of ['example.test/app', 'example.test/app/internal/adapters/immich', 'example.test/app/internal/ai/adapters/sqlite', 'net/http', 'database/sql', 'database/sql/driver', 'modernc.org/sqlite', 'github.com/mattn/go-sqlite3', 'github.com/hashicorp/go-retryablehttp']) {
		write(root, 'backend/internal/ai/analysis/analysis.go', `package analysis\nimport _ "${imported}"\n`);
		const result = check(root);
		assert.equal(result.status, 1, imported);
		assert.ok(result.stderr.includes(`backend/internal/ai/analysis/analysis.go -> ${imported}: prohibited AI core dependency`), result.stderr);
	}
});

test('rejects direct and transitive analysis access to writer or mutation packages', t => {
	const root = createRepository(t);
	write(root, 'backend/go.mod', 'module example.test/app\n\ngo 1.25\n');
	write(root, 'backend/internal/ai/analysis/analysis.go', 'package analysis\nimport _ "example.test/app/internal/ai/jobs"\n');
	write(root, 'backend/internal/ai/jobs/jobs.go', 'package jobs\ntype Reader interface { Read() string }\n');
	write(root, 'backend/internal/ai/writeback/write.go', 'package writeback\n');
	assert.equal(check(root).status, 0);
	write(root, 'backend/internal/ai/jobs/jobs.go', 'package jobs\nimport _ "example.test/app/internal/ai/writeback"\n');
	const transitive = check(root);
	assert.equal(transitive.status, 1);
	assert.match(transitive.stderr, /analysis\/analysis\.go -> example\.test\/app\/internal\/ai\/jobs -> example\.test\/app\/internal\/ai\/writeback: analysis cannot reach writer/);
	write(root, 'backend/internal/ai/analysis/analysis.go', 'package analysis\nimport _ "example.test/app/internal/ai/mutation"\n');
	const direct = check(root);
	assert.equal(direct.status, 1);
	assert.match(direct.stderr, /analysis\/analysis\.go -> example\.test\/app\/internal\/ai\/mutation: analysis cannot reach writer/);
});


test('fails closed on Go module, parse and read errors', t => {
	const root = createRepository(t);
	write(root, 'backend/core.go', 'package core\n');
	for (const config of ['module example.test/app\ninvalid directive\n', 'go 1.25\n']) {
		write(root, 'backend/go.mod', config);
		const result = check(root);
		assert.equal(result.status, 1, config);
		assert.match(result.stderr, /backend\/go\.mod.*configuration error/);
	}
	rmSync(path.join(root, 'backend/go.mod'));
	assert.match(check(root).stderr, /backend\/go\.mod.*configuration error/);
	write(root, 'backend/go.mod', 'module example.test/app\n\ngo 1.25\n');
	write(root, 'backend/core.go', 'package core\nfunc broken( {\n');
	const malformed = check(root);
	assert.equal(malformed.status, 1);
	assert.match(malformed.stderr, /Go parse\/read error backend\/core\.go/);
	git(root, 'add', '.');
	rmSync(path.join(root, 'backend/core.go'));
	mkdirSync(path.join(root, 'backend/core.go'));
	const unreadable = check(root);
	assert.equal(unreadable.status, 1);
	assert.match(unreadable.stderr, /Go parse\/read error backend\/core\.go/);
});

test('recognizes concrete executables and client drivers while permitting adapter I/O and pure external types', t => {
	const root = createRepository(t);
	write(root, 'backend/go.mod', 'module example.test/app\n\ngo 1.25\n');
	write(root, 'backend/cmd/server/main.go', 'package main\n');
	write(root, 'backend/internal/adapters/transport/client.go', 'package transport\nimport _ "net/http"\n');
	write(root, 'backend/internal/ai/model/types.go', 'package model\nimport _ "github.com/google/uuid"\n');
	assert.equal(check(root).status, 0);
	for (const dependency of ['example.test/app/cmd/server', 'example.test/app/internal/providers/immich', 'github.com/jackc/pgx/v5', 'github.com/lib/pq', 'github.com/go-sql-driver/mysql', 'github.com/jmoiron/sqlx', 'gorm.io/gorm', 'github.com/go-resty/resty/v2']) {
		write(root, 'backend/internal/ai/model/types.go', `package model\nimport _ "${dependency}"\n`);
		const result = check(root);
		assert.equal(result.status, 1, dependency);
		assert.ok(result.stderr.includes(`${dependency}: prohibited AI core dependency`), result.stderr);
	}
});


test('validates a present Go configuration without sources and reports unavailable Go tools', t => {
	const root = createRepository(t);
	write(root, 'backend/go.mod', 'invalid go module\n');
	const malformed = check(root);
	assert.equal(malformed.status, 1);
	assert.match(malformed.stderr, /backend\/go\.mod.*configuration error/);
	write(root, 'backend/go.mod', 'module example.test/app\n\ngo 1.25\n');
	write(root, 'backend/core.go', 'package core\n');
	mkdirSync(path.join(root, 'tools'));
	symlinkSync('/usr/bin/git', path.join(root, 'tools/git'));
	const originalPath = process.env.PATH;
	try {
		process.env.PATH = path.join(root, 'tools');
		const missing = check(root);
		assert.equal(missing.status, 1);
		assert.match(missing.stderr, /spawnSync go ENOENT/);
	} finally {
		process.env.PATH = originalPath;
	}
});

test('rejects direct and transitive selection access to writer packages', t => {
	const root = createRepository(t);
	write(root, 'backend/go.mod', 'module example.test/app\n\ngo 1.25\n');
	write(root, 'backend/internal/ai/selection/selection.go', 'package selection\nimport _ "example.test/app/internal/ai/model"\n');
	write(root, 'backend/internal/ai/model/model.go', 'package model\n');
	assert.equal(check(root).status, 0);
	write(root, 'backend/internal/ai/model/model.go', 'package model\nimport _ "example.test/app/internal/ai/writer"\n');
	const transitive = check(root);
	assert.equal(transitive.status, 1);
	assert.match(transitive.stderr, /selection cannot reach writer/);
	write(root, 'backend/internal/ai/selection/selection.go', 'package selection\nimport _ "example.test/app/internal/ai/writer"\n');
	const direct = check(root);
	assert.equal(direct.status, 1);
	assert.match(direct.stderr, /selection cannot reach writer/);
});
