package project

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type UseCase string

const (
	UseApp     UseCase = "app"
	UseBackend UseCase = "backend"
	UseCLI     UseCase = "cli"
	UseInfra   UseCase = "infra"
)

type AppKind string

const (
	AppWeb     AppKind = "web"
	AppDesktop AppKind = "desktop"
)

type WebFramework string

const (
	WebVite WebFramework = "vite"
	WebNext WebFramework = "next"
)

type CLILanguage string

const (
	CLIGo   CLILanguage = "go"
	CLINode CLILanguage = "node"
)

type ReadmeLanguage string

const (
	ReadmeJapanese ReadmeLanguage = "ja"
	ReadmeEnglish  ReadmeLanguage = "en"
)

type Action string

const (
	ActionGenerate Action = "generate"
	ActionPreview  Action = "preview"
	ActionCancel   Action = "cancel"
)

type Features struct {
	UIUX         string
	Localization bool
	State        string
	Form         string
	Validate     string
	Icon         string
	FontEnabled  bool
	Fonts        []string
	ORM          string
}

type Spec struct {
	Name             string
	Directory        string
	ReadmeLanguage   ReadmeLanguage
	UseCases         []UseCase
	AppKind          AppKind
	WebFramework     WebFramework
	DesktopPlatforms []string
	CLILanguage      CLILanguage
	InfraProvider    string
	Features         Features
	Action           Action
}

var projectNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func DefaultSpec() Spec {
	return Spec{
		ReadmeLanguage: ReadmeJapanese,
		UseCases:       []UseCase{UseApp},
		AppKind:        AppWeb,
		WebFramework:   WebVite,
		CLILanguage:    CLIGo,
		InfraProvider:  "azure",
		Action:         ActionGenerate,
	}
}

func (s Spec) Has(useCase UseCase) bool {
	return slices.Contains(s.UseCases, useCase)
}

func (s Spec) IsNext() bool {
	return s.Has(UseApp) && s.AppKind == AppWeb && s.WebFramework == WebNext
}

func (s Spec) IsVite() bool {
	return s.Has(UseApp) && (s.AppKind == AppDesktop || s.WebFramework == WebVite)
}

func (s Spec) HasGoCLI() bool {
	return s.Has(UseCLI) && s.CLILanguage == CLIGo
}

func (s Spec) HasNodeCLI() bool {
	return s.Has(UseCLI) && s.CLILanguage == CLINode
}

