# Review Loop

Review and fix a branch, pull request, or local changes until actionable findings are resolved, with a default limit of five complete review passes.
Use it when you want the agent to carry fixes through regression checks and another review.
A request for review alone does not grant permission to edit or publish changes.

## Install

From the root of an arca checkout, run the appropriate commands below on macOS or Linux.
They link the entire skill folder so its supporting files remain available.
The guard leaves an existing installation untouched; inspect it before choosing to replace it.

### Codex

```sh
mkdir -p "$HOME/.agents/skills"
test ! -e "$HOME/.agents/skills/review-loop" && \
  test ! -L "$HOME/.agents/skills/review-loop" && \
  ln -s "$PWD/skills/review-loop" "$HOME/.agents/skills/review-loop"
```

Codex discovers personal skills in `~/.agents/skills` and supports symlinked skill folders.
If the skill does not appear, restart Codex.
See the [Codex skills documentation](https://developers.openai.com/codex/skills/).

### Claude Code

```sh
mkdir -p "$HOME/.claude/skills"
test ! -e "$HOME/.claude/skills/review-loop" && \
  test ! -L "$HOME/.claude/skills/review-loop" && \
  ln -s "$PWD/skills/review-loop" "$HOME/.claude/skills/review-loop"
```

Claude Code supports personal skills and symlinked skill folders in `~/.claude/skills`.
See the [Claude Code skills documentation](https://code.claude.com/docs/en/skills).

## Use

In Codex:

```text
$review-loop Review and fix the changes on this branch against main.
```

In Claude Code:

```text
/review-loop Review and fix the changes on this branch against main.
```

Specify a positive integer to override the default limit:

```text
$review-loop Review and fix my local changes, using at most 3 complete review passes.
```

The first review counts toward the limit.
Fixes happen between passes, so a limit of one produces a review without a repair cycle.
Findings still present on the last pass are reported without further edits.
The agent reports progress and results in your language.

## Results

| Outcome | Meaning |
| --- | --- |
| Complete | The final full review is clean, accepted findings are verified resolved, and required checks passed. |
| Limit reached — incomplete | Findings or review coverage remain unresolved when no passes are left. |
| Blocked — incomplete | A required decision, permission, reviewer, or verification prevents progress. |

Completed fixes are retained when the run stops without completion.
The result includes the reviewed state, verification results, and any remaining work.

The [agent instructions](SKILL.md) are the source of truth for execution and counting rules.
Maintainers can use the [evaluation scenarios](evals/scenarios.md) to check behavior.
