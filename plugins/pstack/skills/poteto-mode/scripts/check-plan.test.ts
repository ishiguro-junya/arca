import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { expect, test } from "bun:test";

const checker = join(import.meta.dir, "check-plan.mjs");
const rule =
	"Tests alone are not sufficient verification. A PR is verified only when its unit, live, and perf boxes are all checked.";

const validPlan = `# Demo

## How to read this
One box is one unit of work and names the evidence. Check a box only when its evidence exists. See playbooks/.
${rule}

## Program checklist
### Arm the program
### Spawn owners
### PR mechanics
### Verdict and merge
### Boot recipe
An explicitly requested hourly Codex heartbeat posts a status message only for changes.

## PR 1
**Depends on.** None.
**Files.**
- [ ] Do the thing.
**Build.**
- [ ] Do the thing.
**You see.**
- [ ] Do the thing.
**Verify, unit.** ${rule}
- [ ] Check the thing.
**Verify, live.** ${rule} Ten lanes on \`example-model\` at the PR head
${Array.from({ length: 10 }, (_, i) => `- [ ] Lane ${i + 1}. Save \`lane-${i + 1}.png\`. Pass when evidence exists.`).join("\n")}
**Verify, perf.** ${rule}
- [ ] Metric. Recorded.
- [ ] Probe. Recorded.
- [ ] Baseline. Recorded.
- [ ] Rule. Recorded.
**Review gate.** None.
**Merge.**
- [ ] Merge when approved.

## Close the program
Close.

## Appendix Prototype evidence
Recorded.
`;

const runPlan = (example: string, expectedExit: number, plan = validPlan, command = checker) => {
	const directory = mkdtempSync(join(tmpdir(), "check-plan-test-"));
	const file = join(directory, "plan with spaces.md");
	writeFileSync(file, plan + example);
	try {
		const result = spawnSync("node", [command, file], { cwd: directory, encoding: "utf8" });
		if (expectedExit === 0) expect(result.stderr).toBe("");
		expect(result.status).toBe(expectedExit);
		return result;
	} finally {
		rmSync(directory, { recursive: true, force: true });
	}
};

test("accepts an hourly Codex heartbeat without legacy loop commands or repository-local procedures", () => {
	runPlan("", 0);
});

test("rejects a program without its Codex heartbeat", () => {
	const result = runPlan("", 1, validPlan.replace("hourly Codex heartbeat", "/loop 1h"));
	expect(result.stderr).toContain('Program checklist lacks "hourly Codex heartbeat"');
});

test("the documented checker validates the bundled template from an unrelated workspace", () => {
	const plugin = resolve(import.meta.dir, "../../..");
	const playbook = readFileSync(join(plugin, "skills/poteto-mode/playbooks/multi-phase-plan.md"), "utf8");
	const populated = playbook
		.replaceAll("<plugin>", plugin)
		.replaceAll("<execution playbook>", "autopilot-full")
		.replaceAll("<swarm workers model>", "inherit-parent")
		.replaceAll("<each other leaf skill the program uses>", "how/SKILL.md");
	const template = populated.match(/````markdown\n([\s\S]*?)\n````/)?.[1];
	const command = populated.match(/Run `node "([^"]+)" "<plan.md>"`/)?.[1];
	expect(template).toBeDefined();
	expect(command).toBeDefined();
	if (template === undefined || command === undefined) throw new Error("missing documented plan or checker");
	expect(template).not.toContain("origin/main:pstack/");
	expect(template).not.toContain("Re-read the execution playbook from trunk");
	const paths = [...template.matchAll(/`([^`]+)`/g)].map((match) => match[1]).filter((path) => path.startsWith(`${plugin}/`));
	expect(paths.length).toBeGreaterThan(0);
	for (const path of paths) expect(existsSync(path)).toBe(true);
	runPlan("", 0, template, command);
});

test("accepts ordinary, indented, and longer closing fences", () => {
	for (const indent of ["", " ", "  ", "   "])
		runPlan(`${indent}\`\`\`markdown\n## Example only\n${indent}\`\`\`\n`, 0);
	runPlan("```markdown\n## Example only\n````\n", 0);
});

test("accepts nested backticks and tilde info strings", () => {
	runPlan("````markdown\n```markdown\n## Example only\n```\n````\n", 0);
	runPlan("~~~markdown `info`\n## Example only\n ~~~\n", 0);
});

test("keeps mismatched, shorter, and suffixed closers inside the fence", () => {
	runPlan("````markdown\n~~~\n## Example only\n````\n", 0);
	runPlan("````markdown\n```\n## Example only\n````\n", 0);
	runPlan("````markdown\n``` info\n## Example only\n````\n", 0);
});

test("rejects headings and prose outside a proper close", () => {
	for (const suffix of ["  ", "\t"]) {
		const heading = runPlan(`\`\`\`markdown\n## Example only\n\`\`\`${suffix}\n## Example only\n`, 1);
		expect(heading.stderr).toContain('"## Example only" after Close the program is not an appendix');
	}

	const prose = runPlan("```markdown\nThis prose has a mid-sentence colon: inside code\n```\nThis prose has a mid-sentence colon: bad.\n", 1);
	expect(prose.stderr).toContain("mid-sentence colon");
});

test("rejects backtick info strings containing a backtick", () => {
	const result = runPlan("```markdown `invalid`\n## Example only\n```\n", 1);
	expect(result.stderr).toContain('"## Example only" after Close the program is not an appendix');
});

test("ignores prose punctuation inside matching multi-backtick code spans", () => {
	for (const count of [2, 3, 4]) {
		const delimiter = "`".repeat(count);
		runPlan(`Use ${delimiter}echo \`x\`: “quoted” — example${delimiter}.\n`, 0);
	}
});

test("checks prose following a multi-backtick span and unmatched delimiters", () => {
	const outside = runPlan("Use ``echo `x`: y``. Bad: prose.\n", 1);
	expect(outside.stderr).toContain("mid-sentence colon");
	const unmatched = runPlan("Use ``unclosed: bad` here.\n", 1);
	expect(unmatched.stderr).toContain("mid-sentence colon");
});
