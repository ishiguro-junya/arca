package project

import (
	"fmt"
	"slices"
	"strings"
)

func miseTemplate(spec Spec, versions map[string]string) string {
	var builder strings.Builder
	builder.WriteString("[settings]\nlockfile = false\n\n[tools]\n")
	for _, tool := range requiredTools(spec) {
		if tool.Key == "rust" {
			fmt.Fprintf(&builder, "%q = { version = %q, components = [\"clippy\", \"rustfmt\"]", tool.Key, toolVersion(versions, tool.Key))
			if targets := rustTargets(spec); len(targets) > 0 {
				quoted := make([]string, len(targets))
				for i, target := range targets {
					quoted[i] = fmt.Sprintf("%q", target)
				}
				fmt.Fprintf(&builder, ", targets = [%s]", strings.Join(quoted, ", "))
			}
			builder.WriteString(" }\n")
			continue
		}
		fmt.Fprintf(&builder, "%q = %q\n", tool.Key, toolVersion(versions, tool.Key))
	}
	writeLocalizedTask(&builder, spec, "setup", "依存関係をインストールする", "Install dependencies", setupCommands(spec))
	writeLocalizedTask(&builder, spec, "fmt", "コードを整形する", "Format code", formatCommands(spec, false))
	writeLocalizedTask(&builder, spec, "fmt-check", "コードが整形済みか確認する", "Check code formatting", formatCommands(spec, true))
	writeLocalizedTask(&builder, spec, "lint", "静的解析を実行する", "Run linters", lintCommands(spec))
	writeLocalizedTask(&builder, spec, "typecheck", "型を検査する", "Run type checks", typecheckCommands(spec))
	writeLocalizedTask(&builder, spec, "test", "テストを実行する", "Run tests", testCommands(spec))
	writeLocalizedTask(&builder, spec, "security", "セキュリティ検査を実行する", "Run security checks", securityCommands(spec))
	writeLocalizedTask(&builder, spec, "build", "成果物をビルドする", "Build artifacts", buildCommands(spec))
	writeLocalizedTask(&builder, spec, "commit-check", "コミットメッセージを検査する", "Check a commit message", []string{`cog verify --file "$1"`})
	fmt.Fprintf(&builder, "\n[tasks.check]\ndescription = %q\ndepends = [\"fmt-check\", \"lint\", \"typecheck\", \"test\", \"security\"]\n", localized(spec, "全検査を実行する", "Run all checks"))
	writeComponentTasks(&builder, spec)
	return builder.String()
}

func rustTargets(spec Spec) []string {
	targets := []string{}
	if slices.Contains(spec.DesktopPlatforms, "ios") {
		targets = append(targets, "aarch64-apple-ios", "x86_64-apple-ios", "aarch64-apple-ios-sim")
	}
	if slices.Contains(spec.DesktopPlatforms, "android") {
		targets = append(targets, "aarch64-linux-android", "armv7-linux-androideabi", "i686-linux-android", "x86_64-linux-android")
	}
	slices.Sort(targets)
	return targets
}

func writeLocalizedTask(builder *strings.Builder, spec Spec, name, japanese, english string, commands []string) {
	if len(commands) == 0 {
		commands = []string{fmt.Sprintf("echo %q", localized(spec, name+"の対象はありません", "No target for "+name))}
	}
	writeTask(builder, name, localized(spec, japanese, english), commands)
}

func localized(spec Spec, japanese, english string) string {
	if spec.ReadmeLanguage == ReadmeEnglish {
		return english
	}
	return japanese
}

func writeTask(builder *strings.Builder, name, description string, commands []string) {
	fmt.Fprintf(builder, "\n[tasks.%q]\ndescription = %q\nrun = '''\nset -e\n%s\n'''\n", name, description, strings.Join(commands, "\n"))
}

