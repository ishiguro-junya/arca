package project

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"slices"
	"strings"
)

// pnpmのtrustPolicyで最新版が拒否される間は、検証済みの安定版を共有します。
const prismaVersion = "7.9.0"

type packageManifest struct {
	Name            string            `json:"name"`
	Private         bool              `json:"private"`
	Version         string            `json:"version"`
	Type            string            `json:"type"`
	PackageManager  string            `json:"packageManager"`
	Bin             map[string]string `json:"bin,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

func projectFiles(spec Spec, versions map[string]string) ([]File, error) {
	manifest, err := packageJSON(spec, toolVersion(versions, "pnpm"))
	if err != nil {
		return nil, err
	}
	files := []File{
		textFile("README.md", readmeTemplate(spec)),
		textFile("AGENTS.md", agentsTemplate(spec)),
		textFile(".gitignore", gitignoreTemplate(spec)),
		textFile("package.json", manifest),
		textFile("pnpm-workspace.yaml", pnpmWorkspace(nil)),
		textFile("mise.toml", miseTemplate(spec, versions)),
		textFile("lefthook.yml", lefthookTemplate()),
		textFile(".textlintrc.json", textlintTemplate()),
		textFile(".markdownlint-cli2.jsonc", markdownlintTemplate()),
	}
	if spec.Has(UseApp) || spec.HasNodeCLI() {
		files = append(files,
			textFile(".oxlintrc.json", `{"categories":{"correctness":"error","suspicious":"error"}}`),
			textFile(".oxfmtrc.json", `{"printWidth":100}`),
			textFile("knip.json", knipTemplate(spec)),
		)
	}
	if spec.Has(UseApp) {
		files = append(files, appFiles(spec)...)
	}
	if spec.Has(UseBackend) {
		files = append(files, backendFiles(spec, versions)...)
	}
	if spec.Has(UseCLI) {
		files = append(files, cliFiles(spec, versions)...)
	}
	if spec.Has(UseInfra) {
		files = append(files, infraFiles(spec)...)
	}
	slices.SortFunc(files, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return files, nil
}

func textFile(path, content string) File {
	return File{Path: path, Content: []byte(strings.TrimSpace(content) + "\n"), Mode: 0o644}
}

func toolVersion(versions map[string]string, key string) string {
	if value := versions[key]; value != "" {
		return value
	}
	return "<latest-stable>"
}

func packageJSON(spec Spec, pnpmVersion string) (string, error) {
	dependencies, devDependencies := nodeDependencies(spec)
	manifest := packageManifest{
		Name:            spec.Name,
		Private:         true,
		Version:         "0.1.0",
		Type:            "module",
		PackageManager:  "pnpm@" + pnpmVersion,
		Dependencies:    dependencies,
		DevDependencies: devDependencies,
	}
	if spec.HasNodeCLI() {
		manifest.Bin = map[string]string{spec.Name: "./dist/cli/main.js"}
	}
	data, err := json.Marshal(&manifest, json.Deterministic(true), jsontext.WithIndent("  "))
	if err != nil {
		return "", fmt.Errorf("package.jsonを生成できません: %w", err)
	}
	return string(data), nil
}

func nodeDependencies(spec Spec) (map[string]string, map[string]string) {
	dependencies := map[string]string{}
	devDependencies := map[string]string{
		"@textlint-ja/textlint-rule-preset-ai-writing": "latest",
		"markdownlint-cli2":                            "latest",
		"textlint":                                     "latest",
	}
	if spec.Has(UseApp) || spec.HasNodeCLI() {
		devDependencies["knip"] = "latest"
		devDependencies["oxfmt"] = "latest"
		devDependencies["oxlint"] = "latest"
		devDependencies["typescript"] = "latest"
	}
	if spec.Has(UseApp) {
		addAppDependencies(spec, dependencies, devDependencies)
	}
	if spec.HasNodeCLI() {
		dependencies["commander"] = "latest"
		devDependencies["@types/node"] = "latest"
		devDependencies["tsx"] = "latest"
	}
	return dependencies, devDependencies
}

func dependencySummary(spec Spec) []string {
	dependencies, devDependencies := nodeDependencies(spec)
	result := make([]string, 0, len(dependencies)+len(devDependencies)+8)
	for name := range dependencies {
		result = append(result, "npm:"+name)
	}
	for name := range devDependencies {
		result = append(result, "npm:"+name)
	}
	if spec.Has(UseBackend) {
		result = append(result, "python:fastapi", "python:httpx", "python:pytest", "python:ruff", "python:uvicorn", "python:vulture")
		if spec.Features.ORM == "sqlalchemy" {
			result = append(result, "python:alembic", "python:sqlalchemy")
		}
	}
	if spec.HasGoCLI() {
		result = append(result, "go:github.com/spf13/cobra")
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		result = append(result, "rust:tauri", "rust:tauri-build")
	}
	if spec.Has(UseInfra) {
		result = append(result, "terraform:"+spec.InfraProvider)
	}
	slices.Sort(result)
	return result
}

func addAppDependencies(spec Spec, dependencies, devDependencies map[string]string) {
	dependencies["react"] = "latest"
	dependencies["react-dom"] = "latest"
	devDependencies["@types/react"] = "latest"
	devDependencies["@types/react-dom"] = "latest"
	devDependencies["tailwindcss"] = "latest"
	if spec.IsNext() {
		dependencies["next"] = "latest"
		devDependencies["@tailwindcss/postcss"] = "latest"
	} else {
		devDependencies["@tailwindcss/vite"] = "latest"
		devDependencies["@vitejs/plugin-react"] = "latest"
		devDependencies["vite"] = "latest"
	}
	if spec.AppKind == AppDesktop {
		dependencies["@tauri-apps/api"] = "latest"
		devDependencies["@tauri-apps/cli"] = "latest"
	}
	addFeatureDependencies(spec, dependencies, devDependencies)
}

func addFeatureDependencies(spec Spec, dependencies, devDependencies map[string]string) {
	switch spec.Features.UIUX {
	case "assistant-ui":
		dependencies["@assistant-ui/react"] = "latest"
	case "radix":
		dependencies["@radix-ui/themes"] = "latest"
	case "shadcn":
		dependencies["class-variance-authority"] = "latest"
		dependencies["clsx"] = "latest"
		dependencies["tailwind-merge"] = "latest"
		devDependencies["shadcn"] = "latest"
	case "tanstack-query":
		dependencies["@tanstack/react-query"] = "latest"
	}
	if spec.Features.Localization {
		if spec.IsNext() {
			dependencies["next-intl"] = "latest"
		} else {
			dependencies["i18next"] = "latest"
			dependencies["react-i18next"] = "latest"
		}
	}
	if spec.Features.State != "" {
		dependencies[spec.Features.State] = "latest"
	}
	switch spec.Features.Form {
	case "react-hook-form":
		dependencies["react-hook-form"] = "latest"
	case "tanstack-form":
		dependencies["@tanstack/react-form"] = "latest"
	}
	if spec.Features.Validate != "" {
		dependencies[spec.Features.Validate] = "latest"
	}
	switch spec.Features.Icon {
	case "font-awesome":
		dependencies["@fortawesome/fontawesome-svg-core"] = "latest"
		dependencies["@fortawesome/free-solid-svg-icons"] = "latest"
		dependencies["@fortawesome/react-fontawesome"] = "latest"
	case "lucide":
		dependencies["lucide-react"] = "latest"
	case "mdi":
		dependencies["@mdi/js"] = "latest"
		dependencies["@mdi/react"] = "latest"
	case "react-icons":
		dependencies["react-icons"] = "latest"
	case "svgr":
		devDependencies["@svgr/core"] = "latest"
	}
	if spec.Features.FontEnabled && !spec.IsNext() {
		if slices.Contains(spec.Features.Fonts, "inter") {
			dependencies["@fontsource-variable/inter"] = "latest"
		}
		if slices.Contains(spec.Features.Fonts, "noto-sans-jp") {
			dependencies["@fontsource-variable/noto-sans-jp"] = "latest"
		}
	}
	switch spec.Features.ORM {
	case "drizzle":
		dependencies["drizzle-orm"] = "latest"
	case "prisma":
		dependencies["@prisma/client"] = prismaVersion
		devDependencies["prisma"] = prismaVersion
	}
}

func pnpmWorkspace(allowed []string) string {
	var builder strings.Builder
	builder.WriteString(`packages:
  - "."

minimumReleaseAge: 1440
minimumReleaseAgeStrict: true
trustPolicy: no-downgrade
strictDepBuilds: true
dangerouslyAllowAllBuilds: false
blockExoticSubdeps: true
`)
	if len(allowed) > 0 {
		builder.WriteString("allowBuilds:\n")
		for _, dependency := range allowed {
			fmt.Fprintf(&builder, "  %q: true\n", dependency)
		}
	}
	return builder.String()
}

func readmeTemplate(spec Spec) string {
	selections := readmeSelections(spec)
	tasks := availableTasks(spec)
	if spec.ReadmeLanguage == ReadmeEnglish {
		return fmt.Sprintf(`# %s

Generated by Arca.

## Selected stack

- %s

## mise tasks

- %s

Run `+"`mise run <task>`"+` to execute a task.
`, spec.Name, strings.Join(selections, "\n- "), strings.Join(tasks, "\n- "))
	}
	return fmt.Sprintf(`# %s

Arcaで生成したプロジェクトです。

## 選択内容

- %s

## miseタスク

- %s

`+"`mise run <task>`"+`でタスクを実行します。
`, spec.Name, strings.Join(selections, "\n- "), strings.Join(tasks, "\n- "))
}

func readmeSelections(spec Spec) []string {
	uses := make([]string, len(spec.UseCases))
	for i, useCase := range spec.UseCases {
		uses[i] = map[UseCase]string{UseApp: "App", UseBackend: "Backend", UseCLI: "CLI", UseInfra: "Infra"}[useCase]
	}
	slices.Sort(uses)
	english := spec.ReadmeLanguage == ReadmeEnglish
	label := func(ja, en string) string {
		if english {
			return en
		}
		return ja
	}
	result := []string{label("用途", "Uses") + ": " + strings.Join(uses, ", ")}
	if spec.Has(UseApp) {
		app := "React + Vite"
		if spec.IsNext() {
			app = "Next.js"
		} else if spec.AppKind == AppDesktop {
			app = "Tauri + React + Vite"
		}
		result = append(result, "App: "+app)
		if spec.AppKind == AppDesktop {
			platforms := make([]string, len(spec.DesktopPlatforms))
			for i, platform := range spec.DesktopPlatforms {
				platforms[i] = map[string]string{"macos": "macOS", "linux": "Linux", "windows": "Windows", "ios": "iOS", "android": "Android"}[platform]
			}
			result = append(result, label("対象プラットフォーム", "Platforms")+": "+strings.Join(platforms, ", "))
		}
	}
	if spec.Has(UseCLI) {
		language := "Go"
		if spec.HasNodeCLI() {
			language = "Node.js"
		}
		result = append(result, "CLI: "+language)
	}
	if spec.Has(UseInfra) {
		provider := map[string]string{"aws": "AWS", "azure": "Azure", "gcp": "GCP", "cloudflare": "Cloudflare", "vercel": "Vercel"}[spec.InfraProvider]
		result = append(result, "Infra: Terraform + "+provider)
	}
	features := enabledFeatures(spec)
	if len(features) > 0 {
		result = append(result, label("オプション", "Options")+": "+strings.Join(features, ", "))
	}
	return result
}

func enabledFeatures(spec Spec) []string {
	features := []string{}
	for _, feature := range []string{spec.Features.UIUX, spec.Features.State, spec.Features.Form, spec.Features.Validate, spec.Features.Icon, spec.Features.ORM} {
		if feature != "" {
			features = append(features, map[string]string{
				"assistant-ui":    "assistant-ui",
				"radix":           "Radix UI",
				"shadcn":          "shadcn/ui",
				"tanstack-query":  "TanStack Query",
				"jotai":           "Jotai",
				"zustand":         "Zustand",
				"react-hook-form": "React Hook Form",
				"tanstack-form":   "TanStack Form",
				"valibot":         "Valibot",
				"zod":             "Zod",
				"font-awesome":    "Font Awesome",
				"lucide":          "Lucide",
				"mdi":             "Material Design Icons",
				"react-icons":     "React Icons",
				"svgr":            "SVGR",
				"drizzle":         "Drizzle",
				"prisma":          "Prisma",
				"sqlalchemy":      "SQLAlchemy + Alembic",
				"sqlc":            "sqlc",
			}[feature])
		}
	}
	if spec.Features.Localization {
		features = append(features, "Localization")
	}
	if spec.Features.FontEnabled {
		for _, font := range spec.Features.Fonts {
			features = append(features, map[string]string{"inter": "Inter", "noto-sans-jp": "Noto Sans JP"}[font])
		}
	}
	slices.Sort(features)
	return features
}

func availableTasks(spec Spec) []string {
	tasks := []string{"setup", "fmt", "fmt-check", "lint", "typecheck", "test", "security", "check", "build", "commit-check"}
	if spec.Has(UseApp) {
		tasks = append(tasks, "app:dev", "app:build")
		if spec.AppKind == AppDesktop {
			tasks = append(tasks, "app:dev-web", "app:build-web")
		}
	}
	if spec.Has(UseBackend) {
		tasks = append(tasks, "backend:dev", "backend:test")
	}
	if spec.Has(UseCLI) {
		tasks = append(tasks, "cli:dev", "cli:build")
	}
	if spec.Has(UseInfra) {
		tasks = append(tasks, "infra:validate")
	}
	return tasks
}

func agentsTemplate(spec Spec) string {
	if spec.ReadmeLanguage == ReadmeEnglish {
		return `# AI Agent Guidelines

- Use the tools and tasks defined in mise.toml.
- Do not add package.json scripts.
- Keep dependencies locked by their package manager.
- Run mise run check before completing changes.`
	}
	return `# AIエージェント向けガイドライン

- mise.tomlで定義したツールとタスクを使用してください。
- package.jsonへscriptsを追加しないでください。
- 依存関係は各パッケージ管理ツールで固定してください。
- 変更完了前にmise run checkを実行してください。`
}

func gitignoreTemplate(spec Spec) string {
	lines := []string{".DS_Store", ".env", ".mise.local.toml", "node_modules/", "dist/", ".cache/"}
	if spec.Has(UseBackend) {
		lines = append(lines, "backend/.venv/", "backend/.pytest_cache/", "backend/__pycache__/")
	}
	if spec.IsNext() {
		lines = append(lines, ".next/")
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		lines = append(lines, "src-tauri/target/")
	}
	if spec.Has(UseInfra) {
		lines = append(lines, "infra/.terraform/", "infra/*.tfstate", "infra/*.tfstate.*")
	}
	slices.Sort(lines)
	return strings.Join(lines, "\n")
}

func textlintTemplate() string {
	return `{
  "rules": {
    "@textlint-ja/preset-ai-writing": true
  }
}`
}

func markdownlintTemplate() string {
	return `{
  "config": {
    "default": true,
    "MD013": false
  },
  "globs": ["**/*.md"],
  "ignores": ["node_modules", "dist", ".cache"]
}`
}

func lefthookTemplate() string {
	return `commit-msg:
  commands:
    conventional:
      run: mise run commit-check -- {1}

pre-commit:
  commands:
    fmt:
      run: mise run fmt-check
    lint:
      run: mise run lint

pre-push:
  commands:
    test:
      run: mise run test`
}
