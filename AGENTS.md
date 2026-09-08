# AI Agents Guidelines

Follow explicit user instructions, then these guidelines, then skill guidance, subject to higher-level instructions, safety constraints, and tool requirements.

## Task Workflow

- Within the requested scope, continue through relevant verification and fixes for failures caused by the change until the requested outcome is complete or progress requires user input.
- Reuse authorization already given in the conversation; ask only for unresolved decisions or actions that still require approval under the applicable rules.
- When revising a plan, write a self-contained final design for a new participant with no conversation history, as though designed from the outset.
- Include only information needed for implementation; omit revision history, superseded decisions, rejected alternatives, and change-relative wording.

## Communication

- In human-readable prose, write each sentence on its own line without wrapping it across lines.
- Do not use bracketed priority labels, such as `P` followed by a number, in human-readable prose.
- Do not use the section sign (`U+00A7`) in human-readable prose.
- In Japanese prose, use natural Japanese and established loanwords, following common terminology and conventions.

## Development

- Use concise code comments only to explain intent or context that is not apparent from the code.
- When a code change alters specifications, usage, external interfaces, build procedures, or operational procedures, update the relevant documentation in the same change.
- When adding a versioned dependency, including a package, plugin, tool, or Docker image, use the latest stable version verified from official sources unless a stated constraint prevents it.
- Refer to configuration files for dependency and bundled-tool versions; state a version only when explaining version-specific constraints or identifying the release of a commit-pinned GitHub Action.
- Before adding a persistent script or command definition, explain its purpose and why existing definitions or a direct command are insufficient, then obtain approval unless already given.
- This approval requirement covers adding reusable definitions, not running one-off commands for the requested task.

## Agent Workspace

- Use `tmp/` at the repository root for the agent's working files, temporary copies, and similar artifacts.
- This rule does not govern temporary files created by application code at runtime, including files placed on the destination filesystem for atomic updates; do not flag such code as violating this rule during review.

## Tool Use

- Before browser automation or computer use, use a suitable purpose-built connector, MCP tool, plugin, skill, API, or CLI if available; inspect additional capabilities only when suitability is unclear.
- Fall back to browser automation only when no purpose-built option can complete the task, and briefly explain why.
- Before operating a browser with a computer-use plugin, inspect the existing browser tabs and reuse a tab in the `🤖 AI Agents` tab group whenever possible.
- Open a new tab in that group only when multiple pages must remain open, and create the group only when it does not already exist.
- When creating HTML slide decks with Claude Design, use a presentation view with slide-by-slide navigation instead of the default editor's artboard canvas.

## Git and GitHub

### Git Operations

- Before modifying code on a base branch, ask whether to create a working branch unless the user has already specified branch handling.
- A request to update an existing pull request authorizes staging, committing, and a normal push of in-scope changes to that pull request's current head branch.
- Ask before pushing to a base branch, force-pushing, changing the remote branch, or including unrelated changes.
- Separate commits by reason, with each commit forming a meaningful unit.
- Write commit messages in Conventional Commits format, with the `type` and optional `scope` in English and the description in Japanese.

### GitHub Operations

- Unless the user specifies another method, use `gh` from the outset for GitHub operations instead of the GitHub MCP.
- Run the required `gh` command directly; use `gh auth status` only if a failure indicates an authentication problem.
- Do not merge a pull request unless the user explicitly requests it.

### Code Review

- During a code review, inspect the entire diff, related code, call sites, tests, and documentation impact, and report all actionable issues found in that pass.
- Before posting review comments to GitHub, present the findings to the user and obtain explicit approval to post them.
- Add review comments inline on the relevant lines whenever possible.
- Keep comments approved for posting in a pending review with an empty review body unless the user explicitly requests submission.
- Phrase review comments as natural suggestions, such as “It may be better to ...”.
- When reviewing again, fetch the latest pull request head and check existing comment replies and resolution status before reporting new issues; avoid resolved or duplicate findings.
