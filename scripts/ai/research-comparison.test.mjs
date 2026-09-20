import assert from 'node:assert/strict';
import {test} from 'node:test';

import {compareResearch} from './research-comparison.mjs';

test('keeps matched coarse, unknown and failed outcomes without inventing measured error', () => {
	const runs = [
		{label: 'baseline', status: 'unknown', coordinates: null, estimatedErrorM: null, latencyMs: 50, links: []},
		{
			label: 'research',
			status: 'located',
			coordinates: {latitude: 50, longitude: 14},
			estimatedErrorM: 5000,
			latencyMs: 420000,
			links: ['https://example.org/reference']
		},
		{
			label: 'failed',
			status: 'failed',
			failure: 'timeout',
			coordinates: null,
			estimatedErrorM: null,
			latencyMs: 120000,
			links: []
		}
	];
	const report = compareResearch([
		{
			id: 'synthetic-1',
			photoSHA256: 'a'.repeat(64),
			hint: 'Synthetic hint',
			runs: runs.map(run => ({
				...run,
				provider: 'synthetic',
				model: 'fixture',
				promptVersion: 'fixture-v1',
				schemaVersion: run.label === 'baseline' ? '1.0' : '2.0'
			}))
		}
	]);
	assert.equal(report.rows.length, 3);
	assert.ok(report.rows.every(row => row.actualErrorM === null && row.measurement === 'not_measured'));
	assert.equal(report.rows[1].estimatedErrorM, 5000);
	assert.equal(report.rows[1].latencyMs, 420000);
	assert.deepEqual(report.rows[1].links, ['https://example.org/reference']);
	assert.equal(report.rows[2].failure, 'timeout');
	assert.ok(report.rows.every(row => row.inputDigest === report.rows[0].inputDigest));
});

test('measures independently known camera error separately from the model estimate', () => {
	const report = compareResearch([
		{
			id: 'measured',
			photoSHA256: 'b'.repeat(64),
			hint: '',
			reference: {latitude: 0, longitude: 0, source: 'Survey fixture', uncertaintyM: 20},
			runs: [
				{
					label: 'research',
					status: 'located',
					coordinates: {latitude: 0, longitude: 1},
					estimatedErrorM: 500,
					provider: 'synthetic',
					model: 'fixture',
					promptVersion: 'research-v1',
					schemaVersion: '2.0',
					latencyMs: 10,
					links: []
				}
			]
		}
	]);
	assert.ok(Math.abs(report.rows[0].actualErrorM - 111195) < 5);
	assert.equal(report.rows[0].estimatedErrorM, 500);
	assert.equal(report.rows[0].measurement, 'independent_reference');
	assert.equal(report.rows[0].referenceSource, 'Survey fixture');
});

test('rejects unmatched or invalid measurements instead of producing a misleading comparison', () => {
	const run = {
		label: 'research',
		status: 'located',
		coordinates: {latitude: 0, longitude: 0},
		estimatedErrorM: 500,
		provider: 'synthetic',
		model: 'fixture',
		promptVersion: 'research-v1',
		schemaVersion: '2.0',
		latencyMs: 10,
		links: []
	};
	const item = {id: 'sample', photoSHA256: 'c'.repeat(64), hint: 'same hint', runs: [run]};
	for (const invalid of [
		{...item, photoSHA256: ''},
		{...item, reference: {latitude: 0, longitude: 0, source: ''}},
		{...item, runs: [{...run, coordinates: {latitude: 91, longitude: 0}}]},
		{...item, runs: [{...run, estimatedErrorM: -1}]},
		{...item, runs: [{...run, photoSHA256: 'different'}]},
		{...item, runs: [{...run, hint: 'changed hint'}]}
	]) {
		assert.throws(() => compareResearch([invalid]));
	}
});

test('writes a reproducible JSON report from an explicit local input file', async () => {
	const {mkdtempSync, writeFileSync, rmSync} = await import('node:fs');
	const {tmpdir} = await import('node:os');
	const {join} = await import('node:path');
	const {spawnSync} = await import('node:child_process');
	const directory = mkdtempSync(join(tmpdir(), 'research-comparison-'));
	try {
		const input = join(directory, 'input.json');
		writeFileSync(
			input,
			JSON.stringify([
				{
					id: 'cli',
					photoSHA256: 'd'.repeat(64),
					hint: 'private input',
					runs: [
						{
							label: 'research',
							status: 'unknown',
							coordinates: null,
							estimatedErrorM: null,
							latencyMs: 10,
							links: []
						}
					]
				}
			])
		);
		const result = spawnSync(process.execPath, ['scripts/ai/research-comparison.mjs', input], {encoding: 'utf8'});
		assert.equal(result.status, 0, result.stderr);
		assert.equal(JSON.parse(result.stdout).rows.length, 1);
		assert.ok(!result.stdout.includes('private input'));
	} finally {
		rmSync(directory, {recursive: true, force: true});
	}
});

test('requires known reference uncertainty for measured error and retains that uncertainty', () => {
	const item = {
		id: 'uncertain-reference',
		photoSHA256: 'e'.repeat(64),
		hint: '',
		runs: [{coordinates: {latitude: 0, longitude: 1}, estimatedErrorM: 500}]
	};
	const reference = {latitude: 0, longitude: 0, source: 'Independent survey'};
	const unknown = compareResearch([{...item, reference}]).rows[0];
	assert.equal(unknown.actualErrorM, null);
	assert.equal(unknown.measurement, 'not_measured');
	assert.deepEqual(unknown.coordinates, item.runs[0].coordinates);
	const known = compareResearch([{...item, reference: {...reference, uncertaintyM: 20}}]).rows[0];
	assert.equal(known.referenceUncertaintyM, 20);
	assert.ok(Math.abs(known.actualErrorM - 111195) < 5);
	for (const uncertaintyM of [-1, Infinity]) {
		assert.throws(() => compareResearch([{...item, reference: {...reference, uncertaintyM}}]));
	}
});