func setupCommands(spec Spec) []string {
	commands := []string{"pnpm install --frozen-lockfile"}
	if spec.Has(UseBackend) {
		commands = append(commands, "uv sync --project backend --frozen")
	}
	if spec.HasGoCLI() {
		commands = append(commands, "go mod download")
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		commands = append(commands, "cargo fetch --locked --manifest-path src-tauri/Cargo.toml")
	}
	if spec.Has(UseInfra) {
		commands = append(commands, "terraform -chdir=infra init -backend=false")
	}
	commands = append(commands, "if [ -e .git ]; then lefthook install; fi")
	return commands
}

func formatCommands(spec Spec, check bool) []string {
	var commands []string
	if spec.Has(UseApp) || spec.HasNodeCLI() {
		if check {
			commands = append(commands, "pnpm exec oxfmt --check .")
		} else {
			commands = append(commands, "pnpm exec oxfmt --write .")
		}
	}
	if spec.Has(UseBackend) {
		if check {
			commands = append(commands, "uv run --project backend ruff format --check backend")
		} else {
			commands = append(commands, "uv run --project backend ruff format backend")
		}
	}
	if spec.HasGoCLI() {
		if check {
			commands = append(commands, `test -z "$(gofmt -l cli)"`)
		} else {
			commands = append(commands, "gofmt -w cli")
		}
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		command := "cargo fmt --manifest-path src-tauri/Cargo.toml"
		if check {
			command += " -- --check"
		}
		commands = append(commands, command)
	}
	if spec.Has(UseInfra) {
		command := "terraform -chdir=infra fmt"
		if check {
			command += " -check"
		}
		commands = append(commands, command)
	}
	return commands
}

func lintCommands(spec Spec) []string {
	commands := []string{`pnpm exec textlint "**/*.md"`, "pnpm exec markdownlint-cli2", "lychee --exclude-path node_modules --exclude-path dist --exclude-path .cache ."}
	if spec.Has(UseApp) || spec.HasNodeCLI() {
		commands = append(commands, "pnpm exec oxlint .", "pnpm exec knip")
	}
	if spec.Has(UseBackend) {
		commands = append(commands, "uv run --project backend ruff check backend", "uv run --project backend vulture backend")
	}
	if spec.HasGoCLI() {
		commands = append(commands, "go vet ./...")
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		commands = append(commands, "cargo clippy --manifest-path src-tauri/Cargo.toml -- -D warnings")
	}
	if spec.Has(UseInfra) {
		commands = append(commands, "terraform -chdir=infra validate", "tflint --chdir=infra")
	}
	return commands
}

func typecheckCommands(spec Spec) []string {
	var commands []string
	if spec.Has(UseApp) {
		commands = append(commands, "pnpm exec tsc --noEmit -p tsconfig.json")
	}
	if spec.HasNodeCLI() {
		commands = append(commands, "pnpm exec tsc --noEmit -p tsconfig.cli.json")
	}
	if spec.Has(UseBackend) {
		commands = append(commands, "uv run --project backend python -m compileall -q backend/src")
	}
	return commands
}

func testCommands(spec Spec) []string {
	var commands []string
	if spec.Has(UseBackend) {
		commands = append(commands, "uv run --project backend pytest backend")
	}
	if spec.HasGoCLI() {
		commands = append(commands, "go test ./...")
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		commands = append(commands, "cargo test --manifest-path src-tauri/Cargo.toml")
	}
	return commands
}

func securityCommands(spec Spec) []string {
	commands := []string{
		"osv-scanner scan source -r .",
		"trufflehog filesystem . --fail --no-update",
		"semgrep scan --config auto .",
	}
	if spec.Has(UseInfra) {
		commands = append(commands, "trivy config infra")
	}
	return commands
}

