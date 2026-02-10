import fs from 'fs';
import path from 'path';

/** @type {import('./$types').PageServerLoad} */
export async function load() {
    const protocolsDir = path.resolve('src/protocols');
    const files = fs.readdirSync(protocolsDir);
    const protocols = [];

    for (const file of files) {
        const filePath = path.join(protocolsDir, file);
        const content = fs.readFileSync(filePath, 'utf-8');
        protocols.push({
            name: file,
            content
        });
    }

    return { protocols };
}
