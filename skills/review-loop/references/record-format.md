# Run Record

Keep a compact record in the repository's approved working area or the system temporary directory.
Use existing task notes when they can hold these fields; no special storage format or script is required.
Preserve the record across handoffs and context compaction.

## Task and snapshots

- Target repository, branch or PR, resolved comparison base, and included changes.
- Sources for expected behavior, repository standards, and required checks.
- Effective counted-pass limit, total passes started and completed, counted passes, and current phase.
- Existing unrelated changes and authorized delivery actions.
- For each pass: head, relevant path set and content identity, including staged, unstaged, and untracked files.
- Full review coverage, reviewer completion, and any snapshot invalidation.

Record a content hash or equivalent diff-and-file evidence for dirty inputs.
Count each pass provisionally when it starts; remove it from the counted total if it identifies a confirmed actionable finding.
An invalidated or interrupted pass without a confirmed finding stays counted; a restart starts another pass.
Several reviewers of one snapshot share a pass, not separate counters.

## Findings

Keep one entry per root cause, with related examples attached to it.

| Field | Content |
| --- | --- |
| ID | Stable identifier within the run. |
| Root cause and location | Faulty behavior or explicit rule violation and affected files or call paths. |
| Expected behavior | Requirement or contract source and the expectation it establishes. |
| Evidence and impact | Reproduction, failed check, or concrete code path showing the problem and its effect. |
| State | Open; fixed, awaiting verification; verified resolved; or out of scope / rejected. |
| Resolution | Fix, verification result and snapshot, or reason for exclusion. |
| History | Pass first found, related duplicates, and evidence for any reopening. |

A candidate awaiting a decision remains open.
Only mark a finding verified resolved after checking the fix on the recorded state.
That local verification does not replace the subsequent full review required for completion.

## Verification

For each relevant check, retain the command or method, covered behavior, snapshot, result, and evidence location.
Distinguish passed, failed, unavailable, and reused results.
For reuse, identify the earlier result and why the relevant code, dependencies, configuration, and environment are unchanged.
For unavailable checks, state the blocker and what would enable verification.
Do not imply that a mocked service, subset of tests, or static inspection exercised an untested integration.

## Progress and final report

Progress should state total passes and `counted passes N / limit`, open root causes, new or reopened findings, and remaining checks or blockers.
Report meaningful changes and lengthy waits rather than only growing test counts.

The final report contains:

1. Outcome: complete, limit reached — incomplete, or blocked — incomplete.
2. Total passes started and completed, counted passes, effective limit, and any invalidated passes.
3. Fixes and the exact reviewed state, with commit or PR links when available.
4. Verification performed, justified reuse, and omitted or unavailable checks.
5. Unresolved findings with evidence, incomplete review coverage, and work or decisions needed to continue.
6. Delivery actions actually completed within the request's scope.

Use the user's language and keep the report proportional to the work.
