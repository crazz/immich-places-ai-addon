export function physicalLines(content) {
	return content.split('\n').length - 1 + (content.length > 0 && !content.endsWith('\n') ? 1 : 0);
}

export function sizeViolations(files, baseline, baseCounts, baseBaseline) {
	const violations = [];
	const currentCounts = new Map(files.map(file => [file.path, file.lines]));
	for (const [name, ceiling] of baseline) {
		if (!currentCounts.has(name) || currentCounts.get(name) <= 500) {
			violations.push(`${name}: retire the resolved exception from the coding standards`);
		}
		const currentCount = currentCounts.get(name);
		if (currentCount > 500 && currentCount < baseCounts.get(name) && ceiling > currentCount) {
			violations.push(`${name}: lower the ceiling to ${currentCount} after extraction`);
		}
		if (!baseBaseline.has(name) || !baseCounts.has(name)) {
			violations.push(`${name}: exception has no inherited source at the change base`);
		}
		if (baseBaseline.has(name) && ceiling > baseBaseline.get(name)) {
			violations.push(`${name}: ceiling ${ceiling} exceeds inherited ceiling ${baseBaseline.get(name)}`);
		}
	}
	for (const file of files) {
		const ceiling = baseline.get(file.path) ?? 500;
		if (file.lines > ceiling) {
			violations.push(`${file.path}: ${file.lines} lines exceeds ${ceiling}`);
		}
		if (file.lines > 500 && baseCounts.has(file.path) && file.lines > baseCounts.get(file.path)) {
			violations.push(`${file.path}: ${file.lines} lines exceeds change base ${baseCounts.get(file.path)}`);
		}
	}
	return violations;
}
