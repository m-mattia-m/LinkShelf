#!/usr/bin/env node
// Reads Vitest's coverage/coverage-summary.json (produced by the
// "json-summary" reporter configured in vitest.config.ts) and writes a
// shields.io-style SVG badge to .badges/<branch>/coverage-frontend.svg at
// the repo root, mirroring the backend's Go coverage badge convention (see
// vladopajic/go-test-coverage in .github/workflows/ci.yaml).
import { readFileSync, mkdirSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { makeBadge } from 'badge-maker'

const frontendRoot = dirname(dirname(fileURLToPath(import.meta.url)))
const repoRoot = dirname(frontendRoot)

const branch = process.argv[2]
if (!branch) {
  console.error('Usage: node scripts/generate-coverage-badge.mjs <branch-name>')
  process.exit(1)
}

const summaryPath = join(frontendRoot, 'coverage', 'coverage-summary.json')
const summary = JSON.parse(readFileSync(summaryPath, 'utf8'))
const pct = summary.total.lines.pct

function colorFor(percentage) {
  if (percentage >= 80) return '#4c1'
  if (percentage >= 60) return '#dfb317'
  return '#e05d44'
}

const svg = makeBadge({
  label: 'coverage',
  message: `${Math.round(pct)}%`,
  color: colorFor(pct)
})

const outDir = join(repoRoot, '.badges', branch)
mkdirSync(outDir, { recursive: true })
writeFileSync(join(outDir, 'coverage-frontend.svg'), svg)

console.log(`Frontend coverage: ${pct}% -> .badges/${branch}/coverage-frontend.svg`)
