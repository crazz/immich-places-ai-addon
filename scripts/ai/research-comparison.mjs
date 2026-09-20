import {createHash} from 'node:crypto';
import {readFileSync} from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';

export function compareResearch(cases) {
	if (!Array.isArray(cases)) {
		throw new Error('Expected an array of matched cases.');
	}
	const rows = cases.flatMap(item => {
		if (
			!item ||
			typeof item.id !== 'string' ||
			!/^[a-f0-9]{64}$/i.test(item.photoSHA256) ||
			typeof item.hint !== 'string' ||
			!Array.isArray(item.runs) ||
			!item.runs.length
		) {
			throw new Error('Each case needs an ID, photo SHA-256, exact hint and runs.');
		}
		if (
			item.reference &&
			(!validCoordinates(item.reference) ||
				typeof item.reference.source !== 'string' ||
				!item.reference.source.trim())
		) {
			throw new Error('Independent coordinates need a reference source.');
		}
		const referenceUncertaintyM = item.reference?.uncertaintyM ?? null;
		if (referenceUncertaintyM !== null && (!Number.isFinite(referenceUncertaintyM) || referenceUncertaintyM < 0)) {
			throw new Error('Invalid reference uncertainty.');
		}
		const measuredReference = referenceUncertaintyM !== null ? item.reference : null;
		for (const run of item.runs) {
			if (
				(run.photoSHA256 !== undefined && run.photoSHA256 !== item.photoSHA256) ||
				(run.hint !== undefined && run.hint !== item.hint)
			) {
				throw new Error('Runs must use the same photograph and hint.');
			}
			if (
				(run.coordinates !== null && !validCoordinates(run.coordinates)) ||
				(run.estimatedErrorM !== null && (!Number.isFinite(run.estimatedErrorM) || run.estimatedErrorM < 0))
			) {
				throw new Error('Invalid coordinates or estimated error.');
			}
		}
		const inputDigest = createHash('sha256')
			.update(JSON.stringify([item.photoSHA256, item.hint]))
			.digest('hex');
		return item.runs.map(run => ({
			...run,
			caseId: item.id,
			inputDigest,
			actualErrorM:
				measuredReference && run.coordinates ? distanceMeters(measuredReference, run.coordinates) : null,
			measurement: measuredReference && run.coordinates ? 'independent_reference' : 'not_measured',
			referenceUncertaintyM,
			referenceSource: item.reference?.source ?? null
		}));
	});
	return {version: 'research-comparison-v1', rows};
}

function validCoordinates(point) {
	return (
		point &&
		Number.isFinite(point.latitude) &&
		Math.abs(point.latitude) <= 90 &&
		Number.isFinite(point.longitude) &&
		Math.abs(point.longitude) <= 180
	);
}

function distanceMeters(a, b) {
	const radians = degrees => (degrees * Math.PI) / 180;
	const latitude = Math.sin(radians(b.latitude - a.latitude) / 2);
	const longitude = Math.sin(radians(b.longitude - a.longitude) / 2);
	const arc =
		latitude * latitude + Math.cos(radians(a.latitude)) * Math.cos(radians(b.latitude)) * longitude * longitude;
	return 6371008.8 * 2 * Math.atan2(Math.sqrt(Math.min(1, arc)), Math.sqrt(Math.max(0, 1 - arc)));
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
	try {
		if (process.argv.length !== 3) {
			throw new Error('Pass one comparison input JSON file.');
		}
		const report = compareResearch(JSON.parse(readFileSync(process.argv[2], 'utf8')));
		process.stdout.write(JSON.stringify(report, null, 2) + '\n');
	} catch {
		process.stderr.write('Invalid comparison input. Check the documented case and run fields.\n');
		process.exitCode = 1;
	}
}
