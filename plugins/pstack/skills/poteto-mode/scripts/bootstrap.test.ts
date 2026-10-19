import { afterEach, expect, test } from "bun:test";
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, realpathSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const directories: string[] = [];
afterEach(() => directories.splice(0).forEach((path) => rmSync(path, { recursive: true, force: true })));

test("helpers preserve their package, forward arguments and cwd, and reuse the cache", () => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), "pstack-bootstrap-")));
  directories.push(root);
  const source = join(root, "source with spaces");
  const cache = join(root, "cache with spaces");
  mkdirSync(source);
  mkdirSync(cache);
  const cacheAlias = join(root, "cache alias");
  symlinkSync(cache, cacheAlias);
  for (const name of ["bootstrap.ts", "package.json", "bun.lock"]) {
    copyFileSync(join(import.meta.dir, name), join(source, name));
  }
  const sourceAlias = join(root, "plugin alias");
  symlinkSync(source, sourceAlias);
  const entry = join(source, "fixture.ts");
  writeFileSync(entry, `import {ensureDependenciesInstalled} from './bootstrap.ts';
ensureDependenciesInstalled();
const {Command} = await import('commander');
console.log(JSON.stringify({args:process.argv.slice(2),cwd:process.cwd(),available:!!Command}));`);
  const before = readdirSync(source).map((name) => [name, readFileSync(join(source, name), "utf8")]);
  const run = () => Bun.spawnSync([process.execPath, join(sourceAlias, "fixture.ts"), "argument with spaces"], {
    cwd: root, env: { ...process.env, PSTACK_CACHE_DIR: cacheAlias }, timeout: 15_000,
  });
  const first = run();
  expect(first.exitCode).toBe(0);
  expect(JSON.parse(first.stdout.toString())).toEqual({ args: ["argument with spaces"], cwd: root, available: true });
  const caches = () => readdirSync(cache).filter((name) => /^[a-f0-9]{64}$/.test(name));
  expect(caches().length).toBe(1);
  const key = caches()[0];
  expect(run().exitCode).toBe(0);
  expect(caches()).toEqual([key]);
  expect(existsSync(join(source, "node_modules"))).toBe(false);
  expect(readdirSync(source).map((name) => [name, readFileSync(join(source, name), "utf8")])).toEqual(before);
  writeFileSync(entry, readFileSync(entry, "utf8") + '\nprocess.exitCode = 7;\n');
  expect(run().exitCode).toBe(7);
  expect(caches().length).toBe(2);
}, 30_000);
