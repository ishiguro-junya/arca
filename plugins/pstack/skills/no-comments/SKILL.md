---
name: no-comments
description: "Spawn Comment Sicko, fix accepted findings, and offer encodings for claimed constraints."
---

Read [the Codex runtime contract](../../references/codex-runtime.md) once per session before using this skill.

# Comment cleanup

Review the assigned scope with the [comment reviewer](../../references/agents/comment-sicko.md).
Keep a review-only request read-only.
When cleanup is authorized, remove redundant or misleading comments while retaining required notices and useful constraints according to the project's policy.
Use an independent reviewer when available and permitted; otherwise perform a distinct review phase and say so.
Inspect every proposed deletion against the nearby code and its callers.
For uncertain claims, investigate through how or why before deciding.
Do not treat uncertainty as evidence for deletion.
Keep behavior changes and application-code refactors outside this comment pass unless separately in scope.
Run relevant checks when deletions can affect directives, generated files, or executable examples.
Report the actual deletions, retained exceptions, unresolved issues, and verification.
