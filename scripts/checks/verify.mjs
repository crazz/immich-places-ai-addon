import {parseCheckOptions} from './check-options.mjs';
import {resolveBase} from './inventory.mjs';
import {verificationGates} from './verification-gates.mjs';
import {verify} from './verification-runner.mjs';

const gates = verificationGates(process.cwd()).map(gate => gate.name);
const usage = `Usage: node scripts/checks/verify.mjs [--base <revision>] [--gate <name>]\nGates: ${gates.join(', ')}`;

function main() {
	const args = process.argv.slice(2);
	if (args.length === 1 && args[0] === '--help') {
		console.log(usage);
		return;
	}
	const options = parseCheckOptions(args, usage, {allowGate: true});
	if (options['--gate'] && !gates.includes(options['--gate'])) {
		throw new Error(`Unknown gate: ${options['--gate']}\n${usage}`);
	}
	const base = resolveBase(process.cwd(), options['--base']);
	console.log(`base: ${base}`);
	process.exitCode = verify({root: process.cwd(), base, gate: options['--gate']});
}

try {
	main();
} catch (error) {
	console.error(error.message);
	process.exitCode = 1;
}
