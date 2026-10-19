# Codex runtime contract

This is the runtime contract for the unofficial `pstack@arca` plugin.
Read it once per session, including when a skill is invoked without poteto-mode.
Use it for all bundled skills, playbooks, reference prompts, and delegated work.

## Scope and precedence

Follow system and developer instructions, the user's request, and the target project's AGENTS.md before these workflows.
A workflow describes how to perform authorized work; it grants no additional authority.
Keep questions and review-only requests read-only, and keep Plan Mode non-mutating.
Skip implementation, cleanup, committing, or PR-opening steps that are outside the request.
Do not treat a playbook's final PR step as permission to create a PR.
Carry existing authorization forward without asking again, and ask only for unresolved decisions or permissions.
External messages, review submissions, merges, deployment, deletion, and force-push follow the host's and user's authorization rules.
Do not turn an adjacent problem, a preference, or a tool failure into permission to change another project or open another PR.
When a required capability or permission is missing, continue independent work and report the affected outcome as incomplete.
Do not fill an unavailable result with a placeholder and treat dependent steps as successful.

## Skills and companion procedures

`<plugin>` is the directory containing this file's parent `references/`, `plugin.json`, and `skills/`.
Pass absolute paths to delegates; resolve internal skill references against `<plugin>/skills/<name>/SKILL.md`.
The `/name` notation retained in playbooks names a bundled skill, not a promise that Codex exposes that slash command.
Users can select the plugin's skill entry, such as `pstack:poteto-mode`, or explicitly request the skill by name.
Read the local file when a playbook routes to a skill; never select an unrelated installed skill with the same name.
Skills that were explicit-only upstream remain explicit-only through `agents/openai.yaml`.
There are no startup hooks, sticky Custom Modes, or changes to AGENTS.md installed by this package.
For a new chat, invoke the desired skill again; after compaction, reopen the active skill and run record as needed.

| Procedure in a playbook | Codex implementation |
| --- | --- |
| `skill-creator` | Read the available Codex skill-creator skill; if absent, use the bundled [authoring procedure](authoring.md). |
| `deslop` | Read the bundled [code cleanup procedure](deslop.md). It is not a separately installed skill. |
| `control-ui` | Use the target project's verification skill and available purpose-built connector, API, CLI, or browser tool, in that order when applicable. |
| `control-cli` | Use the target project's test harness or a terminal/PTY tool with observable input and output. |
| `Task` / delegate | Use the session's actual subagent API and schema, as described below. |
| `AskQuestion` | Use an available structured question tool; otherwise ask a concise question in chat. |
| Todo list | Use the session's planning tool if present; otherwise keep a compact checklist in the run record. |
| MCP discovery | Inspect the session's tool catalog or supported tool-discovery API; do not scan a Cursor-specific directory. |

An unavailable UI driver does not establish that the UI works.
Record the missing verification and the evidence that was actually obtained.
Use GitHub's `gh` CLI for GitHub operations unless the user requests another method.
Use another forge only when the target is actually hosted there or the user has chosen it.
The presence of a CLI executable alone does not change the forge.

## Models

`<codex-home>` means the existing CODEX_HOME value, or `~/.codex` when it is unset.
Read `<codex-home>/pstack/models.md` when first using a role, and reread it after setup-pstack changes it.
Do not alter CODEX_HOME, config.toml, authentication, or the parent chat's model.
No configuration is required: every role defaults to `inherit-parent`, including reasoning effort.
Panel roles default to two independent attempts on the parent model; their cross-judge defaults to one.
`auto` is an alias for `inherit-parent`, not a request to select another model.

The configuration uses the role names in [setup-pstack](../skills/setup-pstack/SKILL.md).
A role contains a model or a comma-separated panel list, with an optional `@effort` on each explicit model.
For example, `arena runners: inherit-parent, inherit-parent` preserves two seats without inventing model IDs.
An explicit model is valid only when the current session's tool schema or supported model listing confirms it.
Keep the model ID and effort separate when calling a tool; do not synthesize Cursor-style model-effort slugs.
Omit model and effort overrides for `inherit-parent` and `auto`.
When a configured model or effort is unavailable, report the mismatch and ask for a supported choice before using that override.
Do not silently select another family or edit configuration to recover from a rejection.
Same-model independent attempts are not cross-model verification.

## Delegation and worktrees

