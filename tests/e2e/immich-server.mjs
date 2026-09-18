import {createServer} from 'node:http';

const image = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aD1sAAAAASUVORK5CYII=', 'base64');
const users = new Map();
const errors = [];
const blockedGeocoding = [];

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

async function handle(request, response) {
	const url = new URL(request.url, 'http://127.0.0.1:8090');
	if (url.origin !== 'http://127.0.0.1:8090') {
		throw new Error(`Blocked external request: ${request.method} ${url.origin}`);
	}
	if (request.method === 'GET' && url.pathname === '/health') {
		return respond(response, 200, {ready: true});
	}
	if (request.method === 'GET' && url.pathname.startsWith('/__state/')) {
		const state = users.get(url.pathname.slice('/__state/'.length));
		return respond(response, 200, {writes: state?.writes ?? [], errors, blockedGeocoding});
	}
	const key = request.headers['x-api-key'];
	if (typeof key !== 'string' || !key.startsWith('smoke-')) {
		return respond(response, 401, {error: 'Synthetic key required'});
	}
	if (!users.has(key)) {
		users.set(key, {assets: catalog(), writes: []});
	}
	const state = users.get(key);
	if (request.method === 'GET') {
		if (url.pathname === '/api/users/me') {
			return respond(response, 200, {id: key});
		}
		if (['/api/libraries', '/api/stacks', '/api/albums', '/api/tags'].includes(url.pathname)) {
			return respond(response, 200, []);
		}
		if (state.assets.some(asset => url.pathname === `/api/assets/${asset.id}/thumbnail`)) {
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
