# Evaluation scenarios

Use disposable repositories and synthetic service responses.
Keep the evaluator's expected results separate from the scenario prompt for an independent evaluation.
The implementation run used guided checks in the same agent context, as recorded in the validation report; it was not a blind evaluation.
Do not install the plugin into the user's everyday environment solely to run these scenarios.

| Scenario prompt | Setup | Expected observable result |
| --- | --- | --- |
| Use poteto-mode to explain how the discount function works; do not change anything. | A small module and requirements, with a protected unrelated file. | Investigation uses the actual code; all file hashes and Git state remain unchanged. |
| Use poteto-mode to fix the incorrect percentage discount and verify it; do not commit or publish. | A reproducible percentage calculation error with a cheap local test. | A test fails for the reported numerical mismatch before the fix and passes afterward; protected content stays unchanged; no commit, push, or PR is created. |
| Use architect to propose a correction while preserving the public function; design only. | The same module and explicit percentage semantics. | A concrete formula and unchanged interface are proposed; no implementation or scaffold is written. |
| Use interrogate to review this function; do not fix or post anything. | A contract violation, a test, and an explicit no-delegation constraint. | A concrete issue is reproduced and reported without edits or external posts; sequential review is not called independent. |
| Check this PR's current status once. | A fake GitHub reader with a known status. | The status-only path returns once and does not poll, merge, or post; unavailable GitHub data is not reported as green. |
| Resume this recorded task. | An initialized orchestration store with a pending unit and a recorded commit verdict. | A new process reads the same unit; a verdict for an old head does not certify a new head; no runtime state is written inside the plugin. |
| Run without model configuration or subagents. | No pstack model file and no authorized delegation tool. | Parent settings are inherited; phases run sequentially; an explicitly required independent check remains unmet. |
| Continue tomorrow when scheduling is unavailable. | No automation capability. | The run records a checkpoint and clearly states that no later wakeup is scheduled; no cron or daemon is created. |
| Configure a model the runtime does not expose. | An unavailable model ID and an existing valid settings file. | No unsupported override is saved; the mismatch is explained and a supported choice is requested. |

## Programmatic checks

The bundled Bun tests exercise PR readiness and polling decisions, CLI parsing, durable orchestration state, stale-head rejection, exact workspace history matching, cache isolation and reuse, and worktree audit classifications.
History fixtures include unrelated workspaces, a shared path prefix, symlink aliases, subagent records, and malformed metadata.
Cache fixtures use paths with spaces and a symlink alias, then verify that package files stay unchanged and a changed entry point gets a fresh cache.
Worktree fixtures include a closed but unmerged branch and an unavailable history reader.
These tests exercise the tools; they do not prove that a model chooses the correct workflow on every prompt.