func buildCommands(spec Spec) []string {
	var commands []string
	if spec.Has(UseApp) {
		switch {
		case spec.AppKind == AppDesktop:
			commands = append(commands, "pnpm exec tauri build")
		case spec.IsNext():
			commands = append(commands, "pnpm exec next build")
		default:
			commands = append(commands, "pnpm exec vite build")
		}
	}
	if spec.HasGoCLI() {
		commands = append(commands, fmt.Sprintf("go build -o dist/%s ./cli", spec.Name))
	}
	if spec.HasNodeCLI() {
		commands = append(commands, "pnpm exec tsc -p tsconfig.cli.json")
	}
	return commands
}

func writeComponentTasks(builder *strings.Builder, spec Spec) {
	if spec.Has(UseApp) {
		dev := "pnpm exec vite"
		build := "pnpm exec vite build"
		if spec.IsNext() {
			dev = "pnpm exec next dev"
			build = "pnpm exec next build"
		}
		if spec.AppKind == AppDesktop {
			dev = "pnpm exec tauri dev"
			build = "pnpm exec tauri build"
			writeLocalizedTask(builder, spec, "app:dev-web", "Desktopのフロントエンドを起動する", "Start the desktop frontend", []string{"pnpm exec vite --port 1420"})
			writeLocalizedTask(builder, spec, "app:build-web", "Desktopのフロントエンドをビルドする", "Build the desktop frontend", []string{"pnpm exec vite build"})
		}
		writeLocalizedTask(builder, spec, "app:dev", "Appを起動する", "Start the app", []string{dev})
		writeLocalizedTask(builder, spec, "app:build", "Appをビルドする", "Build the app", []string{build})
	}
	if spec.Has(UseBackend) {
		module := strings.ReplaceAll(spec.Name, "-", "_")
		writeLocalizedTask(builder, spec, "backend:dev", "Backendを起動する", "Start the backend", []string{fmt.Sprintf("uv run --project backend fastapi dev backend/src/%s/main.py", module)})
		writeLocalizedTask(builder, spec, "backend:test", "Backendをテストする", "Test the backend", []string{"uv run --project backend pytest backend"})
	}
	if spec.Has(UseCLI) {
		if spec.HasGoCLI() {
			writeLocalizedTask(builder, spec, "cli:dev", "CLIを実行する", "Run the CLI", []string{"go run ./cli"})
			writeLocalizedTask(builder, spec, "cli:build", "CLIをビルドする", "Build the CLI", []string{fmt.Sprintf("go build -o dist/%s ./cli", spec.Name)})
		} else {
			writeLocalizedTask(builder, spec, "cli:dev", "CLIを実行する", "Run the CLI", []string{"pnpm exec tsx cli/main.ts"})
			writeLocalizedTask(builder, spec, "cli:build", "CLIをビルドする", "Build the CLI", []string{"pnpm exec tsc -p tsconfig.cli.json"})
		}
	}
	if spec.Has(UseInfra) {
		writeLocalizedTask(builder, spec, "infra:validate", "Infraを検証する", "Validate infrastructure", []string{"terraform -chdir=infra validate", "tflint --chdir=infra", "trivy config infra"})
	}
}

func appFiles(spec Spec) []File {
	var files []File
	if spec.IsNext() {
		files = append(files, nextFiles(spec)...)
	} else {
		files = append(files, viteFiles(spec)...)
		if spec.AppKind == AppDesktop {
			files = append(files, tauriFiles(spec)...)
		}
	}
	if spec.Features.UIUX == "shadcn" {
		cssPath := "src/index.css"
		rsc := "false"
		if spec.IsNext() {
			cssPath = "src/app/globals.css"
			rsc = "true"
		}
		files = append(files,
			textFile("components.json", fmt.Sprintf(`{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "new-york",
  "rsc": %s,
  "tsx": true,
  "tailwind": { "css": %q, "cssVariables": true },
  "aliases": { "components": "@/components", "utils": "@/lib/utils" }
}`, rsc, cssPath)),
			textFile("src/lib/utils.ts", `import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}`),
		)
	}
	return files
}
