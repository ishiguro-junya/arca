package project

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"charm.land/huh/v2"
)

func TestDefaultSpec(t *testing.T) {
	spec := DefaultSpec()
	if !slices.Equal(spec.UseCases, []UseCase{UseApp}) {
		t.Fatalf("既定用途がAppではありません: %v", spec.UseCases)
	}
	if spec.AppKind != AppWeb || spec.WebFramework != WebVite || spec.CLILanguage != CLIGo || spec.InfraProvider != "azure" {
		t.Fatalf("既定値が仕様と異なります: %+v", spec)
	}
	if spec.ReadmeLanguage != ReadmeJapanese {
		t.Fatalf("READMEの既定言語が日本語ではありません: %s", spec.ReadmeLanguage)
	}
}

func TestSpecValidation(t *testing.T) {
	tests := []struct {
		name string
		spec Spec
		want string
	}{
		{name: "用途なし", spec: Spec{Name: "demo", Directory: "demo", ReadmeLanguage: ReadmeJapanese}, want: "用途"},
		{name: "不正な名前", spec: Spec{Name: "Demo", Directory: "demo", ReadmeLanguage: ReadmeJapanese, UseCases: []UseCase{UseBackend}}, want: "プロジェクト名"},
		{name: "フォントなし", spec: Spec{Name: "demo", Directory: "demo", ReadmeLanguage: ReadmeJapanese, UseCases: []UseCase{UseApp}, AppKind: AppWeb, WebFramework: WebVite, Features: Features{FontEnabled: true}}, want: "フォント"},
		{name: "互換性のないORM", spec: Spec{Name: "demo", Directory: "demo", ReadmeLanguage: ReadmeJapanese, UseCases: []UseCase{UseBackend}, Features: Features{ORM: "prisma"}}, want: "Next.js"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.spec.Validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("期待したエラーではありません: %v", err)
			}
		})
	}
}

func TestFeatureAndORMOptions(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.UseCases = []UseCase{UseApp, UseBackend, UseCLI}
	spec.WebFramework = WebNext
	spec.CLILanguage = CLIGo
	categories := optionValues(featureCategoryOptions(spec))
	for _, expected := range []string{"uiux", "localization", "state", "form", "validate", "icon", "font", "orm"} {
		if !slices.Contains(categories, expected) {
			t.Errorf("カテゴリがありません: %s", expected)
		}
	}
	orms := optionValues(ormOptions(spec))
	if !slices.Equal(orms, []string{"drizzle", "prisma", "sqlalchemy", "sqlc"}) {
		t.Fatalf("ORM候補が不正です: %v", orms)
	}
}

func TestProjectGoldens(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*Spec)
	}{
		{name: "app", configure: func(spec *Spec) { spec.UseCases = []UseCase{UseApp} }},
		{name: "desktop", configure: func(spec *Spec) {
			spec.UseCases = []UseCase{UseApp}
			spec.AppKind = AppDesktop
			spec.DesktopPlatforms = []string{"macos", "ios", "android"}
		}},
		{name: "backend", configure: func(spec *Spec) { spec.UseCases = []UseCase{UseBackend} }},
		{name: "go_cli", configure: func(spec *Spec) { spec.UseCases = []UseCase{UseCLI}; spec.CLILanguage = CLIGo }},
		{name: "node_cli", configure: func(spec *Spec) { spec.UseCases = []UseCase{UseCLI}; spec.CLILanguage = CLINode }},
		{name: "infra", configure: func(spec *Spec) { spec.UseCases = []UseCase{UseInfra} }},
		{name: "composite", configure: func(spec *Spec) {
			spec.UseCases = []UseCase{UseApp, UseBackend, UseCLI, UseInfra}
			spec.WebFramework = WebNext
			spec.CLILanguage = CLINode
			spec.InfraProvider = "cloudflare"
			spec.Features = Features{
				UIUX:         "shadcn",
				Localization: true,
				State:        "jotai",
				Form:         "react-hook-form",
				Validate:     "zod",
				Icon:         "lucide",
				FontEnabled:  true,
				Fonts:        []string{"inter", "noto-sans-jp"},
				ORM:          "sqlalchemy",
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validSpec("generated")
			test.configure(&spec)
			plan, err := BuildPlan(spec, versionsFor(spec))
			if err != nil {
				t.Fatal(err)
			}
			if hasFile(plan, ".github/workflows/ci.yml") || hasFile(plan, "mise.lock") {
				t.Fatal("対象外のCI設定またはmise.lockが生成されています")
			}
			assertGolden(t, test.name, plan.Files)
		})
	}
}

