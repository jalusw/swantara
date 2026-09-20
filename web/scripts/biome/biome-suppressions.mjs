import { readdirSync, readFileSync, statSync } from "node:fs";
import { dirname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), "..", "..");
const ALLOWLIST_PATH = join(ROOT, "scripts", "biome", "biome-suppressions.allowlist.json");

const IGNORED_DIRS = new Set([
  "node_modules",
  ".next",
  ".git",
  "coverage",
  "dist",
  "build",
  ".turbo",
  "playwright-report",
]);
const IGNORED_FILES = new Set([ALLOWLIST_PATH]);
const SOURCE_EXTS = new Set([".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts"]);

function walk(dir) {
  const out = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (entry.startsWith(".") && entry !== ".") continue;
    const stat = statSync(full);
    if (stat.isDirectory()) {
      if (!IGNORED_DIRS.has(entry)) out.push(...walk(full));
    } else {
      if (SOURCE_EXTS.has(entry.slice(entry.lastIndexOf(".")))) out.push(full);
    }
  }
  return out;
}

const files = walk(ROOT).filter((f) => !IGNORED_FILES.has(f));

function withForwardSlashes(p) {
  return p.replaceAll("\\", "/");
}
const SUPPRESSION_RE = /biome-ignore(?:-all|-line|-range)?\s+([^:\r\n]+)/g;

/** @type {Map<string, { line: number; rules: string[]; file: string }>} */
const found = new Map();

for (const file of files) {
  const rel = withForwardSlashes(relative(ROOT, file));
  const source = readFileSync(file, "utf8");
  const lines = source.split(/\r?\n/);
  lines.forEach((line, index) => {
    SUPPRESSION_RE.lastIndex = 0;
    for (const match of line.matchAll(SUPPRESSION_RE)) {
      const rules = match[1].trim().split(/\s+/).filter(Boolean);
      found.set(`${rel}:${index + 1}:${rules.join(" ")}`, {
        file: rel,
        line: index + 1,
        rules,
      });
    }
  });
}

const entries = JSON.parse(readFileSync(ALLOWLIST_PATH, "utf8")).entries;
const used = new Set();

const errors = [];
for (const hit of found.values()) {
  const allowed = entries.findIndex(
    (entry) =>
      entry.path === hit.file &&
      entry.rules.every((rule) => hit.rules.includes(rule)) &&
      hit.rules.every((rule) => entry.rules.includes(rule)),
  );
  if (allowed === -1) {
    errors.push(`\n  ${hit.file}:${hit.line}: ${hit.rules.join(" ")}\n    not in the allowlist.`);
  } else {
    used.add(allowed);
  }
}

// Stale allowlist entries (suppression removed or rule no longer suppressed).
const stale = entries
  .map((entry, index) => ({ entry, index }))
  .filter(({ index }) => !used.has(index))
  .map(({ entry }) => `\n  ${entry.path}: ${entry.rules.join(" ")}`);

if (errors.length > 0) {
  console.error(
    `\nbiome-suppressions: ${errors.length} suppression(s) not justified in the allowlist.\n` +
      "Add a reason to scripts/biome/biome-suppressions.allowlist.json BEFORE using biome-ignore." +
      errors.join(""),
  );
  process.exit(1);
}

if (stale.length > 0) {
  console.warn(
    `\nbiome-suppressions: ${stale.length} stale allowlist entr(ies) no longer match any suppression.\n` +
      "Remove them to keep the list honest." +
      stale.join(""),
  );
}

console.log(`biome-suppressions: OK — ${found.size} suppression(s) checked, all justified.`);
