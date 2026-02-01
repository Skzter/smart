#!/usr/bin/env bash
# Render all mermaid diagrams in the docs directory

set -euo pipefail

find ../docs/architecture -name '*.mmd' -print0 | while IFS= read -r -d '' file; do
    js_code="
import { readFile } from 'node:fs/promises';
import { renderMermaid } from 'https://esm.sh/beautiful-mermaid@latest'

let contents = await readFile('${file}', 'utf8');
const svg = await renderMermaid(contents);
console.log(svg);
"
    echo "$js_code" | deno run --allow-read - > "${file}.svg"
done