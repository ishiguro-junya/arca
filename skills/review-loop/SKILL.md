---
name: review-loop
description: Review and fix code changes through repeated complete review passes, stopping when actionable findings are resolved or the review limit is reached. Use when the user asks to iterate on review and fixes or fix findings until none remain; a review-only request does not authorize edits.
---

# Review Loop

Drive a bounded review-and-fix task to a verified result.
Keep the agreed scope and preserve unrelated work.
Report in the user's language.

## Establish the task

- Identify the repository, branch or pull request, resolved comparison base, and requested changes.
  Include applicable staged, unstaged, and untracked files; a committed diff alone may omit the target work.
- Read the applicable instructions, specification, coding standards, and verification procedures.
  Infer unambiguous context from the repository and conversation; ask only when an unresolved choice changes the target or expected behavior.
- Read existing findings and, for a pull request, current replies and resolution states.
  Recheck the remote head before subsequent PR reviews without overwriting local work.
- Identify existing edits, collaborators, and permitted delivery actions.
  The skill does not itself authorize publishing, submitting review comments, or merging.
- Create a compact run record using [the record format](references/record-format.md).
  Keep working records outside tracked deliverables unless requested, using the repository's temporary-file convention or the system temporary directory.

## Count complete review passes

Use a maximum of **5 counted passes**, unless the user specifies a different positive integer.
State the effective limit before starting.
An invalid or ambiguous explicit limit needs clarification; do not silently replace it with the default.

Count a complete review pass provisionally when it starts.
If the pass identifies at least one confirmed actionable finding, remove that pass from the count; repair and review again regardless of the remaining limit.
An unresolved candidate is not a confirmed finding and does not remove a pass from the count.
Reviewers dividing the same snapshot and review scope share one pass.
A full restart after the target changes starts another pass; an interrupted or invalidated pass stays counted unless it identified a confirmed actionable finding.
Targeted reproductions and tests do not consume a pass and cannot substitute for a complete review.
An additional complete review called a final check, another perspective, or an independent review still consumes a pass.
Preserve the counter and limit across handoffs and context compaction; do not reset or extend them implicitly.

## Run a pass

1. **Pin the current state.**
   Record the base, head, included paths, and enough content identity to detect changes, including uncommitted and untracked inputs.
   Do not rely on the head SHA alone for a dirty working tree.
   Keep reviewers read-only during the pass.
2. **Review the full scope.**
   Examine the entire current diff, related code and callers, tests, and documentation impact.
   Cover correctness, agreed behavior, and documented standards; add concerns such as authorization, persistence, or concurrency when the change makes them relevant.
   Review completion requires coverage of the full scope, not just the latest fix.
3. **Validate and consolidate candidates.**
   Apply the finding rules below and group reports by root cause.
   Check existing findings before adding duplicates or reopening a resolved issue.
   Record why a candidate is excluded or deferred outside scope.
4. **Verify the state and required checks.**
   Confirm the reviewed inputs still match and that required reviewers have finished.
   Run applicable checks or establish justified reuse of prior results.
   A failed check with a concrete code cause is an actionable finding; an unavailable check is a blocker, not a pass.
   If inputs changed, invalidate affected results and begin another full pass only if the counted-pass limit permits it.
5. **Choose the outcome.**
   Complete only when this full pass is valid, no new actionable findings remain, all accepted findings are verified resolved, and required checks succeed.
   If actionable findings remain, fix them and verify the changes before starting another pass; the finding-bearing pass does not consume the limit.
   At the limit without confirmed actionable findings, stop without making additional fixes that would lack a subsequent full review.
   If a required decision, reviewer, permission, or verification cannot be obtained, report a blocker rather than claiming success.

Do not start another pass after the completion conditions hold.
If the last counted pass was invalidated or left coverage incomplete, report the limit reached with incomplete review coverage even if no finding has yet been confirmed.
Retain completed fixes and evidence when stopping.
Perform delivery actions only within the user's authorized scope and accurately distinguish delivered work from a verified clean result.

## Judge findings and fixes

Accept a finding when it identifies an expected behavior grounded in the request, specification, existing contract, or documented standard, and provides a reproduction or concrete code path with its impact.
An explicit standard violation can be actionable without a runtime failure.
Do not turn preferences, unsupported possibilities, or unrequested features into mandatory fixes.
Do not invent a new requirement, then rewrite the specification or tests to justify it.

Judge pre-existing bugs by their relationship to the target scope and the change's impact; age alone neither includes nor excludes them.
Report out-of-scope issues separately without silently expanding the completion criteria.
Require new evidence to reopen a resolved finding.
An undecided candidate stays unresolved; do not hide it by calling the review clean.

Fix accepted findings by root cause, including related candidates in the same repair where appropriate.
If a cause recurs, or fixes repeatedly undo one another, re-examine the shared logic and expected behavior before adding more special cases.
Use representative regressions, boundaries, and counterexamples rather than expanding combinations without a requirement.
Do not overwrite, discard, or include unrelated work in a commit.

Reuse a verification result only when its relevant code, dependencies, configuration, and environment remain unchanged and the repository permits reuse.
Record the evidence supporting reuse; an unchanged head alone is insufficient.
Changes require appropriate new verification, not automatic reruns of every unrelated test.

## Coordinate and report

Use separate reviewers when available and authorized, without requiring a particular model, tool, or number of agents.
Give reviewers the pinned scope and raw requirements rather than an expected verdict.
Use one integration owner to apply fixes after the review, and account for every required reviewer's completion.
Without a separate reviewer, run a distinct review phase and describe it accurately rather than claiming independent review.

At each pass boundary and during lengthy work, report total passes and counted passes against the limit, open root causes, new or reopened findings, and remaining verification or blockers.
Use the [record format](references/record-format.md) for the final report, with one outcome:

- **Complete:** a valid full pass found no actionable issues, all accepted findings are verified resolved, and required checks succeeded.
- **Limit reached — incomplete:** no counted passes remain, and review coverage is still unresolved without a confirmed actionable finding.
- **Blocked — incomplete:** a required decision, permission, reviewer, or verification prevents progress.

Include the fixes, exact reviewed state, verification performed and omitted, unresolved findings, and work needed to continue.
Do not count pending verification, absent reviewer responses, or unreviewed changes as zero findings.
