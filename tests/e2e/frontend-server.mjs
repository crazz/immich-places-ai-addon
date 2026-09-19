import {cpSync} from 'node:fs';

// Next.js standalone output intentionally omits public and static assets.
// Match the container's packaging while keeping all copies in generated output.
cpSync('public', '.next/standalone/public', {recursive: true});
cpSync('.next/static', '.next/standalone/.next/static', {recursive: true});
await import('../../.next/standalone/server.js');