Check the user's and project's delegation restrictions before spawning.
In particular, a project that requires permission for Astra subagents still requires that permission.
Use the actual offered tool family, for example collaboration spawn/message/wait tools or spawn_agent/send_input/wait, and only supported fields.
Do not pass Cursor fields such as `subagent_type`, `readonly`, `environment`, `cloud_base_branch`, or `run_in_background` to a tool that does not accept them.
Do not use user-visible chat creation or messaging tools as a substitute for subagents without the authorization those tools require.

Read [poteto-agent](agents/poteto-agent.md) for playbook implementation delegates and [comment-sicko](agents/comment-sicko.md) for the comment-review role.
Pass the relevant role instructions, absolute skill path, task scope, permissions, output location, and verification requirement in the brief.
A read-only role gets explicit no-edit instructions and a post-run status check; it is not an enforcement flag unless the tool supports one.
By default inherit the parent's model and reasoning.
Only request a configured override when the tool supports it; use a fresh context when the host requires one for overrides.

Use available local worktrees for independent writers and pin their starting branch or commit.
Never reset or discard a dirty checkout to simulate isolation.
Keep one writer per checkout or partition disjoint writes explicitly; reviewers remain read-only.
Concurrency and nesting must stay within the runtime's actual limits, not the upstream cloud-agent limits.
Poll with status/wait tools rather than restarting an agent just to inspect its state.
If delegation is unavailable or not authorized, perform the roles sequentially in this session, preserving the phases and recording that the review was not independent.
If independent or cross-model verification is an explicit acceptance criterion, report it as unmet instead of substituting self-review.

## Long-running work

Within the current session, work toward the user's checkable completion condition using existing process and subagent tools.
Use a goal API only when the user explicitly requested a goal and the host permits it.
For an explicit request to monitor, schedule, or continue later, use the available Codex automation tool, preferring a thread heartbeat when supported.
Reuse a matching automation and stop or disable it when its requested completion condition is met.
Keep unchanged, non-actionable monitor runs quiet.
Do not invent a `/loop` command, install a daemon, or add cron as a fallback.
If scheduling is unavailable, complete what can run in this session, save a checkpoint, and explain that later wakeups are not armed.
An automation request does not authorize unrelated chat creation or external messages.

Use `watch-pr --status-only` for a single PR status question.
It exits 0 after a successful read, or 7 on a query failure without retrying or sleeping.
The regular watcher is for an authorized continuing watch; do not block a one-time question in a polling loop.
It waits while GitHub reports `BEHIND` or `UNKNOWN`, including an unknown `mergeable` result, instead of declaring the PR ready.
Existing Bugbot comments and checks may be read through GitHub; starting or configuring a Cursor service is unsupported.

## State, transcripts, and helpers

Keep runtime state outside the installed plugin.
For a resumable run, use `<codex-home>/pstack/runs/<run-id>/`, recording the exact repository path and shared store path in the run brief.
Use an existing user-selected state directory when one was supplied.
Pass the run's orchestration directory explicitly with `orch --store <directory>` or ORCH_STORE.
Only the existing bootstrap helper's dependency cache uses `<cache>/pstack/tools/`; `<cache>` is XDG_CACHE_HOME or `~/.cache`.
The PSTACK_CACHE_DIR override selects a disposable cache for testing.
The first helper execution installs locked dependencies in that cache, not in the plugin.
Bun and shell prerequisites are documented in the [README](../README.md).
In poteto-mode playbooks, `scripts/` refers to `<plugin>/skills/poteto-mode/scripts/`; pass the absolute helper path while keeping the working directory at the target repository.
The show-me-your-work logger instead lives in that skill's own `scripts/` directory.

For chat history, prefer available Codex list/read-chat tools and filter by the requested workspace and time window.
For local CLI history, use `python3 <plugin>/skills/poteto-mode/scripts/codex-history.py --workspace <absolute-path>` to list matching sessions.
It reads only each file's metadata record until the canonical workspace path matches.
Then read only relevant matching conversations; exclude the current chat when mining prior work.
Session paths are not Cursor transcript paths, and UUID ordering is not chronology.
Do not search the contents of unrelated projects' transcripts.
Keep retained run records and facts distinct from verified current repository state.

## Unsupported features

This port does not implement Cursor cloud machines, Custom Modes, the Grok Bot webhook UI (`make-bot-ui`), or Benny's Cursor Automations service.
Local subagents and worktrees cover ordinary delegation; Codex automation tools cover supported scheduled work.
If the user specifically requires a missing service, explain the gap without creating an unrequested replacement backend.
