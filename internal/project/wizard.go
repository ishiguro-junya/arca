package project

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"charm.land/huh/v2"
)

func RunWizard(ctx context.Context, in io.Reader, out io.Writer) error {
	spec := DefaultSpec()
	if err := runForm(ctx, in, out,
		huh.NewInput().Title("プロジェクト名").Value(&spec.Name).Validate(validateProjectName),
	); err != nil {
		return err
	}
	spec.Directory = "./" + spec.Name
	if err := runForm(ctx, in, out,
		huh.NewInput().Title("生成先").Value(&spec.Directory).Validate(requireValue("生成先")),
		huh.NewSelect[ReadmeLanguage]().Title("READMEの言語").Options(
			huh.NewOption("日本語", ReadmeJapanese),
			huh.NewOption("English", ReadmeEnglish),
		).Value(&spec.ReadmeLanguage),
		huh.NewMultiSelect[UseCase]().Title("用途").Options(
			huh.NewOption("App", UseApp).Selected(true),
			huh.NewOption("Backend", UseBackend),
			huh.NewOption("CLI", UseCLI),
			huh.NewOption("Infra", UseInfra),
		).Value(&spec.UseCases).Validate(requireUseCase),
	); err != nil {
		return err
	}
	if spec.Has(UseApp) {
		if err := collectApp(ctx, in, out, &spec); err != nil {
			return err
		}
	}
	if spec.Has(UseCLI) {
		if err := runForm(ctx, in, out,
			huh.NewSelect[CLILanguage]().Title("CLIの言語").Options(
				huh.NewOption("Go", CLIGo),
				huh.NewOption("Node.js", CLINode),
			).Value(&spec.CLILanguage),
		); err != nil {
			return err
		}
	}
	if spec.Has(UseInfra) {
		if err := runForm(ctx, in, out,
			huh.NewSelect[string]().Title("Terraformプロバイダー").Options(
				huh.NewOption("Azure", "azure"),
				huh.NewOption("AWS", "aws"),
				huh.NewOption("GCP", "gcp"),
				huh.NewOption("Cloudflare", "cloudflare"),
				huh.NewOption("Vercel", "vercel"),
			).Value(&spec.InfraProvider),
		); err != nil {
			return err
		}
	}
	if err := collectFeatures(ctx, in, out, &spec); err != nil {
		return err
	}
	if err := spec.Validate(); err != nil {
		return err
	}
	if err := Preview(spec, out); err != nil {
		return err
	}
	if err := runForm(ctx, in, out,
		huh.NewSelect[Action]().Title("実行内容").Options(
			huh.NewOption("生成する", ActionGenerate),
			huh.NewOption("プレビューだけで終了する", ActionPreview),
			huh.NewOption("キャンセルする", ActionCancel),
		).Value(&spec.Action),
	); err != nil {
		return err
	}
	switch spec.Action {
	case ActionGenerate:
		return Generate(ctx, spec, out)
	case ActionPreview:
		return nil
	case ActionCancel:
		_, err := fmt.Fprintln(out, "キャンセルしました。")
		return err
	default:
		return errors.New("実行内容を選択してください")
	}
}

func collectApp(ctx context.Context, in io.Reader, out io.Writer, spec *Spec) error {
	if err := runForm(ctx, in, out,
		huh.NewSelect[AppKind]().Title("Appの種類").Options(
			huh.NewOption("Web", AppWeb),
			huh.NewOption("Desktop", AppDesktop),
		).Value(&spec.AppKind),
	); err != nil {
		return err
	}
	if spec.AppKind == AppWeb {
		return runForm(ctx, in, out,
			huh.NewSelect[WebFramework]().Title("Webフレームワーク").Options(
				huh.NewOption("React + Vite", WebVite),
				huh.NewOption("Next.js", WebNext),
			).Value(&spec.WebFramework),
		)
	}
	spec.DesktopPlatforms = []string{"macos"}
	return runForm(ctx, in, out,
		huh.NewMultiSelect[string]().Title("Desktopの対象プラットフォーム").Options(
			huh.NewOption("macOS", "macos").Selected(true),
			huh.NewOption("Linux", "linux"),
			huh.NewOption("Windows", "windows"),
			huh.NewOption("iOS", "ios"),
			huh.NewOption("Android", "android"),
		).Value(&spec.DesktopPlatforms).Validate(requireSelection("対象プラットフォーム")),
	)
}

