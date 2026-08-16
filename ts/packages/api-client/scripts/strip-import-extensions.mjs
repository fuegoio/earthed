// Strips `.js` extensions from relative import specifiers across the generated
// client. @hey-api/openapi-ts emits `.js` extensions (correct for emitted
// ESM), but this package ships TypeScript source consumed via transpilePackages,
// and Turbopack cannot resolve `.js` to the `.ts` source. Run after `openapi-ts`.
import { readdir, readFile, writeFile } from "node:fs/promises";
import { join, extname } from "node:path";

const ROOT = new URL("../src/generated", import.meta.url).pathname;

async function walk(dir) {
  const out = [];
  for (const entry of await readdir(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) out.push(...(await walk(full)));
    else if (extname(entry.name) === ".ts") out.push(full);
  }
  return out;
}

// Matches `from './x.js'` or `from "../y/z.js"` (any relative path ending in
// `.js` immediately before the closing quote). The `.js` is removed so the
// import resolves to the `.ts` source via the bundler's extension resolution.
const re = /from\s+(['"])(\.{1,2}[^'"]+?)\.js\1/g;

let changed = 0;
for (const file of await walk(ROOT)) {
  const src = await readFile(file, "utf8");
  const next = src.replace(re, "from $1$2$1");
  if (next !== src) {
    await writeFile(file, next);
    changed++;
  }
}
console.log(`strip-import-extensions: processed generated sources (${changed} file(s) updated)`);
