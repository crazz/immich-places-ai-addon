export function aiCoverage(profile) {
	const [mode, ...lines] = profile.trim().split(/\r?\n/);
	if (!/^mode: (atomic|count|set)$/.test(mode)) {
		throw new Error('Malformed coverage mode');
	}
	const blocks = new Map();
	for (const line of lines) {
		const match = /^(\S+\.go:\d+\.\d+,\d+\.\d+) (\d+) (\d+)$/.exec(line);
		if (!match) {
			throw new Error('Malformed coverage record');
		}
		const [, block, statements, executions] = match;
		if (!/^immich-places-backend\/(?:ai[^/]*\.go:|internal\/ai(?:adapters)?\/)/.test(block)) {
			continue;
		}
		const prior = blocks.get(block);
		if (prior && prior.statements !== Number(statements)) {
			throw new Error('Malformed coverage: inconsistent statement count');
		}
		blocks.set(block, {statements: Number(statements), covered: Number(executions) > 0 || prior?.covered});
	}
	let covered = 0;
	let total = 0;
	for (const block of blocks.values()) {
		total += block.statements;
		if (block.covered) {
			covered += block.statements;
		}
	}
	if (!total) {
		throw new Error('No AI statements measured');
	}
	const percent = covered / total * 100;
	if (percent < 80) {
		throw new Error(`AI Go statement coverage ${percent.toFixed(2)}% is below 80% (${covered}/${total})`);
	}
	return {covered, total, percent};
}