func collectFeatures(ctx context.Context, in io.Reader, out io.Writer, spec *Spec) error {
	categories := []string{}
	options := featureCategoryOptions(*spec)
	if len(options) == 0 {
		return nil
	}
	if err := runForm(ctx, in, out,
		huh.NewMultiSelect[string]().Title("オプションカテゴリ").Options(options...).Value(&categories),
	); err != nil {
		return err
	}
	if slices.Contains(categories, "uiux") {
		if err := selectString(ctx, in, out, "UI/UX", &spec.Features.UIUX,
			huh.NewOption("assistant-ui", "assistant-ui"),
			huh.NewOption("Radix UI", "radix"),
			huh.NewOption("shadcn/ui", "shadcn"),
			huh.NewOption("TanStack Query", "tanstack-query"),
		); err != nil {
			return err
		}
	}
	spec.Features.Localization = slices.Contains(categories, "localization")
	if slices.Contains(categories, "state") {
		if err := selectString(ctx, in, out, "State", &spec.Features.State, huh.NewOption("Jotai", "jotai"), huh.NewOption("Zustand", "zustand")); err != nil {
			return err
		}
	}
	if slices.Contains(categories, "form") {
		if err := selectString(ctx, in, out, "Form", &spec.Features.Form, huh.NewOption("React Hook Form", "react-hook-form"), huh.NewOption("TanStack Form", "tanstack-form")); err != nil {
			return err
		}
	}
	if slices.Contains(categories, "validate") {
		if err := selectString(ctx, in, out, "Validate", &spec.Features.Validate, huh.NewOption("Valibot", "valibot"), huh.NewOption("Zod", "zod")); err != nil {
			return err
		}
	}
	if slices.Contains(categories, "icon") {
		if err := selectString(ctx, in, out, "Icon", &spec.Features.Icon,
			huh.NewOption("Font Awesome", "font-awesome"),
			huh.NewOption("Lucide", "lucide"),
			huh.NewOption("Material Design Icons", "mdi"),
			huh.NewOption("React Icons", "react-icons"),
			huh.NewOption("SVGR", "svgr"),
		); err != nil {
			return err
		}
	}
	if slices.Contains(categories, "font") {
		spec.Features.FontEnabled = true
		spec.Features.Fonts = []string{"inter", "noto-sans-jp"}
		if err := runForm(ctx, in, out,
			huh.NewMultiSelect[string]().Title("Font").Options(
				huh.NewOption("Noto Sans JP", "noto-sans-jp").Selected(true),
				huh.NewOption("Inter", "inter").Selected(true),
			).Value(&spec.Features.Fonts).Validate(requireSelection("フォント")),
		); err != nil {
			return err
		}
	}
	if slices.Contains(categories, "orm") {
		options := ormOptions(*spec)
		if err := selectString(ctx, in, out, "ORM", &spec.Features.ORM, options...); err != nil {
			return err
		}
	}
	return nil
}

func featureCategoryOptions(spec Spec) []huh.Option[string] {
	options := []huh.Option[string]{}
	if spec.Has(UseApp) {
		options = append(options,
			huh.NewOption("UI/UX", "uiux"),
			huh.NewOption("Localization", "localization"),
			huh.NewOption("State", "state"),
			huh.NewOption("Form", "form"),
			huh.NewOption("Validate", "validate"),
			huh.NewOption("Icon", "icon"),
			huh.NewOption("Font", "font"),
		)
	}
	if len(ormOptions(spec)) > 0 {
		options = append(options, huh.NewOption("ORM", "orm"))
	}
	return options
}

func ormOptions(spec Spec) []huh.Option[string] {
	options := []huh.Option[string]{}
	if spec.IsNext() {
		options = append(options, huh.NewOption("Drizzle", "drizzle"), huh.NewOption("Prisma", "prisma"))
	}
	if spec.Has(UseBackend) {
		options = append(options, huh.NewOption("SQLAlchemy + Alembic", "sqlalchemy"))
	}
	if spec.HasGoCLI() {
		options = append(options, huh.NewOption("sqlc", "sqlc"))
	}
	return options
}

func selectString(ctx context.Context, in io.Reader, out io.Writer, title string, value *string, options ...huh.Option[string]) error {
	return runForm(ctx, in, out, huh.NewSelect[string]().Title(title).Options(options...).Value(value))
}

func runForm(ctx context.Context, in io.Reader, out io.Writer, fields ...huh.Field) error {
	return huh.NewForm(huh.NewGroup(fields...)).
		WithInput(in).
		WithOutput(out).
		WithAccessible(strings.EqualFold(os.Getenv("ARCA_ACCESSIBLE"), "true")).
		RunWithContext(ctx)
}

func validateProjectName(value string) error {
	if !projectNamePattern.MatchString(value) {
		return errors.New("小文字英字で始まる英小文字、数字、ハイフンだけで入力してください")
	}
	return nil
}

func requireValue(name string) func(string) error {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%sを入力してください", name)
		}
		return nil
	}
}

func requireUseCase(values []UseCase) error {
	if len(values) == 0 {
		return errors.New("用途を1つ以上選択してください")
	}
	return nil
}

func requireSelection(name string) func([]string) error {
	return func(values []string) error {
		if len(values) == 0 {
			return fmt.Errorf("%sを1つ以上選択してください", name)
		}
		return nil
	}
}
