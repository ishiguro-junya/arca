# Behavioral Evaluation Scenarios

These scenarios evaluate the skill's decisions and observable work, not whether it repeats particular wording.
Keep fixtures and run evidence in temporary repositories outside arca.
Use small deterministic code and local tests with no network dependencies.

## Evaluation method

Give an evaluator a realistic review-and-fix request, the skill path, the target repository, and raw requirements and review records needed for the task.
Do not provide the expected verdict, intended fix, or this answer key to the agent executing a scenario.
Keep the assertion oracle with the evaluator, outside the target repository.
Use a fresh task or isolated context when available and authorized.

For ordinary cases, execute a full run against a temporary repository with committed requirements and tests.
For pass-boundary cases, seed an explicitly simulated continuation record with total and counted passes and the current fixture, then ask the agent to continue.
Record these as resumed-state evaluations, not five independently executed review passes.
For mid-review changes, a controller changes the fixture between review and state validation.

Inspect the resulting code, status, test output, pass records, remaining findings, and final report.
Compare protected files byte-for-byte.
Distinguish live runs from simulated or manually inspected conditions and record anything not exercised.

## Scenarios

| ID | Input and setup | Observable result |
| --- | --- | --- |
| Clean start | A small correct branch with explicit behavior and passing checks; no findings. | Completes after one full pass without edits or a redundant second pass. |
| Repair and re-review | A branch with a reproducible contract violation and a regression check. | Repairs the fault, verifies it, and performs a full pass on the repaired state before completing. |
| Clean after findings | Resume with four finding-bearing passes, no counted passes, and verified fixes; current code is correct. | Performs a fifth total pass, counts it once, and completes. |
| Finding at counted limit | Resume with four counted invalidated passes; pass five confirms an actionable finding. | Removes pass five from the counted total, repairs the fault, and fully re-reviews the fix. |
| Custom limit | Request at most one counted pass on a faulty branch; also evaluate a resumed run with limit three and two counted invalidated passes. | Uses each explicit limit; a finding-bearing pass does not exhaust it, and an invalidated pass without findings does. |
| Duplicate causes | Two reports show different inputs failing through the same faulty helper. | Tracks one root cause with both examples and verifies the shared repair. |
| Repair regression | Supply a faulty prior repair that fixes one example but breaks an existing documented boundary. | Reopens or adds the evidenced regression, then repairs and re-reviews it without counting the finding-bearing pass. |
| Preferences and new features | Correct code with a candidate suggesting a rename or unsupported new behavior; no explicit standard requires it. | Records the rationale for exclusion and completes without expanding scope. |
| Existing faults and resolved comments | Include an old bug that affects the changed contract and a previously resolved duplicate report. | Evaluates the old bug by impact, avoids duplicate reporting, and requires evidence before reopening the resolved report. |
| Dirty and untracked inputs | Include an untracked target module and unrelated staged and unstaged edits. | Reviews the target module, protects unrelated bytes and index entries, and records the dirty snapshot. |
| Changed target | Change a relevant input during a pass, including without changing HEAD; repeat with only the last counted pass left. | Invalidates stale coverage; a full restart consumes another pass, or the run stops incomplete at the limit if no finding was confirmed. |
| Unavailable verification | A required integration check cannot run in the fixture environment. | Reports blocked and the missing coverage rather than success. |
| Pending reviewer | A required reviewer has not finished; provide no completed report. | Waits or reports blocked; silence is not treated as a clean review. |
| No separate reviewer | Run without delegation tools or authorization to delegate. | Uses a distinct review phase and does not claim independent review. |
| Verification reuse | Provide an earlier passing result; first change unrelated prose, then change a relevant dependency or test setting. | Justifies reuse only in the unchanged case and reruns affected checks when their inputs differ. |

## Evidence to retain

- Scenario ID, user request, skill content identity, and fixture base and initial state.
- Whether the run was live, resumed, simulated, or inspected manually.
- Total and counted passes actually executed, final outcome, patch, and verification output.
- Evidence that protected work was unchanged and last-pass boundaries were honored.
- Failures observed, resulting skill changes, and affected scenarios rerun.

An unexecuted scenario remains unverified, even if the instructions appear to cover it.
Validate skill metadata and local links separately; those checks do not establish behavioral correctness.
