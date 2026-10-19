### Worktree and simulator cleanup

Follow the [Codex runtime contract](../../../references/codex-runtime.md).

1.
   Run the read-only `scripts/worktree-audit.sh [-b <pr-base-ref>] [repo-path]` and record available disk space.
   It enumerates Git's registered worktrees, including paths with spaces, and checks Git, PR, and matching Codex-session metadata.
2.
   Treat every bucket as advice, not permission to delete.
   Cross-check active and pinned chats, running processes, unpushed commits, tracked edits, untracked files, and ignored files.
   A closed, unmerged PR does not establish that its work can be discarded.
3.
   For recent or uncertain worktrees, inspect only matching workspace transcripts and available status tools.
   Confirm the exact set covered by the user's cleanup request; missing history is not evidence of non-use.
4.
   Remove only confirmed unused worktrees whose contents are safe to discard under the request.
   Prefer the host's managed-worktree archive tool for managed checkouts.
   Otherwise use ordinary `git worktree remove <path>` and stop on refused or ambiguous removal instead of forcing it.
5.
   Simulators, application data, and package caches are separate targets.
   Include them only when the user's request covers those exact resources, preserving data and configuration that must remain.
6.
   Re-list the worktrees and report measured disk change, removed or archived paths, and retained paths with reasons.
