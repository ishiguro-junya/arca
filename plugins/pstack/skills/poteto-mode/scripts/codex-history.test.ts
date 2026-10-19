import { afterEach, expect, test } from "bun:test";
import { mkdtempSync, mkdirSync, rmSync, symlinkSync, utimesSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const directories: string[] = [];
afterEach(() => directories.splice(0).forEach((path) => rmSync(path, { recursive: true, force: true })));

test("history selects only exact workspace metadata in activity order", () => {
  const root = mkdtempSync(join(tmpdir(), "pstack-history-"));
  directories.push(root);
  const workspace = join(root, "project with spaces");
  const sessions = join(root, "sessions");
  mkdirSync(workspace);
  mkdirSync(sessions);
  const add = (name: string, cwd: string, modified: number, source: unknown = "cli") => {
    const path = join(sessions, `${name}.jsonl`);
    writeFileSync(path, JSON.stringify({ type: "session_meta", payload: { id: name, cwd, source } }) +
      '\n{"message":"private content must never be emitted"}\n');
    utimesSync(path, modified, modified);
  };
  add("older", workspace, 10);
  add("newer", workspace, 20);
  add("prefix-collision", `${workspace}-other`, 30);
  add("unrelated", join(root, "unrelated"), 40);
  add("child", workspace, 50, { subagent: { thread_spawn: {} } });
  add("relative", "project with spaces", 60);
  writeFileSync(join(sessions, "broken.jsonl"), "not json\n");
  const alias = join(root, "alias");
  symlinkSync(workspace, alias);
  const args = ["python3", join(import.meta.dir, "codex-history.py"), "--workspace", alias, "--sessions-root", sessions];
  const result = Bun.spawnSync(args);
  expect(result.exitCode).toBe(0);
  expect(result.stdout.toString()).not.toContain("private content");
  expect(JSON.parse(result.stdout.toString()).map((row: { id: string }) => row.id)).toEqual(["newer", "older"]);
  expect(Bun.spawnSync([...args, "--latest-activity"]).stdout.toString().trim()).toBe("20");
  expect(JSON.parse(Bun.spawnSync([...args, "--include-subagents"]).stdout.toString()).length).toBe(3);
});

test("absent history has no known activity", () => {
  const result = Bun.spawnSync(["python3", join(import.meta.dir, "codex-history.py"), "--workspace", "/nonexistent-pstack-project", "--sessions-root", "/nonexistent-pstack-sessions", "--latest-activity"]);
  expect(result.exitCode).toBe(0);
  expect(result.stdout.toString().trim()).toBe("0");
});
