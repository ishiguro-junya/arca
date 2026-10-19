import { expect, test } from "bun:test";
import { chmodSync, existsSync, mkdirSync, mkdtempSync, realpathSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

test("audit preserves worktrees with spaces and never calls an unmerged closed PR safe", () => {
  const root = realpathSync(mkdtempSync(join(tmpdir(), "pstack-worktree-")));
  try {
    const repo = join(root, "main checkout");
    const worktree = join(root, "closed checkout with spaces");
    const bin = join(root, "bin");
    mkdirSync(repo);
    mkdirSync(bin);
    const git = (cwd: string, ...args: string[]) => {
      const result = Bun.spawnSync(["git", "-c", "user.name=Pstack Test", "-c", "user.email=test@example.invalid", ...args], { cwd });
      expect(result.exitCode).toBe(0);
      return result.stdout.toString().trim();
    };
    git(repo, "init", "-b", "main");
    git(repo, "commit", "--allow-empty", "-m", "fixture");
    git(repo, "worktree", "add", "-b", "closed", worktree);
    git(worktree, "commit", "--allow-empty", "-m", "unmerged work");
    const executable = (name: string, body: string) => {
      const path = join(bin, name);
      writeFileSync(path, "#!/bin/sh\n" + body + "\n");
      chmodSync(path, 0o755);
    };
    executable("gh", `printf '%s' '[{"number":1,"state":"CLOSED","headRefName":"closed"}]'`);
    executable("origin", "exit 99");
    executable("python3", 'test -f "$1" || exit 1\necho 0');
    const run = () => Bun.spawnSync(["bash", "worktree-audit.sh", "-b", "main", repo], {
      cwd: import.meta.dir,
      env: { ...process.env, PATH: `${bin}:${process.env.PATH}`, PSTACK_FORGE: "github" },
    });
    const first = run();
    expect(first.exitCode).toBe(0);
    expect(first.stdout.toString()).toContain(`\treview\t${worktree}`);
    expect(first.stdout.toString()).toContain("#1/closed");
    expect(existsSync(worktree)).toBe(true);
    executable("python3", "exit 1");
    expect(run().stdout.toString()).toContain(`\treview-history-unavailable\t${worktree}`);
    executable("gh", `printf '%s' '[{"number":1,"state":"OPEN","headRefName":"closed"}]'`);
    expect(run().stdout.toString()).toContain(`\thold-open-pr\t${worktree}`);
    expect(git(worktree, "status", "--porcelain")).toBe("");
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
