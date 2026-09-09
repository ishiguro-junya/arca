package project

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var reviewedBuildDependencies = map[string]bool{
	"@parcel/watcher@2.6.0":            true,
	"@prisma/engines@" + prismaVersion: true,
	"@swc/core@1.16.1":                 true,
	"esbuild@0.28.2":                   true,
	"prisma@" + prismaVersion:          true,
}

type commandRunner interface {
	Run(context.Context, string, []string, io.Writer) error
	Output(context.Context, string, []string) (string, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, directory string, args []string, out io.Writer) error {
	// #nosec G204 -- 実行するコマンドはArca内部の固定済み生成計画だけから渡されます。
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), "MISE_YES=1")
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%sの実行に失敗しました: %w", strings.Join(args, " "), err)
	}
	return nil
}

func (execRunner) Output(ctx context.Context, directory string, args []string) (string, error) {
	// #nosec G204 -- 実行するコマンドはArca内部の固定済み生成計画だけから渡されます。
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = directory
	cmd.Env = os.Environ()
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%sの実行に失敗しました: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func Generate(ctx context.Context, spec Spec, out io.Writer) error {
	return generate(ctx, spec, out, execRunner{})
}

func generate(ctx context.Context, spec Spec, out io.Writer, runner commandRunner) error {
	if err := spec.Validate(); err != nil {
		return err
	}
	target, err := spec.TargetDirectory()
	if err != nil {
		return err
	}
	if err := validateEmptyTarget(target); err != nil {
		return err
	}
	if _, err := exec.LookPath("mise"); err != nil {
		return errors.New("miseが見つかりません")
	}
	if err := validateMobileSDKs(spec); err != nil {
		return err
	}
	previewPlan, err := BuildPlan(spec, nil)
	if err != nil {
		return err
	}
	versions, err := resolveToolVersions(ctx, previewPlan.Tools, runner)
	if err != nil {
		return err
	}
	plan, err := BuildPlan(spec, versions)
	if err != nil {
		return err
	}
	if err := writeProjectFiles(target, plan.Files); err != nil {
		return err
	}
	for _, command := range plan.Commands {
		if slices.Equal(command.Args, ignoredBuildsCommand) {
			output, err := runner.Output(ctx, target, command.Args)
			if err != nil {
				return fmt.Errorf("ビルド処理が必要な依存関係を取得できません: %w", err)
			}
			if err := configureAllowedBuilds(target, output); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(out, "実行: %s\n", strings.Join(command.Args, " ")); err != nil {
			return err
		}
		directory := pathFor(target, command.Directory)
		if err := runner.Run(ctx, directory, command.Args, out); err != nil {
			return err
		}
	}
	if spec.Has(UseInfra) {
		terraformCache := filepath.Join(target, "infra", ".terraform")
		if err := os.RemoveAll(terraformCache); err != nil {
			return fmt.Errorf("作業用Terraformディレクトリを削除できません: %w", err)
		}
	}
	_, err = fmt.Fprintf(out, "%sを生成しました。\n", target)
	return err
}

func validateEmptyTarget(target string) error {
	info, err := os.Stat(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("生成先を確認できません: %w", err)
	}
	if !info.IsDir() {
		return errors.New("生成先にディレクトリ以外が存在します")
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return fmt.Errorf("生成先を読み取れません: %w", err)
	}
	if len(entries) > 0 {
		return errors.New("生成先ディレクトリが空ではありません")
	}
	return nil
}

func validateMobileSDKs(spec Spec) error {
	if !spec.Has(UseApp) || spec.AppKind != AppDesktop {
		return nil
	}
	if slices.Contains(spec.DesktopPlatforms, "ios") {
		for _, command := range []string{"pod", "xcodebuild", "xcrun"} {
			if _, err := exec.LookPath(command); err != nil {
				return fmt.Errorf("iOS向け生成には%sが必要です", command)
			}
		}
		// #nosec G204 -- コマンドと引数は固定で、iOS SDKの存在確認だけに使用します。
		if err := exec.Command("xcrun", "--sdk", "iphoneos", "--show-sdk-path").Run(); err != nil {
			return errors.New("iOS向け生成には完全なXcodeとiOS SDKが必要です")
		}
	}
	if slices.Contains(spec.DesktopPlatforms, "android") {
		androidSDK := firstNonEmpty(os.Getenv("ANDROID_HOME"), os.Getenv("ANDROID_SDK_ROOT"))
		androidNDK := firstNonEmpty(os.Getenv("NDK_HOME"), os.Getenv("ANDROID_NDK_HOME"))
		if androidSDK == "" || androidNDK == "" {
			return errors.New("Android向け生成にはAndroid SDKとNDKの環境変数が必要です")
		}
		if !directoryExists(androidSDK) || !directoryExists(androidNDK) {
			return errors.New("指定されたAndroid SDKまたはNDKのディレクトリが見つかりません")
		}
		platforms, _ := filepath.Glob(filepath.Join(androidSDK, "platforms", "android-*", "android.jar"))
		buildTools, _ := filepath.Glob(filepath.Join(androidSDK, "build-tools", "*", "aapt2"))
		if !executableFile(filepath.Join(androidSDK, "platform-tools", "adb")) ||
			!slices.ContainsFunc(platforms, regularFile) ||
			!slices.ContainsFunc(buildTools, executableFile) {
			return errors.New("必要なAndroid SDK Platform、Platform-Tools、Build-Toolsが見つかりません")
		}
		ndkCompilers, _ := filepath.Glob(filepath.Join(androidNDK, "toolchains", "llvm", "prebuilt", "*", "bin", "clang"))
		if !regularFile(filepath.Join(androidNDK, "source.properties")) || !slices.ContainsFunc(ndkCompilers, executableFile) {
			return errors.New("指定されたAndroid NDKが完全ではありません")
		}
		sdkManagers, err := filepath.Glob(filepath.Join(androidSDK, "cmdline-tools", "*", "bin", "sdkmanager"))
		if err != nil || !slices.ContainsFunc(sdkManagers, executableFile) {
			return errors.New("Android向け生成にはAndroid SDK Command-line Toolsが必要です")
		}
	}
	return nil
}

func directoryExists(path string) bool {
	// #nosec G703 -- 利用者が指定したSDKディレクトリの存在確認だけを行います。
	info, err := os.Stat(filepath.Clean(path))
	return err == nil && info.IsDir()
}

func executableFile(path string) bool {
	// #nosec G703 -- 利用者が指定したSDK内の実行ファイルだけを確認します。
	info, err := os.Stat(filepath.Clean(path))
	return err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0
}

func regularFile(path string) bool {
	// #nosec G703 -- 利用者が指定したSDK内のファイルだけを確認します。
	info, err := os.Stat(filepath.Clean(path))
	return err == nil && info.Mode().IsRegular()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func resolveToolVersions(ctx context.Context, tools []Tool, runner commandRunner) (map[string]string, error) {
	versions := make(map[string]string, len(tools))
	for _, tool := range tools {
		version, err := runner.Output(ctx, ".", []string{"mise", "latest", tool.Query})
		if err != nil {
			return nil, fmt.Errorf("%sの安定版を解決できません: %w", tool.Key, err)
		}
		if version == "" || strings.EqualFold(version, "latest") {
			return nil, fmt.Errorf("%sの完全な版番号を解決できません", tool.Key)
		}
		versions[tool.Key] = strings.TrimPrefix(version, "v")
	}
	return versions, nil
}

func writeProjectFiles(target string, files []File) error {
	if err := os.MkdirAll(target, 0o750); err != nil {
		return fmt.Errorf("生成先を作成できません: %w", err)
	}
	for _, file := range files {
		path := pathFor(target, file.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return fmt.Errorf("ディレクトリを作成できません: %w", err)
		}
		if err := os.WriteFile(path, file.Content, file.Mode); err != nil {
			return fmt.Errorf("%sを作成できません: %w", file.Path, err)
		}
	}
	return nil
}

func configureAllowedBuilds(target, ignoredBuildsOutput string) error {
	lockPath := filepath.Join(target, "pnpm-lock.yaml")
	// #nosec G304 -- lockPathは検証済み生成先直下の固定ファイル名です。
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return fmt.Errorf("pnpm-lock.yamlを読み取れません: %w", err)
	}
	dependencies, err := buildDependenciesFromLock(data, ignoredBuildNames(ignoredBuildsOutput))
	if err != nil {
		return err
	}
	workspace := pnpmWorkspace(dependencies)
	// #nosec G306 -- 生成する共有設定ファイルは通常の読み取り権限が必要です。
	if err := os.WriteFile(filepath.Join(target, "pnpm-workspace.yaml"), []byte(workspace), 0o644); err != nil {
		return fmt.Errorf("pnpm-workspace.yamlを更新できません: %w", err)
	}
	return nil
}

func ignoredBuildNames(output string) []string {
	names := []string{}
	inList := false
	for line := range strings.SplitSeq(output, "\n") {
		switch {
		case line == "Automatically ignored builds during installation:":
			inList = true
		case inList && strings.HasPrefix(line, "  "):
			name := strings.TrimSpace(line)
			if name != "None" && !strings.ContainsAny(name, " \t") {
				names = append(names, name)
			}
		case inList && strings.TrimSpace(line) != "":
			inList = false
		}
	}
	slices.Sort(names)
	return slices.Compact(names)
}

func buildDependenciesFromLock(data []byte, ignoredNames []string) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	section := ""
	dependencies := []string{}
	ignored := make(map[string]bool, len(ignoredNames))
	for _, name := range ignoredNames {
		ignored[name] = true
	}
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if line == "packages:" {
			section = "packages"
			continue
		}
		if section != "" && line != "" && line[0] != ' ' {
			section = ""
			continue
		}
		if section == "" {
			continue
		}
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.HasSuffix(trimmed, ":") {
			value := strings.Trim(strings.TrimSuffix(trimmed, ":"), "'\"")
			name, version, ok := splitLockPackage(value)
			if !ok {
				continue
			}
			if !ignored[name] {
				continue
			}
			selector := name + "@" + version
			if !reviewedBuildDependencies[selector] {
				return nil, fmt.Errorf("未確認のビルド処理を検出しました: %s", selector)
			}
			dependencies = append(dependencies, selector)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("pnpm-lock.yamlを解析できません: %w", err)
	}
	slices.Sort(dependencies)
	dependencies = slices.Compact(dependencies)
	for _, name := range ignoredNames {
		if !slices.ContainsFunc(dependencies, func(selector string) bool {
			return strings.HasPrefix(selector, name+"@")
		}) {
			return nil, fmt.Errorf("ビルド依存関係をロックファイルから解決できません: %s", name)
		}
	}
	return dependencies, nil
}

func splitLockPackage(value string) (string, string, bool) {
	if base, _, found := strings.Cut(value, "("); found {
		value = base
	}
	name, version, found := strings.CutLast(value, "@")
	if !found || name == "" || version == "" {
		return "", "", false
	}
	return name, version, true
}
