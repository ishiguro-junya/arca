package command

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ishiguro-junya/arca/internal/buildinfo"
)

func TestCheckPlatform(t *testing.T) {
	if err := checkPlatform("darwin", "arm64"); err != nil {
		t.Fatal(err)
	}
	for _, platform := range [][2]string{{"darwin", "amd64"}, {"linux", "arm64"}, {"windows", "amd64"}} {
		if err := checkPlatform(platform[0], platform[1]); err == nil {
			t.Errorf("未対応環境が許可されました: %s/%s", platform[0], platform[1])
		}
	}
}

func TestVersionJSON(t *testing.T) {
	oldVersion, oldCommit, oldBuildTime := buildinfo.Version, buildinfo.Commit, buildinfo.BuildTime
	t.Cleanup(func() {
		buildinfo.Version, buildinfo.Commit, buildinfo.BuildTime = oldVersion, oldCommit, oldBuildTime
	})
	buildinfo.Version, buildinfo.Commit, buildinfo.BuildTime = "1.2.3", "abc123", "2026-09-02T00:00:00Z"
	var out bytes.Buffer
	cmd := NewRootCommand(bytes.NewReader(nil), &out, &out)
	cmd.SetArgs([]string{"version", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"version": "1.2.3"`, `"commit": "abc123"`, `"buildTime": "2026-09-02T00:00:00Z"`} {
		if !strings.Contains(out.String(), expected) {
			t.Errorf("JSONに%sがありません:\n%s", expected, out.String())
		}
	}
}

func TestPublicCommands(t *testing.T) {
	cmd := NewRootCommand(bytes.NewReader(nil), &bytes.Buffer{}, &bytes.Buffer{})
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, child := range cmd.Commands() {
		if child.Name() == "completion" {
			t.Fatal("対象外のcompletionコマンドが公開されています")
		}
	}
}

func TestUpdateRejectsUnsupportedPlatform(t *testing.T) {
	var out bytes.Buffer
	cmd := newRootCommand(bytes.NewReader(nil), &out, &out, "linux", "arm64")
	cmd.SetArgs([]string{"update", "--check"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "macOS Apple Silicon") {
		t.Fatalf("未対応環境でupdateが拒否されませんでした: %v", err)
	}
}
