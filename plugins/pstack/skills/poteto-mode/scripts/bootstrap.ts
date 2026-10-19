import { createHash } from "node:crypto";
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, realpathSync, renameSync, rmSync, writeFileSync } from "node:fs";
import { homedir } from "node:os";
import { dirname, join, relative, resolve, sep } from "node:path";

function runtimeFiles(directory: string, prefix = ""): string[] {
  return readdirSync(join(directory, prefix), { withFileTypes: true })
    .filter((entry) => entry.name !== "node_modules" && !entry.name.startsWith("."))
    .flatMap((entry) => {
      const path = join(prefix, entry.name);
      return entry.isDirectory() ? runtimeFiles(directory, path) : [path];
    }).sort();
}

export function ensureDependenciesInstalled(): void {
  const source = import.meta.dir;
  const files = runtimeFiles(source);
  const hash = createHash("sha256");
  for (const path of files) {
    hash.update(path).update("\0").update(readFileSync(join(source, path))).update("\0");
  }
  const key = hash.digest("hex");
  const cache = resolve(process.env.PSTACK_CACHE_DIR ??
    join(process.env.XDG_CACHE_HOME || join(homedir(), ".cache"), "pstack", "tools"));
  const target = join(cache, key);
  const ready = () =>
    existsSync(join(target, "node_modules", "commander", "package.json")) &&
    existsSync(join(target, ".pstack-ready")) &&
    readFileSync(join(target, ".pstack-ready"), "utf8") === key;

  if (existsSync(target) && realpathSync(source) === realpathSync(target)) {
    if (!ready()) throw new Error(`Incomplete pstack runtime cache: ${target}`);
    return;
  }
  const entry = relative(realpathSync(source), realpathSync(resolve(process.argv[1])));
  if (entry === ".." || entry.startsWith(`..${sep}`)) {
    throw new Error("The pstack entry point must be inside its scripts directory");
  }
  if (!ready()) {
    mkdirSync(cache, { recursive: true });
    const staging = mkdtempSync(join(cache, ".prepare-"));
    try {
      for (const path of files) {
        const destination = join(staging, path);
        mkdirSync(dirname(destination), { recursive: true });
        copyFileSync(join(source, path), destination);
      }
      const result = Bun.spawnSync(
        [process.execPath, "install", "--frozen-lockfile", "--production", "--ignore-scripts"],
        { cwd: staging, env: { ...process.env, BUN_INSTALL_CACHE_DIR: join(cache, "downloads") }, stdout: "pipe", stderr: "pipe" }
      );
      if (result.exitCode !== 0) {
        process.stderr.write(result.stdout);
        process.stderr.write(result.stderr);
        throw new Error(`pstack dependency installation failed (${result.exitCode})`);
      }
      writeFileSync(join(staging, ".pstack-ready"), key);
      try {
        renameSync(staging, target);
      } catch (error) {
        // Concurrent launches may have completed the identical cache first.
        if (!ready()) throw error;
      }
    } finally {
      rmSync(staging, { recursive: true, force: true });
    }
  }
  const result = Bun.spawnSync([process.execPath, join(target, entry), ...process.argv.slice(2)], {
    cwd: process.cwd(), env: process.env, stdin: "inherit", stdout: "inherit", stderr: "inherit",
  });
  process.exit(result.exitCode ?? 1);
}
