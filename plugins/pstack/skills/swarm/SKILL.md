---
name: swarm
description: "Fan out N parallel workers, drain them, and return one report. Use for /swarm, 'swarm this', or parallel coverage, races, gauntlets, and exploration."
---

Read [the Codex runtime contract](../../references/codex-runtime.md) once per session before using this skill.
It defines tools, models, execution scope, and fallback behavior for every procedure below.

# Swarm

## Start

Open a todolist with one entry per phase before launching anything.

1.
   Frame
2.
   Fan out
3.
   Aggregate
4.
   Report

## Phase A: Frame

1.
   State the done predicate and the artifact or report the swarm must return.
2.
   Choose the shape.
   Partition into slices, race N workers on identical briefs, or mix both.
   For a race or mixed shape, declare `first pass`, `rank all`, or `best-of` before spawning.
3.
   Pick `swarm workers` from `<codex-home>/pstack/models.md`, defaulting to `inherit-parent`.
   For a model race, name each available model and the method before running it.
   Apply the runtime contract's delegation and model rules.
4.
   Give each worker its own writable output when it writes.
   When workers verify or measure commits, each brief names the exact SHAs.
   A measurement brief also names the method (sample count, what one sample is, order).
   The worker records both in its result.

## Phase B: Fan out

Spawn workers within the current concurrency limit, giving each writer an isolated local worktree or disjoint output.
Use the actual subagent tool schema and the role from step 3.
When delegation is unavailable or prohibited, execute the slices sequentially and label the result accordingly.

When a worker must start from a non-default branch, prepare its worktree at that exact branch or commit and include its path in the brief.

Every brief stands alone.
Include the goal, scope, exact slice or race arm, how to verify, and what to report.
Reports use `PASS`, `ISSUES`, or `BLOCKED` with evidence.
A worker that can prove a defect reports `ISSUES` and lists every issue it can prove, not only the first.

If a worker drops out, proceed with N-1 and note it.

## Phase C: Aggregate

Read the terminal results.
Drop a result that does not record the SHAs and method its brief names, and respawn that worker once.
After a second miss, record a gap.
A gap does not count as a pass.
For coverage, every required slice needs a result.
For a race, apply the selection rule declared up front.
Use first pass, rank all, or best-of.
Do not paste raw worker dumps.

## Phase D: Report

Return one consolidated in-chat report with a result table, issue one-liners, gaps or dropouts, and the race rule when used.
