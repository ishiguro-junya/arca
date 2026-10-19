---
name: poteto-help
description: Explain how to install, configure, and use the unofficial pstack Codex port, or choose a skill for the user's task.
---

Read [the Codex runtime contract](../../references/codex-runtime.md) once per session before using this skill.

# Poteto help

Answer the actual help question without starting the work it describes.
A request to perform work routes to the requested skill instead.
Read the relevant local skill before explaining it, and link that local file; the upstream public copy describes a different runtime.

## Setup and invocation

Use the [README](../../README.md) for installation, prerequisites, and limitations.
Select `pstack:poteto-mode` from Codex's skill picker or explicitly request poteto-mode by name.
The `/name` notation in inherited playbooks is shorthand for a skill and is not a Codex slash-command guarantee.
The port does not install a sticky mode or startup hook.
Start each new task with the desired skill; within a task, reopen its instructions after compaction as needed.

Setup is optional.
Without `<codex-home>/pstack/models.md`, all roles inherit the current model and reasoning.
[setup-pstack](../setup-pstack/SKILL.md) changes role models, effort, and panel size after checking what the current runtime supports.
Subagents cost extra tokens; project restrictions and user permissions still apply.
Unavailable delegation uses sequential phases, explicitly identified as such.

## Choose a skill

| The user wants to | Skill |
|---|---|
| Do any non-trivial task with rigor | [`/poteto-mode`](../poteto-mode/SKILL.md) |
| Check that the agent understood their goal and the problem before it starts | [`/in-your-own-words`](../in-your-own-words/SKILL.md) |
| Aim a task higher before starting: more expert, more outcome-focused, not overspecified | [`/sharpen`](../sharpen/SKILL.md) |
| Know how code works now, or where new code should live | [`/how`](../how/SKILL.md) |
| Know why code is shaped this way, or where a number came from | [`/why`](../why/SKILL.md) |
| Understand a change or subsystem, explained plainly | [`/teach`](../teach/SKILL.md) |
| Catch up on their own recent work on a topic | [`/recall`](../recall/SKILL.md) |
| Catch up on what changed in this chat since their last message, in plain words and small visuals, with every item waiting on them listed and decidable from the reply alone | [`/ketchup`](../ketchup/SKILL.md) |
| Know what a small diff could break outside itself | [`/blast-radius`](../blast-radius/SKILL.md) |
| Settle types and module shape before code that crosses a function boundary | [`/architect`](../architect/SKILL.md) |
| Get several attempts at one brief, merged into the best one | [`/arena`](../arena/SKILL.md) |
| Run parallel checks over slices, or race workers, as isolated local agents | [`/swarm`](../swarm/SKILL.md) |
| Have different models review a diff and try to break it | [`/interrogate`](../interrogate/SKILL.md) |
| Fix a bug test-first when a cheap local test exists | [`/tdd`](../tdd/SKILL.md) |
| Apply TypeScript rules to `.ts` or `.tsx` work | [`/typescript-best-practices`](../typescript-best-practices/SKILL.md) |
| Strip comments before review, using a reviewer that didn't write them | [`/no-comments`](../no-comments/SKILL.md) |
| Clean AI tells out of prose | [`/unslop`](../unslop/SKILL.md) |
| Write docs, an RFC, a README, a PR description, or a commit message to a standard | [`/technical-writing`](../technical-writing/SKILL.md) |
| Hear the last reply again in plain words | [`/bro`](../bro/SKILL.md) |
| Give agents a scripted way to drive the app and prove behavior | [`/create-verification-skill`](../create-verification-skill/SKILL.md) |
| Bring a verification skill and its feature map back in line with the app | [`/maintain-verification-skill`](../maintain-verification-skill/SKILL.md) |
| Formally verify a race, invariant, or other property that tests rarely reach | [`/create-formal-verification-skill`](../create-formal-verification-skill/SKILL.md) |
| Vet a performance number before reporting or acting on it | [`/benchmark-checklist`](../benchmark-checklist/SKILL.md) |
| Run a large or cross-cutting change, or one to review after stepping away | [`/figure-it-out`](../figure-it-out/SKILL.md) |
| Keep a decision log during a run, and review it afterward | [`/show-me-your-work`](../show-me-your-work/SKILL.md) |
| Pick a model for each role and a reasoning budget | [`/setup-pstack`](../setup-pstack/SKILL.md) |
| Turn their own working habits into a personal mode skill | [`/automate-me`](../automate-me/SKILL.md) |
| Turn what a finished task taught into skill edits | [`/reflect`](../reflect/SKILL.md) |
| Stop agents from repeating the same mistakes in this repo | [`/correct`](../correct/SKILL.md) |
| Find their way around pstack | `/poteto-help` |

Principle skills are internal supporting references that can also be explicitly requested.
The complete list is in the [skill catalog](../../README.md#skills).

## Choose a playbook

Playbooks are procedures inside poteto-mode, not separately installed skills.
Use the user's requested outcome to select one from [poteto-mode](../poteto-mode/SKILL.md).
A one-time PR status question uses Babysit in check mode.
A review-only question does not authorize fixes, cleanup, or posting review comments.
Babysitting does not authorize merging; shipping requires the user's merge authorization.
A planning request remains a plan, and Plan Mode does not run implementation steps.

Long-running work follows the runtime contract.
Explicit later-wakeup requests use available Codex automations; without them, explain that no wakeup is scheduled.
Cursor cloud machines, Custom Modes, make-bot-ui, and Benny are outside this port.

## Customize or recover

Use [automate-me](../automate-me/SKILL.md) for a requested personal style skill and [reflect](../reflect/SKILL.md) for lessons from a completed run.
Resolve skill-creator, code cleanup, and UI/CLI drivers through the runtime contract, not missing Cursor plugins.
For resumed work, use the Session pickup playbook and verify the saved state against the actual repository.
If an expected verification tool is missing, report the unverified part; a green build alone does not prove UI behavior.

Lead with the answer, then one relevant example and the local source link.