func (s Spec) Validate() error {
	if !projectNamePattern.MatchString(s.Name) {
		return errors.New("プロジェクト名は小文字英字で始まる英小文字、数字、ハイフンだけで指定してください")
	}
	if strings.TrimSpace(s.Directory) == "" {
		return errors.New("生成先を指定してください")
	}
	if len(s.UseCases) == 0 {
		return errors.New("用途を1つ以上選択してください")
	}
	seenUseCases := map[UseCase]bool{}
	for _, useCase := range s.UseCases {
		if !slices.Contains([]UseCase{UseApp, UseBackend, UseCLI, UseInfra}, useCase) {
			return fmt.Errorf("未対応の用途です: %s", useCase)
		}
		if seenUseCases[useCase] {
			return fmt.Errorf("用途が重複しています: %s", useCase)
		}
		seenUseCases[useCase] = true
	}
	if s.ReadmeLanguage != ReadmeJapanese && s.ReadmeLanguage != ReadmeEnglish {
		return errors.New("READMEの言語は日本語または英語を選択してください")
	}
	if s.Has(UseApp) {
		if s.AppKind != AppWeb && s.AppKind != AppDesktop {
			return errors.New("Appの種類を選択してください")
		}
		if s.AppKind == AppWeb && s.WebFramework != WebVite && s.WebFramework != WebNext {
			return errors.New("Webフレームワークを選択してください")
		}
		if s.AppKind == AppDesktop && len(s.DesktopPlatforms) == 0 {
			return errors.New("Desktopの対象プラットフォームを1つ以上選択してください")
		}
		if err := validateChoices("Desktopの対象プラットフォーム", s.DesktopPlatforms, []string{"macos", "linux", "windows", "ios", "android"}); err != nil {
			return err
		}
	}
	if s.Has(UseCLI) && s.CLILanguage != CLIGo && s.CLILanguage != CLINode {
		return errors.New("CLIの言語を選択してください")
	}
	if s.Has(UseInfra) && !slices.Contains([]string{"aws", "azure", "gcp", "cloudflare", "vercel"}, s.InfraProvider) {
		return errors.New("Infraのプロバイダーを選択してください")
	}
	if s.Features.FontEnabled && len(s.Features.Fonts) == 0 {
		return errors.New("フォントを1つ以上選択してください")
	}
	if !s.Features.FontEnabled && len(s.Features.Fonts) > 0 {
		return errors.New("Fontカテゴリが無効な場合はフォントを指定できません")
	}
	if err := validateChoices("フォント", s.Features.Fonts, []string{"inter", "noto-sans-jp"}); err != nil {
		return err
	}
	if err := validateFeatureChoice("UI/UX", s.Features.UIUX, []string{"assistant-ui", "radix", "shadcn", "tanstack-query"}); err != nil {
		return err
	}
	if err := validateFeatureChoice("State", s.Features.State, []string{"jotai", "zustand"}); err != nil {
		return err
	}
	if err := validateFeatureChoice("Form", s.Features.Form, []string{"react-hook-form", "tanstack-form"}); err != nil {
		return err
	}
	if err := validateFeatureChoice("Validate", s.Features.Validate, []string{"valibot", "zod"}); err != nil {
		return err
	}
	if err := validateFeatureChoice("Icon", s.Features.Icon, []string{"font-awesome", "lucide", "mdi", "react-icons", "svgr"}); err != nil {
		return err
	}
	if !s.Has(UseApp) && (s.Features.UIUX != "" || s.Features.Localization || s.Features.State != "" || s.Features.Form != "" || s.Features.Validate != "" || s.Features.Icon != "" || s.Features.FontEnabled) {
		return errors.New("Appを選択していない場合はApp向けオプションを指定できません")
	}
	if err := s.validateORM(); err != nil {
		return err
	}
	return nil
}

func validateFeatureChoice(name, value string, allowed []string) error {
	if value != "" && !slices.Contains(allowed, value) {
		return fmt.Errorf("未対応の%sです: %s", name, value)
	}
	return nil
}

func validateChoices(name string, values, allowed []string) error {
	seen := map[string]bool{}
	for _, value := range values {
		if !slices.Contains(allowed, value) {
			return fmt.Errorf("未対応の%sです: %s", name, value)
		}
		if seen[value] {
			return fmt.Errorf("%sが重複しています: %s", name, value)
		}
		seen[value] = true
	}
	return nil
}

func (s Spec) validateORM() error {
	switch s.Features.ORM {
	case "":
		return nil
	case "drizzle", "prisma":
		if !s.IsNext() {
			return fmt.Errorf("%sはNext.jsでのみ選択できます", s.Features.ORM)
		}
	case "sqlalchemy":
		if !s.Has(UseBackend) {
			return errors.New("SQLAlchemyはBackendでのみ選択できます")
		}
	case "sqlc":
		if !s.HasGoCLI() {
			return errors.New("sqlcはGo CLIでのみ選択できます")
		}
	default:
		return fmt.Errorf("未対応のORMです: %s", s.Features.ORM)
	}
	return nil
}

func (s Spec) TargetDirectory() (string, error) {
	dir, err := filepath.Abs(s.Directory)
	if err != nil {
		return "", fmt.Errorf("生成先を解決できません: %w", err)
	}
	return filepath.Clean(dir), nil
}
