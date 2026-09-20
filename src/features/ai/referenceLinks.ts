export function safeReferenceURL(raw: string): string {
	try {
		if (/\p{Cc}/u.test(decodeURIComponent(raw))) {
			return '';
		}
		const url = new URL(raw);
		if (!['http:', 'https:'].includes(url.protocol) || url.username || url.password) {
			return '';
		}
		const host = url.hostname.toLowerCase().replace(/\.$/, '');
		if (host.startsWith('[')) {
			if (/^\[(?:::|::1|::ffff:.*|f[cd][0-9a-f]{2}:.*|fe[89ab][0-9a-f]:.*)\]$/.test(host)) {
				return '';
			}
		} else {
			if (!host.includes('.') || /\.(localhost|local|localdomain|internal|lan|home|ts\.net)$/.test(host)) {
				return '';
			}
			const octets = host.split('.').map(Number);
			if (octets.length === 4 && octets.every(Number.isInteger)) {
				const [first, second] = octets;
				if (
					first === 0 ||
					first === 10 ||
					first === 127 ||
					first >= 224 ||
					(first === 172 && second >= 16 && second <= 31) ||
					(first === 192 && second === 168) ||
					(first === 169 && second === 254) ||
					(first === 100 && second >= 64 && second <= 127)
				) {
					return '';
				}
			}
		}
		return url.href;
	} catch {
		return '';
	}
}
