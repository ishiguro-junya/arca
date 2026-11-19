package update

import (
	"bufio"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/ishiguro-junya/arca/internal/buildinfo"
	"github.com/spf13/cobra"
)

const (
	repositoryOwner = "ishiguro-junya"
	repositoryName  = "arca"
	brewCommand     = "brew upgrade ishiguro-junya/arca/arca"
)

type options struct {
	checkOnly bool
	yes       bool
	asJSON    bool
}

type Result struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	Prerelease      bool   `json:"prerelease"`
	InstallMethod   string `json:"installMethod"`
}

type releaseInfo struct {
	Version string
	raw     *selfupdate.Release
}

type client interface {
	DetectLatest(context.Context) (releaseInfo, bool, error)
	UpdateTo(context.Context, releaseInfo, string) error
}

type dependencies struct {
	executablePath func() (string, error)
	isInteractive  func(io.Reader) bool
	newClient      func(bool) (client, error)
}

func NewCommand(in io.Reader, out, promptOut io.Writer) *cobra.Command {
	opts := options{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Arcaを更新します",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), in, out, promptOut, opts, defaultDependencies())
		},
	}
	cmd.Flags().BoolVar(&opts.checkOnly, "check", false, "最新版の有無だけを確認します")
	cmd.Flags().BoolVar(&opts.yes, "yes", false, "確認を省略して更新します")
	cmd.Flags().BoolVar(&opts.asJSON, "json", false, "JSON形式で表示します")
	return cmd
}

func defaultDependencies() dependencies {
	return dependencies{
		executablePath: selfupdate.ExecutablePath,
		isInteractive:  isInteractive,
		newClient:      newGitHubClient,
	}
}

func run(ctx context.Context, in io.Reader, out, promptOut io.Writer, opts options, deps dependencies) error {
	current := buildinfo.Version
	executable, err := deps.executablePath()
	if err != nil {
		return fmt.Errorf("実行ファイルのパスを取得できません: %w", err)
	}
	method := detectInstallMethod(executable)
	if method == "homebrew" {
		result := Result{CurrentVersion: current, InstallMethod: method}
		return writeResult(out, opts.asJSON, result, "Homebrewで管理されています。"+brewCommand+"を実行してください。")
	}
	current, err = normalizeSemver(current)
	if err != nil {
		result := Result{CurrentVersion: buildinfo.Version, InstallMethod: method}
		return writeResult(out, opts.asJSON, result, "開発ビルドでは自己更新を利用できません。")
	}
	allowPrerelease := acceptsPrerelease(current)
	updateClient, err := deps.newClient(allowPrerelease)
	if err != nil {
		return err
	}
	latest, found, err := updateClient.DetectLatest(ctx)
	if err != nil {
		return fmt.Errorf("最新版の確認に失敗しました: %w", err)
	}
	if !found {
		return errors.New("macOS Apple Silicon向けのReleaseが見つかりません")
	}
	latestVersion, err := normalizeSemver(latest.Version)
	if err != nil {
		return fmt.Errorf("releaseのバージョンを解釈できません: %w", err)
	}
	comparison, err := compareSemver(current, latestVersion)
	if err != nil {
		return err
	}
	result := Result{
		CurrentVersion:  current,
		LatestVersion:   latestVersion,
		UpdateAvailable: comparison < 0,
		Prerelease:      allowPrerelease,
		InstallMethod:   method,
	}
	if comparison >= 0 {
		return writeResult(out, opts.asJSON, result, "現在のArcaは最新です。")
	}
	if opts.checkOnly {
		return writeResult(out, opts.asJSON, result, "新しいArcaが利用できます。")
	}
	if !opts.yes {
		if !deps.isInteractive(in) {
			return errors.New("非対話環境で更新する場合は--yesを指定してください")
		}
		confirmed, err := confirm(in, promptOut, current, latestVersion)
		if err != nil {
			return err
		}
		if !confirmed {
			return writeResult(out, opts.asJSON, result, "更新をキャンセルしました。")
		}
	}
	if err := updateClient.UpdateTo(ctx, latest, executable); err != nil {
		return fmt.Errorf("自己更新に失敗しました: %w", err)
	}
	return writeResult(out, opts.asJSON, result, "Arcaを更新しました。")
}

func writeResult(out io.Writer, asJSON bool, result Result, message string) error {
	if asJSON {
		data, err := json.Marshal(&result, json.Deterministic(true), jsontext.WithIndent("  "))
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, string(data))
		return err
	}
	_, err := fmt.Fprintf(out, "%s\ncurrent: %s\nlatest: %s\ninstall method: %s\n", message, result.CurrentVersion, result.LatestVersion, result.InstallMethod)
	return err
}

func confirm(in io.Reader, out io.Writer, current, latest string) (bool, error) {
	if _, err := fmt.Fprintf(out, "Arcaを%sから%sへ更新しますか？ [y/N]: ", current, latest); err != nil {
		return false, err
	}
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return false, scanner.Err()
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}

func isInteractive(in io.Reader) bool {
	file, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func detectInstallMethod(executable string) string {
	paths := []string{executable}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil && resolved != executable {
		paths = append(paths, resolved)
	}
	for _, path := range paths {
		normalized := strings.ToLower(filepath.ToSlash(filepath.Clean(path)))
		if strings.Contains(normalized, "/cellar/arca/") || strings.Contains(normalized, "/homebrew/cellar/arca/") {
			return "homebrew"
		}
	}
	return "direct"
}

type selfUpdateClient struct {
	updater *selfupdate.Updater
	repo    selfupdate.Repository
}

func newGitHubClient(prerelease bool) (client, error) {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return nil, err
	}
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:     source,
		Validator:  &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
		Prerelease: prerelease,
		OS:         "darwin",
		Arch:       "arm64",
	})
	if err != nil {
		return nil, err
	}
	return &selfUpdateClient{
		updater: updater,
		repo:    selfupdate.NewRepositorySlug(repositoryOwner, repositoryName),
	}, nil
}

func (c *selfUpdateClient) DetectLatest(ctx context.Context) (releaseInfo, bool, error) {
	release, found, err := c.updater.DetectLatest(ctx, c.repo)
	if err != nil || !found {
		return releaseInfo{}, found, err
	}
	return releaseInfo{Version: release.Version(), raw: release}, true, nil
}

func (c *selfUpdateClient) UpdateTo(ctx context.Context, release releaseInfo, executable string) error {
	if release.raw == nil {
		return errors.New("更新対象のRelease情報が不足しています")
	}
	return c.updater.UpdateTo(ctx, release.raw, executable)
}
