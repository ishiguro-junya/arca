package update

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/ishiguro-junya/arca/internal/buildinfo"
)

type fakeClient struct {
	latest releaseInfo
	found  bool
	err    error
	update bool
}

func (f *fakeClient) DetectLatest(context.Context) (releaseInfo, bool, error) {
	return f.latest, f.found, f.err
}

func (f *fakeClient) UpdateTo(context.Context, releaseInfo, string) error {
	f.update = true
	return f.err
}

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		left  string
		right string
		want  int
	}{
		{left: "1.2.3", right: "1.2.4", want: -1},
		{left: "1.2.3", right: "1.2.3", want: 0},
		{left: "2.0.0", right: "1.9.9", want: 1},
		{left: "1.0.0-alpha.1", right: "1.0.0-alpha.2", want: -1},
		{left: "1.0.0-rc.1", right: "1.0.0", want: -1},
	}
	for _, test := range tests {
		got, err := compareSemver(test.left, test.right)
		if err != nil {
			t.Fatal(err)
		}
		if got != test.want {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", test.left, test.right, got, test.want)
		}
	}
}

func TestPrereleaseSelection(t *testing.T) {
	for _, version := range []string{"1.0.0-alpha.1", "1.0.0-beta.2", "1.0.0-rc.1"} {
		if !acceptsPrerelease(version) {
			t.Errorf("プレリリースとして認識されません: %s", version)
		}
	}
	for _, version := range []string{"1.0.0", "1.0.0-preview.1", "dev"} {
		if acceptsPrerelease(version) {
			t.Errorf("プレリリースとして誤認識されました: %s", version)
		}
	}
}

func TestInstallMethod(t *testing.T) {
	for _, path := range []string{"/opt/homebrew/Cellar/arca/1.0.0/bin/arca", "/usr/local/Homebrew/Cellar/arca/1.0.0/bin/arca"} {
		if method := detectInstallMethod(path); method != "homebrew" {
			t.Errorf("Homebrewを検出できません: %s = %s", path, method)
		}
	}
	if method := detectInstallMethod("/Users/test/.local/bin/arca"); method != "direct" {
		t.Fatalf("直接インストールを誤判定しました: %s", method)
	}
}

func TestCheckJSON(t *testing.T) {
	setVersion(t, "1.0.0")
	client := &fakeClient{latest: releaseInfo{Version: "1.1.0"}, found: true}
	var out bytes.Buffer
	err := run(t.Context(), bytes.NewReader(nil), &out, io.Discard, options{checkOnly: true, asJSON: true}, testDependencies(client, "/Users/test/.local/bin/arca", false))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"currentVersion": "1.0.0"`, `"latestVersion": "1.1.0"`, `"updateAvailable": true`, `"installMethod": "direct"`} {
		if !strings.Contains(out.String(), expected) {
			t.Errorf("JSONに%sがありません:\n%s", expected, out.String())
		}
	}
	if client.update {
		t.Fatal("--checkで更新されました")
	}
}

func TestInteractiveJSONKeepsPromptSeparate(t *testing.T) {
	setVersion(t, "1.0.0")
	client := &fakeClient{latest: releaseInfo{Version: "1.1.0"}, found: true}
	var out bytes.Buffer
	var prompt bytes.Buffer
	err := run(t.Context(), strings.NewReader("y\n"), &out, &prompt, options{asJSON: true}, testDependencies(client, "/Users/test/.local/bin/arca", true))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "更新しますか") || !strings.Contains(prompt.String(), "更新しますか") {
		t.Fatalf("確認文がJSON出力から分離されていません: out=%q prompt=%q", out.String(), prompt.String())
	}
}

func TestNonInteractiveUpdateRequiresYes(t *testing.T) {
	setVersion(t, "1.0.0")
	client := &fakeClient{latest: releaseInfo{Version: "1.1.0"}, found: true}
	err := run(t.Context(), bytes.NewReader(nil), io.Discard, io.Discard, options{}, testDependencies(client, "/Users/test/.local/bin/arca", false))
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("--yes必須エラーではありません: %v", err)
	}
	if client.update {
		t.Fatal("確認なしで更新されました")
	}
}

func TestYesUpdatesDirectInstall(t *testing.T) {
	setVersion(t, "1.0.0-alpha.1")
	fake := &fakeClient{latest: releaseInfo{Version: "1.0.0-alpha.2"}, found: true}
	var prerelease bool
	deps := testDependencies(fake, "/Users/test/.local/bin/arca", false)
	deps.newClient = func(value bool) (client, error) {
		prerelease = value
		return fake, nil
	}
	if err := run(t.Context(), bytes.NewReader(nil), io.Discard, io.Discard, options{yes: true}, deps); err != nil {
		t.Fatal(err)
	}
	if !fake.update || !prerelease {
		t.Fatalf("プレリリース更新が実行されませんでした: update=%v prerelease=%v", fake.update, prerelease)
	}
}

func TestHomebrewAndDevDoNotCreateClient(t *testing.T) {
	for _, test := range []struct {
		name    string
		version string
		path    string
		want    string
	}{
		{name: "homebrew", version: "1.0.0", path: "/opt/homebrew/Cellar/arca/1.0.0/bin/arca", want: brewCommand},
		{name: "dev", version: "dev", path: "/Users/test/arca", want: "開発ビルド"},
	} {
		t.Run(test.name, func(t *testing.T) {
			setVersion(t, test.version)
			deps := testDependencies(nil, test.path, false)
			deps.newClient = func(bool) (client, error) { return nil, errors.New("呼び出されました") }
			var out bytes.Buffer
			if err := run(t.Context(), bytes.NewReader(nil), &out, io.Discard, options{}, deps); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), test.want) {
				t.Fatalf("案内が不正です:\n%s", out.String())
			}
		})
	}
}

func setVersion(t *testing.T, version string) {
	t.Helper()
	old := buildinfo.Version
	buildinfo.Version = version
	t.Cleanup(func() { buildinfo.Version = old })
}

func testDependencies(updateClient client, path string, interactive bool) dependencies {
	return dependencies{
		executablePath: func() (string, error) { return path, nil },
		isInteractive:  func(io.Reader) bool { return interactive },
		newClient:      func(bool) (client, error) { return updateClient, nil },
	}
}
