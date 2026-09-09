package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallRejectsUnsupportedOS(t *testing.T) {
	for _, test := range []struct {
		name string
		os   string
		arch string
	}{
		{name: "Linux", os: "Linux", arch: "arm64"},
		{name: "macOS Intel", os: "Darwin", arch: "x86_64"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fakeBin := t.TempDir()
			writeExecutable(t, fakeBin, "uname", "#!/bin/sh\nif [ \"$1\" = \"-s\" ]; then echo "+test.os+"; else echo "+test.arch+"; fi")
			cmd := exec.CommandContext(t.Context(), "/bin/sh", "install.sh")
			cmd.Env = append(os.Environ(), "PATH="+fakeBin+":/usr/bin:/bin")
			output, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "macOS Apple Silicon") {
				t.Fatalf("未対応環境が拒否されませんでした: err=%v output=%s", err, output)
			}
		})
	}
}

func TestInstallStopsOnChecksumFailure(t *testing.T) {
	fakeBin := installFakeCommands(t, true)
	installDir := filepath.Join(t.TempDir(), "bin")
	cmd := exec.CommandContext(t.Context(), "/bin/sh", "install.sh")
	cmd.Env = append(os.Environ(), "PATH="+fakeBin+":/usr/bin:/bin", "ARCA_INSTALL_DIR="+installDir)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("チェックサム不一致で停止しませんでした: %s", output)
	}
	if _, err := os.Stat(filepath.Join(installDir, "arca")); !os.IsNotExist(err) {
		t.Fatalf("検証失敗後にバイナリが配置されました: %v", err)
	}
}

func TestInstallPlacesVerifiedBinary(t *testing.T) {
	fakeBin := installFakeCommands(t, false)
	installDir := filepath.Join(t.TempDir(), "bin")
	cmd := exec.CommandContext(t.Context(), "/bin/sh", "install.sh")
	cmd.Env = append(os.Environ(), "PATH="+fakeBin+":/usr/bin:/bin", "ARCA_INSTALL_DIR="+installDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("インストールに失敗しました: %v\n%s", err, output)
	}
	data, err := os.ReadFile(filepath.Join(installDir, "arca"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "binary\n" {
		t.Fatalf("配置されたバイナリが不正です: %q", data)
	}
}

func installFakeCommands(t *testing.T, checksumFailure bool) string {
	t.Helper()
	directory := t.TempDir()
	writeExecutable(t, directory, "uname", `#!/bin/sh
if [ "$1" = "-s" ]; then echo Darwin; else echo arm64; fi`)
	writeExecutable(t, directory, "curl", `#!/bin/sh
output=""
url=""
previous=""
for argument in "$@"; do
  if [ "$previous" = "-o" ]; then output="$argument"; fi
  previous="$argument"
  url="$argument"
done
case "$url" in
  */releases/latest) printf '%s' 'https://github.com/ishiguro-junya/arca/releases/tag/v1.2.3' ;;
  */checksums.txt) printf '%s\n' 'deadbeef  arca_darwin_arm64.tar.gz' > "$output" ;;
  */arca_darwin_arm64.tar.gz) printf '%s\n' 'archive' > "$output" ;;
esac`)
	shasum := `#!/bin/sh
cat >/dev/null
exit 0`
	if checksumFailure {
		shasum = `#!/bin/sh
cat >/dev/null
exit 1`
	}
	writeExecutable(t, directory, "shasum", shasum)
	writeExecutable(t, directory, "tar", `#!/bin/sh
destination=""
previous=""
for argument in "$@"; do
  if [ "$previous" = "-C" ]; then destination="$argument"; fi
  previous="$argument"
done
printf '%s\n' binary > "$destination/arca"`)
	writeExecutable(t, directory, "install", `#!/bin/sh
cp "$3" "$4"
chmod 0755 "$4"`)
	return directory
}

func writeExecutable(t *testing.T, directory, name, content string) {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(content+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}
