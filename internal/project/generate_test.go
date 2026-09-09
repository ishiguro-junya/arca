package project

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type fakeRunner struct {
	commands [][]string
}

func (f *fakeRunner) Output(_ context.Context, _ string, args []string) (string, error) {
	if slices.Contains(args, "ignored-builds") {
		return "No automatically ignored builds", nil
	}
	return "1.2.3", nil
}

func (f *fakeRunner) Run(_ context.Context, directory string, args []string, _ io.Writer) error {
	f.commands = append(f.commands, slices.Clone(args))
	joined := strings.Join(args, " ")
	switch {
	case strings.Contains(joined, "pnpm install --config.strict-dep-builds=false"):
		return os.WriteFile(filepath.Join(directory, "pnpm-lock.yaml"), []byte("lockfileVersion: '9.0'\npackages:\n"), 0o644)
	case strings.Contains(joined, "uv lock"):
		return os.WriteFile(filepath.Join(directory, "backend", "uv.lock"), []byte("version = 1\n"), 0o644)
	case strings.Contains(joined, "go mod tidy"):
		return os.WriteFile(filepath.Join(directory, "go.sum"), []byte("example sum\n"), 0o644)
	case strings.Contains(joined, "terraform -chdir=infra init"):
		if err := os.MkdirAll(filepath.Join(directory, "infra", ".terraform"), 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "infra", ".terraform.lock.hcl"), []byte("provider lock\n"), 0o644)
	default:
		return nil
	}
}

func TestGenerateWritesProjectAndLocks(t *testing.T) {
	target := filepath.Join(t.TempDir(), "generated")
	spec := validSpec(target)
	spec.UseCases = []UseCase{UseBackend, UseCLI, UseInfra}
	spec.CLILanguage = CLIGo
	spec.InfraProvider = "azure"
	runner := &fakeRunner{}
	if err := generate(t.Context(), spec, io.Discard, runner); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"mise.toml", "pnpm-lock.yaml", "backend/uv.lock", "go.sum", "infra/.terraform.lock.hcl", "cli/main.go"} {
		if _, err := os.Stat(filepath.Join(target, path)); err != nil {
			t.Errorf("生成物がありません: %s: %v", path, err)
		}
	}
	for _, path := range []string{"mise.lock", "cmd", "infra/.terraform"} {
		if _, err := os.Stat(filepath.Join(target, path)); !os.IsNotExist(err) {
			t.Errorf("不要な生成物があります: %s: %v", path, err)
		}
	}
}

func TestGenerateRejectsNonEmptyTargetBeforeCommands(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "existing.txt"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := validSpec(target)
	runner := &fakeRunner{}
	err := generate(t.Context(), spec, io.Discard, runner)
	if err == nil || !strings.Contains(err.Error(), "空ではありません") {
		t.Fatalf("空でない生成先が拒否されませんでした: %v", err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("拒否前にコマンドが実行されました: %v", runner.commands)
	}
}

func TestValidateMobileSDKsRejectsMissingAndroidDirectories(t *testing.T) {
	t.Setenv("ANDROID_HOME", filepath.Join(t.TempDir(), "missing-sdk"))
	t.Setenv("NDK_HOME", filepath.Join(t.TempDir(), "missing-ndk"))
	spec := validSpec(t.TempDir())
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"android"}
	err := validateMobileSDKs(spec)
	if err == nil || !strings.Contains(err.Error(), "ディレクトリ") {
		t.Fatalf("存在しないAndroid SDKが拒否されませんでした: %v", err)
	}
}

func TestValidateMobileSDKsRejectsIncompleteAndroidSDK(t *testing.T) {
	androidSDK := t.TempDir()
	androidNDK := t.TempDir()
	writeSDKFile(t, filepath.Join(androidNDK, "source.properties"), 0o644)
	writeSDKFile(t, filepath.Join(androidNDK, "toolchains", "llvm", "prebuilt", "darwin-x86_64", "bin", "clang"), 0o755)
	sdkManager := filepath.Join(androidSDK, "cmdline-tools", "latest", "bin", "sdkmanager")
	writeSDKFile(t, sdkManager, 0o755)
	t.Setenv("ANDROID_HOME", androidSDK)
	t.Setenv("NDK_HOME", androidNDK)
	spec := validSpec(t.TempDir())
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"android"}
	err := validateMobileSDKs(spec)
	if err == nil || !strings.Contains(err.Error(), "Android SDK") {
		t.Fatalf("不完全なAndroid SDKが拒否されませんでした: %v", err)
	}
}

func TestValidateMobileSDKsRejectsNonExecutableSDKManager(t *testing.T) {
	androidSDK := t.TempDir()
	androidNDK := t.TempDir()
	writeSDKFile(t, filepath.Join(androidNDK, "source.properties"), 0o644)
	writeSDKFile(t, filepath.Join(androidNDK, "toolchains", "llvm", "prebuilt", "darwin-x86_64", "bin", "clang"), 0o755)
	writeSDKFile(t, filepath.Join(androidSDK, "platform-tools", "adb"), 0o755)
	writeSDKFile(t, filepath.Join(androidSDK, "platforms", "android-35", "android.jar"), 0o644)
	writeSDKFile(t, filepath.Join(androidSDK, "build-tools", "35.0.0", "aapt2"), 0o755)
	writeSDKFile(t, filepath.Join(androidSDK, "cmdline-tools", "latest", "bin", "sdkmanager"), 0o644)
	t.Setenv("ANDROID_HOME", androidSDK)
	t.Setenv("NDK_HOME", androidNDK)
	spec := validSpec(t.TempDir())
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"android"}
	err := validateMobileSDKs(spec)
	if err == nil || !strings.Contains(err.Error(), "Command-line Tools") {
		t.Fatalf("実行できないsdkmanagerが拒否されませんでした: %v", err)
	}
}

func TestValidateMobileSDKsAcceptsCompleteAndroidSDK(t *testing.T) {
	androidSDK := t.TempDir()
	androidNDK := t.TempDir()
	writeSDKFile(t, filepath.Join(androidNDK, "source.properties"), 0o644)
	writeSDKFile(t, filepath.Join(androidNDK, "toolchains", "llvm", "prebuilt", "darwin-x86_64", "bin", "clang"), 0o755)
	writeSDKFile(t, filepath.Join(androidSDK, "platform-tools", "adb"), 0o755)
	writeSDKFile(t, filepath.Join(androidSDK, "platforms", "android-35", "android.jar"), 0o644)
	writeSDKFile(t, filepath.Join(androidSDK, "build-tools", "35.0.0", "aapt2"), 0o755)
	writeSDKFile(t, filepath.Join(androidSDK, "cmdline-tools", "latest", "bin", "sdkmanager"), 0o755)
	t.Setenv("ANDROID_HOME", androidSDK)
	t.Setenv("NDK_HOME", androidNDK)
	spec := validSpec(t.TempDir())
	spec.AppKind = AppDesktop
	spec.DesktopPlatforms = []string{"android"}
	if err := validateMobileSDKs(spec); err != nil {
		t.Fatalf("完全なAndroid SDKが拒否されました: %v", err)
	}
}

func writeSDKFile(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fixture\n"), mode); err != nil {
		t.Fatal(err)
	}
}
