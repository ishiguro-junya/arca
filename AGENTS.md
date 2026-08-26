# AI Agents Guidelines

## General

- When the user gives explicit instructions, prioritize them. Unless they conflict with higher-level instructions or safety constraints, prioritize these guidelines over skill- or tool-specific procedures.
- In human-readable prose, put each sentence on its own line and do not insert line breaks within a sentence.
- Do not use bracketed priority labels, such as `P` followed by a number, in human-readable prose.
- Do not use the section sign (`U+00A7`) in human-readable prose.
- In Japanese prose, prefer common terminology and conventions. Avoid unnecessary mixing of English terms and unnatural literal translations, and use natural Japanese or loanwords established in Japanese.

## Coding

- Add concise code comments when the intent or context would be difficult to understand from the code alone.
- Do not add comments that merely restate what the code does.
- As a rule, do not include fixed versions of dependencies or bundled tools in documentation or code comments. Refer to the configuration files that manage those versions instead. This does not apply when explaining specifications or constraints that depend on a particular version.
- When a code change alters specifications, usage, external interfaces, build procedures, or operational procedures, update the relevant documentation in the same change.
- When specifying a location for temporary files, working copies, or similar artifacts, use the `tmp/` directory at the repository root and do not use another temporary location.

## Tools

- When adding a versioned package, plugin, tool, Docker image, or similar dependency, check official sources for the latest stable version and use it. If compatibility or another constraint prevents this, explain why.
- Before adding a new script or command, explain its purpose, why it is necessary, and why an existing definition or direct command is insufficient, then ask the user for approval.
- Unless the user specifies another method, use `gh` from the outset for GitHub operations instead of the GitHub MCP.
- Do not run `gh auth status` as a preliminary check. Run the required `gh` command first. If it fails, inspect the error and resulting state, and investigate authentication only when it appears to be the cause.

## Git

- Do not stage or unstage files, commit, or push unless the user explicitly requests it.
- Keep changes with different reasons in separate commits, and divide commits into meaningful units.
- Write commit messages in Conventional Commits format, with the `type` and optional `scope` in English and the description in Japanese.

## Pull Requests

- Do not merge a pull request unless the user explicitly requests it.
- During a review, inspect the entire diff as well as related code, call sites, tests, and documentation impact. Identify as many issues as can reasonably be found at that time in a single review pass.
- Add review comments inline on the relevant lines whenever possible.
- Unless the user explicitly requests submission, do not submit review comments. Keep them in a pending review with an empty body.
- Phrase review comments as natural suggestions, such as “It may be better to ...,” rather than requests such as “Could you ...?”
- When reviewing again after changes, fetch the latest pull request head and inspect replies to existing comments and their resolution status. Do not add comments for resolved issues or duplicate existing comments.
