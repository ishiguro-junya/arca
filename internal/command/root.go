package command

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/ishiguro-junya/arca/internal/buildinfo"
	"github.com/ishiguro-junya/arca/internal/project"
	"github.com/ishiguro-junya/arca/internal/update"
	"github.com/spf13/cobra"
)

func Execute() error {
	return NewRootCommand(os.Stdin, os.Stdout, os.Stderr).Execute()
}

func NewRootCommand(in io.Reader, out, errOut io.Writer) *cobra.Command {
	return newRootCommand(in, out, errOut, runtime.GOOS, runtime.GOARCH)
}

func newRootCommand(in io.Reader, out, errOut io.Writer, goos, goarch string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "arca",
		Short:             "開発プロジェクトを対話形式でセットアップします",
		SilenceErrors:     true,
		SilenceUsage:      true,
		Args:              cobra.NoArgs,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := checkPlatform(goos, goarch); err != nil {
				return err
			}
			return project.RunWizard(cmd.Context(), in, out)
		},
	}
	updateCommand := update.NewCommand(in, out, errOut)
	updateCommand.PreRunE = func(*cobra.Command, []string) error {
		return checkPlatform(goos, goarch)
	}
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.AddCommand(newVersionCommand(out), updateCommand)
	return cmd
}

func checkPlatform(goos, goarch string) error {
	if goos != "darwin" || goarch != "arm64" {
		return fmt.Errorf("arca は macOS Apple Silicon のみ対応しています: %s/%s", goos, goarch)
	}
	return nil
}

func newVersionCommand(out io.Writer) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "バージョン情報を表示します",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			info := buildinfo.Current()
			if asJSON {
				data, err := json.Marshal(&info, json.Deterministic(true), jsontext.WithIndent("  "))
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(out, string(data))
				return err
			}
			_, err := fmt.Fprintf(out, "arca %s\ncommit: %s\nbuild time: %s\n", info.Version, info.Commit, info.BuildTime)
			return err
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON形式で表示します")
	return cmd
}
