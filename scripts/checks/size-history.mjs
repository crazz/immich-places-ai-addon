import {parseBaseline, policyPath} from './baseline.mjs';
import {git, hasPathHistory, readRevisionFile} from './inventory.mjs';

export function readSizeBaseline(root, base, basePaths) {
	if (basePaths.has(policyPath)) {
		try {
			return parseBaseline(readRevisionFile(root, base, policyPath));
		} catch (error) {
			throw new Error(`Cannot validate historical size policy ${base}:${policyPath}: ${error.message}`);
		}
	}
	if (hasPathHistory(root, base, policyPath)) {
		throw new Error(`Historical size policy was deleted at ${base}: ${policyPath}`);
	}
	if (git(root, ['rev-parse', '--is-shallow-repository']).trim() !== 'false') {
		throw new Error('Cannot establish size policy history from a shallow repository; fetch the complete base history');
	}
	return null;
}
