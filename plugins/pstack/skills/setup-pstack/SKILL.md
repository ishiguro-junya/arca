---
name: setup-pstack
description: Configure pstack's per-role models and reasoning settings when the user asks to set up pstack, change its models, or adjust its budget.
---

Read [the Codex runtime contract](../../references/codex-runtime.md) once per session before using this skill.

# Setup pstack

Configure only `<codex-home>/pstack/models.md`.
All roles work without this file by inheriting the parent model and reasoning.
Do not modify Codex's global model, config.toml, AGENTS.md, or plugin files.

## Inspect and choose

1.
   Read an existing pstack model file, if present, and keep its supported role choices.
2.
   Inspect the actual subagent schema or supported model listing for this session.
   List only confirmed model IDs and their supported reasoning efforts.
   Always offer `inherit-parent` and `auto`, which both omit overrides.
   If discovery is unavailable, inherited operation still works; ask for a confirmed choice only if the user wants an explicit override.
3.
   Ask which roles or panel sizes the user wants to change.
   Show the current choices and offer inheriting reasoning or a supported explicit effort for a selected explicit model.
   Do not encode speed or effort into a model ID, and do not infer a model family from its marketing name.
4.
   Show the complete proposed role table, including panel seat counts, and resolve any unsupported model or effort before saving.
   Every explicit model must be in the detected set; `@effort` is valid only on an explicit model with that supported effort.
   Preserve existing choices not being changed and explain any obsolete role being removed.
   An explicit request naming the complete valid choices is sufficient authorization to save those choices.

## Save

Write a Markdown file with this format and one line per role.
Panel entries are comma-separated; a model's optional effort follows `@` without spaces.
This example leaves every choice inherited:

```text
# pstack models
feature, refactoring: inherit-parent
bug-fix: inherit-parent
perf-issue: inherit-parent
hillclimb: inherit-parent
judgment and prose: inherit-parent
hardest tasks: inherit-parent
how explorer: inherit-parent
how explainer: inherit-parent
why investigators: inherit-parent
why synthesizer: inherit-parent
reflect tooling: inherit-parent
reflect judgment, divergent, synthesizer: inherit-parent
arena runners: inherit-parent, inherit-parent
arena cross-judge pool: inherit-parent
swarm workers: inherit-parent
architect runners: inherit-parent, inherit-parent
interrogate reviewers: inherit-parent, inherit-parent
```

Replace this pstack-owned file atomically after validating the proposed contents.
A missing role falls back to inherited operation; a missing panel line uses the runtime contract's default seat count.
Rerunning setup updates this file rather than appending duplicate role lines.
Explain that the current run rereads it after this change and other runs read it when they next resolve their roles.
Do not promise cross-model review when the selected seats use the same model.

## Optional project verification

If the project has no verification harness, offer once to create one with [create-verification-skill](../create-verification-skill/SKILL.md).
If a formal verification skill would address a relevant property and none exists, offer it once through [create-formal-verification-skill](../create-formal-verification-skill/SKILL.md).
Create either only when requested; setup does not authorize creating project scripts or installing dependencies.
