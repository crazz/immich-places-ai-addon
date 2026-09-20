import {readFileSync} from 'node:fs';
import {createServer} from 'node:http';
import {setTimeout as delay} from 'node:timers/promises';

const image = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAAEElEQVR4nGIySpkGCAAA//8CAAEvG2k5CAAAAABJRU5ErkJggg==', 'base64');
const users = new Map();
const errors = [];
const blockedGeocoding = [];
const providerRequests = [];
const unknownResult = JSON.parse(readFileSync(new URL('../../backend/internal/ai/results/testdata/ai-analysis-result.unknown.json', import.meta.url), 'utf8'));
const reviewResult = JSON.parse(readFileSync(new URL('../../docs/ai-locate/examples/ai-analysis-result.synthetic.json', import.meta.url), 'utf8'));
reviewResult.descriptions[1].status = 'unavailable';
reviewResult.descriptions[1].text = null;
reviewResult.descriptions[1].unavailable_reason = 'Synthetic translation unavailable.';
const ambiguousResult = structuredClone(reviewResult);
ambiguousResult.outcome = 'ambiguous';
ambiguousResult.selected_candidate_id = null;
const alternative = structuredClone(ambiguousResult.candidates[0]);
alternative.id = 'candidate-2';
alternative.place_name = 'Synthetic alternative viewpoint';
alternative.camera_location.latitude = 51;
alternative.camera_location.longitude = 15;
alternative.subject.location.latitude = 51.001;
alternative.subject.location.longitude = 15.001;
ambiguousResult.candidates.push(alternative);

function catalog() {
	return [
		['11111111-1111-4111-8111-111111111111', '2026-08-01T10:05:00Z'],
		['22222222-2222-4222-8222-222222222222', '2026-08-03T12:00:00Z']
	].map(([id, date]) => ({
		id,
		type: 'IMAGE',
		originalFileName: `${id}.png`,
		originalPath: `/synthetic/${id}.png`,
		fileCreatedAt: date,
		libraryId: null,
		exifInfo: {latitude: null, longitude: null, city: null, state: null, country: null, dateTimeOriginal: date}
	}));
}

function respond(response, status, body) {
	response.writeHead(status, {'Content-Type': 'application/json'});
	response.end(JSON.stringify(body));
}

function recordProviderError(message) {
	errors.push(message);
	throw new Error(message);
}

async function handleProvider(request, response, url, raw) {
	if (request.method !== 'POST' || url.pathname !== '/v1/chat/completions') {
		recordProviderError(`Unexpected provider route: ${request.method} ${url.pathname}`);
	}
	let payload;
	try {
		payload = raw ? JSON.parse(raw) : {};
	} catch {
		recordProviderError('Provider body was not JSON');
	}
	const messages = Array.isArray(payload.messages) ? payload.messages : [];
	const content = messages.flatMap(message => Array.isArray(message.content) ? message.content : []);
	const hasImage = content.some(part => part?.type === 'image_url' && typeof part.image_url?.url === 'string' && part.image_url.url.startsWith('data:image/'));
	if (!hasImage) {
		recordProviderError('Synthetic provider request missing image_url data URL');
	}
	if (payload.stream !== false) {
		recordProviderError('Synthetic provider request must set stream:false');
	}
	if (typeof payload.model !== 'string' || payload.model.length === 0) {
		recordProviderError('Synthetic provider request missing model');
	}
	providerRequests.push({
		path: url.pathname,
		authorization: request.headers.authorization ?? '',
		model: payload.model,
		hasImage: true,
		stream: payload.stream
	});
	const bodyText = raw;
	let assistant = 'color: blue\nshape: circle';
	if (bodyText.includes('"json_object"') || bodyText.includes('"json_schema"')) {
		assistant = '{"color":"blue","shape":"circle"}';
	}
	if (bodyText.includes('schema_version')) {
		const answer = structuredClone(payload.model.endsWith('located') ? reviewResult : payload.model.endsWith('ambiguous') ? ambiguousResult : unknownResult);
		if (bodyText.includes('Research mode')) {
			answer.schema_version = '2.0';
			answer.sources = [{id: 'web-reference', url: 'https://example.org/reference', title: 'Public reference', relevance: 'Matching synthetic facade.'}];
			for (const candidate of answer.candidates) {
				candidate.camera_location.granularity = 'city';
				candidate.camera_location.estimated_radius_m = 5000;
				candidate.camera_location.radius_basis = 'model_estimate';
				candidate.source_refs = ['web-reference'];
			}
		}
		assistant = JSON.stringify(answer);
		providerRequests.at(-1).analysis = true;
		providerRequests.at(-1).outputTokens = payload.max_tokens;
		providerRequests.at(-1).contextHintIncluded = bodyText.includes('Synthetic context hint');
		if (payload.model === 'delayed-second-analysis' && providerRequests.filter(item => item.model === payload.model && item.analysis).length > 1) {
			await delay(5000);
		}
	}
	return respond(response, 200, {
		id: 'cmpl-smoke',
		model: payload.model,
		choices: [{message: {role: 'assistant', content: assistant}, finish_reason: 'stop'}]
	});
}