func assertGolden(t *testing.T, name string, files []File) {
	t.Helper()
	var snapshot strings.Builder
	for _, file := range files {
		fmt.Fprintf(&snapshot, "=== %s %04o ===\n%s\n", file.Path, file.Mode.Perm(), file.Content)
	}
	path := filepath.Join("testdata", name+".golden")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(snapshot.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != snapshot.String() {
		t.Fatalf("生成結果が%sと一致しません", path)
	}
}

func TestDesktopAllowsMultiplePlatforms(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"macos", "linux", "windows", "ios", "android"}
	if err := spec.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBuildPlanForViteApp(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.Features = Features{
		Localization: true,
		FontEnabled:  true,
		Fonts:        []string{"inter", "noto-sans-jp"},
	}
	plan, err := BuildPlan(spec, versionsFor(spec))
	if err != nil {
		t.Fatal(err)
	}
	manifest := fileContent(t, plan, "package.json")
	for _, dependency := range []string{"@fontsource-variable/inter", "@fontsource-variable/noto-sans-jp", "i18next", "react-i18next", "oxlint", "oxfmt"} {
		if !strings.Contains(manifest, dependency) {
			t.Errorf("package.jsonに%sがありません", dependency)
		}
	}
	if strings.Contains(manifest, `"scripts"`) || strings.Contains(manifest, "biome") || strings.Contains(manifest, "eslint") || strings.Contains(manifest, "prettier") {
		t.Fatalf("不要なscriptsまたはツールが含まれています:\n%s", manifest)
	}
	css := fileContent(t, plan, "src/index.css")
	if !strings.Contains(css, `"Inter Variable", "Noto Sans JP Variable", sans-serif`) {
		t.Fatalf("フォント順が不正です:\n%s", css)
	}
	mise := fileContent(t, plan, "mise.toml")
	if !strings.Contains(mise, "lockfile = false") || strings.Contains(mise, "<latest-stable>") {
		t.Fatalf("mise.tomlに完全な版またはlockfile無効化がありません:\n%s", mise)
	}
	if hasFile(plan, "mise.lock") {
		t.Fatal("mise.lockが生成計画に含まれています")
	}
	workspace := fileContent(t, plan, "pnpm-workspace.yaml")
	if !strings.Contains(workspace, "packages:\n  - \".\"") {
		t.Fatalf("pnpmワークスペースがルートだけを対象にしていません:\n%s", workspace)
	}
	readme := fileContent(t, plan, "README.md")
	if !strings.Contains(readme, "React + Vite") || !strings.Contains(readme, "app:build") {
		t.Fatalf("READMEに選択内容またはmiseタスクがありません:\n%s", readme)
	}
}

func TestBuildPlanForNextFonts(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.WebFramework = WebNext
	spec.Features = Features{FontEnabled: true, Fonts: []string{"inter", "noto-sans-jp"}}
	plan, err := BuildPlan(spec, versionsFor(spec))
	if err != nil {
		t.Fatal(err)
	}
	layout := fileContent(t, plan, "src/app/layout.tsx")
	if !strings.Contains(layout, `from "next/font/google"`) || !strings.Contains(layout, "Noto_Sans_JP({ preload: false") {
		t.Fatalf("next/font設定が不正です:\n%s", layout)
	}
	manifest := fileContent(t, plan, "package.json")
	if strings.Contains(manifest, "@fontsource") {
		t.Fatalf("Next.jsへFontsourceが追加されています:\n%s", manifest)
	}
	tsconfig := fileContent(t, plan, "tsconfig.json")
	if !strings.Contains(tsconfig, `"jsx": "preserve"`) || !strings.Contains(tsconfig, `".next/types/**/*.ts"`) || !strings.Contains(tsconfig, `"plugins": [{ "name": "next" }]`) {
		t.Fatalf("Next.jsのTypeScript設定が不足しています:\n%s", tsconfig)
	}
}

func TestEnglishDocumentationAndTasks(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.ReadmeLanguage = ReadmeEnglish
	plan, err := BuildPlan(spec, versionsFor(spec))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fileContent(t, plan, "README.md"), "Selected stack") || !strings.Contains(fileContent(t, plan, "AGENTS.md"), "AI Agent Guidelines") {
		t.Fatal("English documentation was not generated")
	}
	mise := fileContent(t, plan, "mise.toml")
	if !strings.Contains(mise, `description = "Install dependencies"`) || strings.Contains(mise, "対象はありません") {
		t.Fatalf("mise task descriptions are not in English:\n%s", mise)
	}
}

