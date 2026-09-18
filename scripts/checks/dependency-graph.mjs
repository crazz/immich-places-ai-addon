export function runtimeCycles(edges) {
	const adjacency = new Map();
	for (const {from, to} of edges) {
		if (!adjacency.has(from)) {
			adjacency.set(from, new Set());
		}
		adjacency.get(from).add(to);
	}
	const cycles = [];
	for (const {from, to} of edges) {
		const queue = [[to]];
		const seen = new Set();
		while (queue.length) {
			const route = queue.shift();
			const last = route.at(-1);
			if (last === from) {
				cycles.push([from, ...route]);
				break;
			}
			if (seen.has(last)) {
				continue;
			}
			seen.add(last);
			for (const next of adjacency.get(last) ?? []) {
				queue.push([...route, next]);
			}
		}
	}
	return cycles;
}
