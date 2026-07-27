/**
 * Copies the curated dataset into public/ so the built site can serve it as a
 * downloadable file (the footer and the methodology page link to it).
 *
 * The app itself imports ../data/atlas.json directly through a Vite alias, so
 * there is still exactly one source of truth; this copy is for readers who want
 * the raw data, and it is git-ignored.
 *
 * Runs automatically from `npm run dev` and `npm run build`.
 */
import { copyFileSync, mkdirSync, existsSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const repoRoot = resolve(here, '../..')
const outDir = resolve(here, '../public/data')

const files = [
  ['data/atlas.json', 'atlas.json'],
  ['swiss_layers.json', 'swiss_layers.json'],
]

mkdirSync(outDir, { recursive: true })

for (const [from, to] of files) {
  const src = resolve(repoRoot, from)
  if (!existsSync(src)) {
    console.error(`copy-data: missing ${from}, the build needs it, aborting.`)
    process.exit(1)
  }
  copyFileSync(src, resolve(outDir, to))
  console.log(`copy-data: ${from} -> public/data/${to}`)
}