func TestBuildPlanUsesBackendAndCLIDirectories(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.UseCases = []UseCase{UseBackend, UseCLI, UseInfra}
	spec.CLILanguage = CLINode
	spec.InfraProvider = "azure"
	plan, err := BuildPlan(spec, versionsFor(spec))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"backend/pyproject.toml", "cli/main.ts", "infra/main.tf"} {
		if !hasFile(plan, path) {
			t.Errorf("生成ファイルがありません: %s", path)
		}
	}
	if !strings.Contains(fileContent(t, plan, "infra/main.tf"), "features {}") {
		t.Fatal("Azureプロバイダーの必須features設定がありません")
	}
	for _, file := range plan.Files {
		if strings.HasPrefix(file.Path, "cmd/") {
			t.Fatalf("cmd/配下が生成されています: %s", file.Path)
		}
	}
	mise := fileContent(t, plan, "mise.toml")
	if !strings.Contains(mise, `[tasks."backend:dev"]`) || !strings.Contains(mise, `[tasks."cli:dev"]`) || !strings.Contains(mise, `[tasks."cli:build"]`) || !strings.Contains(mise, `[tasks."infra:validate"]`) {
		t.Fatalf("用途別miseタスクが不足しています:\n%s", mise)
	}
	for _, path := range []string{".oxlintrc.json", ".oxfmtrc.json", "knip.json"} {
		if !hasFile(plan, path) {
			t.Errorf("Node.js CLIの品質設定がありません: %s", path)
		}
	}
	if !strings.Contains(fileContent(t, plan, "knip.json"), `"cli/main.ts"`) {
		t.Fatal("KnipのNode.js CLIエントリがありません")
	}
	if len(plan.Dependencies) == 0 || !slices.Contains(plan.Dependencies, "npm:commander") {
		t.Fatalf("確認用の依存関係が不足しています: %v", plan.Dependencies)
	}
	lefthook := fileContent(t, plan, "lefthook.yml")
	if !strings.Contains(lefthook, "mise run commit-check -- {1}") || strings.Contains(lefthook, "cog verify") {
		t.Fatalf("Lefthookがmiseタスクだけを呼び出していません:\n%s", lefthook)
	}
}

func TestDesktopBuildUsesMiseFrontendTasks(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"macos"}
	plan, err := BuildPlan(spec, versionsFor(spec))
	if err != nil {
		t.Fatal(err)
	}
	config := fileContent(t, plan, "src-tauri/tauri.conf.json")
	for _, expected := range []string{`"beforeDevCommand": "mise run app:dev-web"`, `"beforeBuildCommand": "mise run app:build-web"`} {
		if !strings.Contains(config, expected) {
			t.Errorf("Tauriのフロントエンド連携がありません: %s", expected)
		}
	}
	if !strings.Contains(fileContent(t, plan, "mise.toml"), `[tasks."app:build"]`) {
		t.Fatal("Appのビルドタスクがありません")
	}
}

func TestViteAndNodeCLIUseSeparateOutputDirectories(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.UseCases = []UseCase{UseApp, UseCLI}
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"macos"}
	spec.CLILanguage = CLINode
	plan, err := BuildPlan(spec, versionsFor(spec))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fileContent(t, plan, "vite.config.ts"), `outDir: "dist/app"`) {
		t.Fatal("Viteの出力先がApp専用ディレクトリではありません")
	}
	if !strings.Contains(fileContent(t, plan, "tsconfig.cli.json"), `"outDir": "dist/cli"`) {
		t.Fatal("Node.js CLIの出力先がCLI専用ディレクトリではありません")
	}
	if !strings.Contains(fileContent(t, plan, "src-tauri/tauri.conf.json"), `"frontendDist": "../dist/app"`) {
		t.Fatal("TauriがApp専用のフロントエンド出力を参照していません")
	}
}

func TestDesktopGitignoreKeepsMobileProjects(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"ios", "android"}
	plan, err := BuildPlan(spec, versionsFor(spec))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(fileContent(t, plan, ".gitignore"), "src-tauri/gen/") {
		t.Fatal("Tauriのモバイルプロジェクト全体がGit管理から除外されています")
	}
}

