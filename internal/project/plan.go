package project

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

type Tool struct {
	Key   string
	Query string
}

type File struct {
	Path    string
	Content []byte
	Mode    fs.FileMode
}

type Command struct {
	Directory string
	Args      []string
}

type Plan struct {
	Files        []File
	Commands     []Command
	Tools        []Tool
	Dependencies []string
}

var ignoredBuildsCommand = []string{"mise", "exec", "--", "pnpm", "ignored-builds"}

func BuildPlan(spec Spec, versions map[string]string) (Plan, error) {
	if err := spec.Validate(); err != nil {
		return Plan{}, err
	}
	files, err := projectFiles(spec, versions)
	if err != nil {
		return Plan{}, err
	}
	return Plan{
		Files:        files,
		Commands:     projectCommands(spec),
		Tools:        requiredTools(spec),
		Dependencies: dependencySummary(spec),
	}, nil
}

func Preview(spec Spec, out io.Writer) error {
	plan, err := BuildPlan(spec, nil)
	if err != nil {
		return err
	}
	dir, err := spec.TargetDirectory()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "生成先: %s\n\nファイル:\n", dir); err != nil {
		return err
	}
	for _, file := range plan.Files {
		if _, err := fmt.Fprintf(out, "  %s\n", file.Path); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, "\nツール:"); err != nil {
		return err
	}
	for _, tool := range plan.Tools {
		if _, err := fmt.Fprintf(out, "  mise latest %s\n", tool.Query); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, "\n依存関係:"); err != nil {
		return err
	}
	for _, dependency := range plan.Dependencies {
		if _, err := fmt.Fprintf(out, "  %s\n", dependency); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, "\nコマンド:"); err != nil {
		return err
	}
	for _, command := range plan.Commands {
		line := strings.Join(command.Args, " ")
		if command.Directory != "." {
			line = fmt.Sprintf("(cd %s && %s)", command.Directory, line)
		}
		if _, err := fmt.Fprintf(out, "  %s\n", line); err != nil {
			return err
		}
	}
	return nil
}

func requiredTools(spec Spec) []Tool {
	tools := []Tool{
		{Key: "node", Query: "node"},
		{Key: "pnpm", Query: "pnpm"},
		{Key: "lefthook", Query: "lefthook"},
		{Key: "cocogitto", Query: "cocogitto"},
		{Key: "lychee", Query: "lychee"},
		{Key: "osv-scanner", Query: "osv-scanner"},
		{Key: "trufflehog", Query: "trufflehog"},
		{Key: "semgrep", Query: "semgrep"},
	}
	if spec.Has(UseBackend) {
		tools = append(tools, Tool{Key: "python", Query: "python"}, Tool{Key: "uv", Query: "uv"})
	}
	if spec.HasGoCLI() {
		tools = append(tools, Tool{Key: "go", Query: "go"})
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		tools = append(tools, Tool{Key: "rust", Query: "rust"})
		if slices.Contains(spec.DesktopPlatforms, "android") {
			tools = append(tools, Tool{Key: "java", Query: "java"})
		}
	}
	if spec.Has(UseInfra) {
		tools = append(tools,
			Tool{Key: "terraform", Query: "terraform"},
			Tool{Key: "tflint", Query: "tflint"},
			Tool{Key: "trivy", Query: "trivy"},
		)
	}
	if spec.Features.ORM == "sqlc" {
		tools = append(tools, Tool{Key: "sqlc", Query: "sqlc"})
	}
	slices.SortFunc(tools, func(a, b Tool) int { return strings.Compare(a.Key, b.Key) })
	return slices.CompactFunc(tools, func(a, b Tool) bool { return a.Key == b.Key })
}

func projectCommands(spec Spec) []Command {
	commands := []Command{
		{Directory: ".", Args: []string{"mise", "install"}},
		{Directory: ".", Args: []string{"mise", "exec", "--", "pnpm", "install", "--config.strict-dep-builds=false"}},
		{Directory: ".", Args: slices.Clone(ignoredBuildsCommand)},
		{Directory: ".", Args: []string{"mise", "exec", "--", "pnpm", "rebuild"}},
	}
	if spec.Has(UseBackend) {
		commands = append(commands, Command{Directory: ".", Args: []string{"mise", "exec", "--", "uv", "lock", "--project", "backend"}})
	}
	if spec.HasGoCLI() {
		commands = append(commands,
			Command{Directory: ".", Args: []string{"mise", "exec", "--", "go", "get", "github.com/spf13/cobra@latest"}},
			Command{Directory: ".", Args: []string{"mise", "exec", "--", "gofmt", "-w", "cli"}},
			Command{Directory: ".", Args: []string{"mise", "exec", "--", "go", "mod", "tidy"}},
		)
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		commands = append(commands, Command{Directory: ".", Args: []string{"mise", "exec", "--", "cargo", "generate-lockfile", "--manifest-path", "src-tauri/Cargo.toml"}})
		if slices.Contains(spec.DesktopPlatforms, "ios") {
			commands = append(commands, Command{Directory: ".", Args: []string{"mise", "exec", "--", "pnpm", "exec", "tauri", "ios", "init"}})
		}
		if slices.Contains(spec.DesktopPlatforms, "android") {
			commands = append(commands, Command{Directory: ".", Args: []string{"mise", "exec", "--", "pnpm", "exec", "tauri", "android", "init"}})
		}
	}
	if spec.Has(UseInfra) {
		commands = append(commands, Command{Directory: ".", Args: []string{"mise", "exec", "--", "terraform", "-chdir=infra", "init", "-backend=false"}})
	}
	return commands
}

func pathFor(root, relative string) string {
	if relative == "." {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(relative))
}