async function handle(request, response) {
	const url = new URL(request.url, 'http://127.0.0.1:8090');
	if (url.origin !== 'http://127.0.0.1:8090') {
		throw new Error(`Blocked external request: ${request.method} ${url.origin}`);
	}
	if (request.method === 'GET' && url.pathname === '/health') {
		return respond(response, 200, {ready: true});
	}
	if (request.method === 'GET' && url.pathname === '/__provider_state') {
		return respond(response, 200, {requests: providerRequests, errors});
	}
	if (request.method === 'GET' && url.pathname.startsWith('/__state/')) {
		const state = users.get(url.pathname.slice('/__state/'.length));
		return respond(response, 200, {writes: state?.writes ?? [], imageRequests: state?.imageRequests ?? 0, errors, blockedGeocoding});
	}
	if (url.pathname.startsWith('/v1/')) {
		let raw = '';
		for await (const chunk of request) {
			raw += chunk;
			if (raw.length > 300_000) {
				throw new Error('Oversized provider fixture request');
			}
		}
		return handleProvider(request, response, url, raw);
	}
	const key = request.headers['x-api-key'];
	if (typeof key !== 'string' || !key.startsWith('smoke-')) {
		return respond(response, 401, {error: 'Synthetic key required'});
	}
	if (!users.has(key)) {
		users.set(key, {assets: catalog(), writes: [], imageRequests: 0});
	}
	const state = users.get(key);
	if (request.method === 'GET') {
		const asset = state.assets.find(item => url.pathname === `/api/assets/${item.id}`);
		if (asset) {
			return respond(response, 200, {...asset, ownerId: '99999999-9999-4999-8999-999999999999', visibility: 'timeline', isTrashed: false, checksum: Buffer.alloc(20, 1).toString('base64'), updatedAt: '2026-08-04T00:00:00Z'});
		}
		if (url.pathname === '/api/users/me') {
			return respond(response, 200, {id: key});
		}
		if (['/api/libraries', '/api/stacks', '/api/albums', '/api/tags'].includes(url.pathname)) {
			return respond(response, 200, []);
		}
		if (state.assets.some(asset => url.pathname === `/api/assets/${asset.id}/thumbnail`)) {
			state.imageRequests++;
			response.writeHead(200, {'Content-Type': 'image/png'});
			return response.end(image);
		}
	}
	let raw = '';
	for await (const chunk of request) {
		raw += chunk;
		if (raw.length > 100_000) {
			throw new Error('Oversized fixture request');
		}
	}
	const body = raw ? JSON.parse(raw) : {};
	if (request.method === 'POST' && url.pathname === '/api/search/metadata') {
		return respond(response, 200, {assets: {items: state.assets, nextPage: null}});
	}
	if (request.method === 'PATCH' && url.pathname === '/api/assets') {
		state.writes.push(body);
		for (const asset of state.assets) {
			if (body.ids.includes(asset.id)) {
				asset.exifInfo.latitude = body.latitude;
				asset.exifInfo.longitude = body.longitude;
			}
		}
		return respond(response, 204);
	}
	throw new Error(`Unexpected fixture request: ${request.method} ${url.pathname}`);
}

const server = createServer((request, response) => {
	void handle(request, response).catch(error => {
		errors.push(error.message);
		respond(response, 500, {error: error.message});
	});
});
server.on('connect', (request, socket) => {
	// Reload can refresh the frequent-location label after a confirmed save.
	// Deliberately exercise its offline fallback; never tunnel to the public API.
	if (request.url === 'nominatim.openstreetmap.org:443') {
		blockedGeocoding.push(request.url);
	} else {
		errors.push(`Blocked external connection: ${request.url}`);
	}
	socket.end('HTTP/1.1 403 Forbidden\r\n\r\n');
});
server.listen(8090, '127.0.0.1');
process.on('SIGTERM', () => {
	server.close();
	server.closeAllConnections();
});