func TestDesktopMiseInstallsMobileRustTargets(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"ios", "android"}
	plan, err := BuildPlan(spec, versionsFor(spec))
	if err != nil {
		t.Fatal(err)
	}
	mise := fileContent(t, plan, "mise.toml")
	for _, expected := range []string{
		`components = ["clippy", "rustfmt"]`,
		`"aarch64-apple-ios"`,
		`"x86_64-apple-ios"`,
		`"aarch64-apple-ios-sim"`,
		`"aarch64-linux-android"`,
		`"armv7-linux-androideabi"`,
		`"i686-linux-android"`,
		`"x86_64-linux-android"`,
	} {
		if !strings.Contains(mise, expected) {
			t.Errorf("mise.tomlにRust設定がありません: %s", expected)
		}
	}
}

func TestPreviewDoesNotCreateTarget(t *testing.T) {
	target := filepath.Join(t.TempDir(), "preview")
	spec := validSpec(target)
	var out bytes.Buffer
	if err := Preview(spec, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("プレビューで生成先が作成されました: %v", err)
	}
	if !strings.Contains(out.String(), "package.json") || !strings.Contains(out.String(), "依存関係:") || !strings.Contains(out.String(), "npm:react") || !strings.Contains(out.String(), "mise install") {
		t.Fatalf("プレビュー内容が不足しています:\n%s", out.String())
	}
	if strings.Contains(out.String(), "arca:") || !strings.Contains(out.String(), "mise exec -- pnpm ignored-builds") {
		t.Fatalf("プレビューの実行予定コマンドが不正です:\n%s", out.String())
	}
}

func TestPrismaBuildAllowlistUsesDependencyVersion(t *testing.T) {
	spec := validSpec(t.TempDir())
	spec.WebFramework = WebNext
	spec.Features.ORM = "prisma"
	manifest, err := packageJSON(spec, "1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	for _, dependency := range []string{"prisma", "@prisma/client"} {
		if !strings.Contains(manifest, fmt.Sprintf("%q: %q", dependency, prismaVersion)) {
			t.Fatalf("%sが共通のPrisma版を使用していません:\n%s", dependency, manifest)
		}
	}
	for _, dependency := range []string{"prisma@" + prismaVersion, "@prisma/engines@" + prismaVersion} {
		if !reviewedBuildDependencies[dependency] {
			t.Fatalf("Prismaのビルド依存関係が許可一覧にありません: %s", dependency)
		}
	}
}

func TestBuildDependenciesFromLock(t *testing.T) {
	lock := []byte(`lockfileVersion: '9.0'
packages:
  esbuild@0.28.2:
    resolution: {integrity: example}
  react@19.0.0:
    resolution: {integrity: example}
`)
	ignored := ignoredBuildNames(`Automatically ignored builds during installation:
  esbuild
hint: add it to allowBuilds`)
	dependencies, err := buildDependenciesFromLock(lock, ignored)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(dependencies, []string{"esbuild@0.28.2"}) {
		t.Fatalf("ビルド依存関係が不正です: %v", dependencies)
	}
	workspace := pnpmWorkspace(dependencies)
	if !strings.Contains(workspace, `"esbuild@0.28.2": true`) {
		t.Fatalf("allowBuildsが完全版指定ではありません:\n%s", workspace)
	}
}

func TestIgnoredBuildNamesAllowsNoNodeModules(t *testing.T) {
	names := ignoredBuildNames("Automatically ignored builds during installation:\n  Cannot identify as no node_modules found\n")
	if len(names) != 0 {
		t.Fatalf("ビルド依存関係なしの出力を誤認しました: %v", names)
	}
}

func TestUnknownBuildDependencyFails(t *testing.T) {
	lock := []byte("packages:\n  esbuild@99.0.0:\n")
	_, err := buildDependenciesFromLock(lock, []string{"esbuild"})
	if err == nil || !strings.Contains(err.Error(), "未確認") {
		t.Fatalf("未確認のビルド処理が拒否されませんでした: %v", err)
	}
}

func validSpec(directory string) Spec {
	spec := DefaultSpec()
	spec.Name = "demo"
	spec.Directory = directory
	return spec
}

func versionsFor(spec Spec) map[string]string {
	versions := map[string]string{}
	for _, tool := range requiredTools(spec) {
		versions[tool.Key] = "1.2.3"
	}
	return versions
}

func optionValues(options []huh.Option[string]) []string {
	values := make([]string, len(options))
	for i, option := range options {
		values[i] = option.Value
	}
	return values
}

func fileContent(t *testing.T, plan Plan, path string) string {
	t.Helper()
	for _, file := range plan.Files {
		if file.Path == path {
			return string(file.Content)
		}
	}
	t.Fatalf("ファイルがありません: %s", path)
	return ""
}

func hasFile(plan Plan, path string) bool {
	return slices.ContainsFunc(plan.Files, func(file File) bool { return file.Path == path })
}
