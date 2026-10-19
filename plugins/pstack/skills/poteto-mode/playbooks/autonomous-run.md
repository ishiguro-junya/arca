### Autonomous run

Follow the [Codex runtime contract](../../../references/codex-runtime.md), including authorization, unavailable capabilities, and state locations.

**You own the exit condition.
Define done, then drive to it without stopping.**

1.
   State the exit condition as a checkable predicate before the first iteration (tests green, repro fixed, all N PRs merged, pixel-diff zero).
2.
   Use the runtime contract's long-running-work procedure.
   Continue in this session toward the completion condition.
   For an explicit later-wakeup request, use an available Codex automation; if unavailable, save a checkpoint and state that no wakeup is scheduled.
3.
   Each iteration makes the smallest change the evidence justifies, verifies it against the predicate, commits if it advanced, discards changes that didn't help.
   Belt-and-suspenders that "might help" gets reverted, not left to ride.
   Sequence the work via the **sequence-verifiable-units** principle skill, verifying each unit before the next instead of batching checks at the end.
4.
   Investigate findings that affect the requested completion condition.
   Fix in-scope causes and verify the result.
   Report unrelated issues separately; do not expand the assignment or open additional PRs without authorization.
   Ask only for unresolved product decisions, missing permissions, or a genuine blocker.
5.
   Checkpoint every iteration via the **show-me-your-work** skill, a row for what changed and whether the predicate moved.
6.
   Stop when the predicate is met.
   A plateau is not a stop, so keep going and pivot your approach to push past it.
   Surface a genuine dead end rather than spinning, and never relax the predicate to declare victory.

**Reply:** the exit condition, iterations run, what landed, what was discarded, final predicate state.
