### Pause safely

Follow the [Codex runtime contract](../../../references/codex-runtime.md), including authorization, unavailable capabilities, and state locations.

**You own a clean stop.
Leave a checkpoint a cold-start agent can resume from.** This is explicit only.
On "keep going", "going to bed, keep going", or "don't stop", do not pause.

1.
   Stop at a safe boundary.
   Finish the current atomic step or back out of it.
   Start nothing new, and cancel any nested subagents.
2.
   Take no irreversible action to pause.
   No PR and no push unless you already had one out.
3.
   Make the work durable within existing permissions.
   Commit only in-scope edits when committing is authorized; otherwise retain them and describe their exact state in the checkpoint.
   Do not publish merely to pause.
4.
   Write the resume note off-context.
   Capture intent, what you were doing, progress and what's verified, current state, next steps, key files, and gotchas.
   For the compaction trigger write it to a file like `/tmp/<slug>-resume.md`.
   If a show-me-your-work trail exists, point at it instead of duplicating it.

**Reply:** where you are in the loop, what's on disk versus still in your head (paths, no diff dumps), the commits you made and whether the tree is clean, and the first action on resume.
This is a pause, not a final report.
