export function parseCheckOptions(args, usage, {allowGate = false} = {}) {
	const options = {};
	for (let index = 0; index < args.length; index += 2) {
		const option = args[index];
		if ((option !== '--base' && !(allowGate && option === '--gate')) || options[option] || !args[index + 1] || args[index + 1].startsWith('--')) {
			throw new Error(usage);
		}
		options[option] = args[index + 1];
	}
	return options;
}
