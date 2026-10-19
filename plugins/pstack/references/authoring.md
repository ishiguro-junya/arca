# Authoring a Codex skill

Use the available skill-creator skill when present.
Otherwise create a folder containing SKILL.md with YAML `name` and `description` fields.
Use a lowercase hyphenated name and a description that identifies the actual task and trigger.
Keep the main instructions focused on decisions and outcomes; link conditional details from references/.
Use scripts only when deterministic execution is needed and the user has authorized adding reusable commands.
Preserve the user's scope, target repository rules, and approval boundaries.
Put project skills under `.agents/skills/` and personal skills under `<codex-home>/skills/` only when that destination is requested or agreed.
For explicit-only skills, add `policy.allow_implicit_invocation: false` in agents/openai.yaml.
Validate metadata and all local links, then try representative tasks in a disposable workspace.
Record what was executed and what remains unverified.
