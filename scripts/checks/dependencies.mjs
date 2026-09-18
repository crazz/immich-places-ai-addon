import {parseCheckOptions} from './check-options.mjs';
import {runtimeCycles} from './dependency-graph.mjs';
import {dependencyInput} from './dependency-input.mjs';
import {frontendGraph, frontendViolations} from './frontend-dependencies.mjs';
import {goDependencies} from './go-dependencies.mjs';
import {resolveBase} from './inventory.mjs';

const usage = 'Usage: node scripts/checks/dependencies.mjs [--base <revision>]';

function baseGraph(root, base) {
	try {
		return frontendGraph(dependencyInput(root, base));
	} catch (error) {
		throw new Error(`Cannot analyze change base ${base}: ${error.message}`);
	}
}

function main() {
	const args = process.argv.slice(2);
	if (args.length === 1 && args[0] === '--help') {
		console.log(usage);
		return;
	}
	const options = parseCheckOptions(args, usage);
	const base = resolveBase(process.cwd(), options['--base']);
	const graph = frontendGraph(dependencyInput(process.cwd()));
	const previous = baseGraph(process.cwd(), base);
	const baseEdges = new Set(previous.edges.filter(edge => edge.runtime).map(({from, to}) => JSON.stringify([from, to])));
	const go = goDependencies(dependencyInput(process.cwd()));
	const boundaryErrors = [...frontendViolations(graph), ...go.violations];
	for (const error of boundaryErrors) {
		console.error(error);
	}
	let violations = boundaryErrors.length;
	const cycles = runtimeCycles(graph.edges.filter(edge => edge.runtime));
	for (const cycle of cycles) {
		const inherited = cycle.slice(1).every((to, index) => baseEdges.has(JSON.stringify([cycle[index], to])));
		const ai = cycle.some(name => name.startsWith('src/features/ai/'));
		if (inherited && !ai) {
			console.log(`inherited runtime cycle: ${cycle.join(' -> ')}`);
		} else {
			console.error(`${ai ? 'AI' : 'new'} runtime cycle: ${cycle.join(' -> ')}`);
			violations += 1;
		}
	}
	process.exitCode = violations ? 1 : 0;
	console.log(`dependencies: checked ${graph.files.length} frontend files, ${graph.edges.filter(edge => edge.runtime).length} runtime edges, ${go.files.length} Go files against ${base}`);
}

try {
	main();
} catch (error) {
	console.error(error.message);
	process.exitCode = 1;
}
