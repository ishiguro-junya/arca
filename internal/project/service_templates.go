package project

import (
	"fmt"
	"slices"
	"strings"
)

func backendFiles(spec Spec, versions map[string]string) []File {
	module := strings.ReplaceAll(spec.Name, "-", "_")
	dependencies := []string{`"fastapi[standard]"`, `"httpx"`}
	if spec.Features.ORM == "sqlalchemy" {
		dependencies = append(dependencies, `"alembic"`, `"sqlalchemy"`)
	}
	slices.Sort(dependencies)
	pythonVersion := majorMinor(toolVersion(versions, "python"))
	return []File{
		textFile("backend/pyproject.toml", fmt.Sprintf(`[project]
name = %q
version = "0.1.0"
requires-python = ">=%s"
dependencies = [%s]

[dependency-groups]
dev = ["pytest", "ruff", "vulture"]

[tool.pytest.ini_options]
pythonpath = ["src"]`, spec.Name+"-backend", pythonVersion, strings.Join(dependencies, ", "))),
		textFile("backend/src/"+module+"/__init__.py", ""),
		textFile("backend/src/"+module+"/main.py", `from fastapi import FastAPI

app = FastAPI()


@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}`),
		textFile("backend/tests/test_health.py", fmt.Sprintf(`from fastapi.testclient import TestClient

from %s.main import app


def test_health() -> None:
    response = TestClient(app).get("/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}`, module)),
	}
}

func cliFiles(spec Spec, versions map[string]string) []File {
	if spec.HasGoCLI() {
		return []File{
			textFile("go.mod", fmt.Sprintf("module github.com/example/%s\n\ngo %s", spec.Name, majorMinor(toolVersion(versions, "go")))),
			textFile("cli/main.go", fmt.Sprintf(`package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

func main() {
    cmd := &cobra.Command{Use: %q, Short: %q}
    if err := cmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}`, spec.Name, spec.Name+" command line interface")),
		}
	}
	return []File{
		textFile("tsconfig.cli.json", `{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "outDir": "dist/cli",
    "rootDir": "cli",
    "types": ["node"]
  },
  "include": ["cli/**/*.ts"]
}`),
		textFile("cli/main.ts", fmt.Sprintf(`#!/usr/bin/env node
import { Command } from "commander";

new Command()
  .name(%q)
  .description(%q)
  .version("0.1.0")
  .parse();`, spec.Name, spec.Name+" command line interface")),
	}
}

func majorMinor(version string) string {
	parts := strings.Split(strings.TrimPrefix(version, "v"), ".")
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}

func infraFiles(spec Spec) []File {
	sources := map[string]string{
		"aws":        "hashicorp/aws",
		"azure":      "hashicorp/azurerm",
		"gcp":        "hashicorp/google",
		"cloudflare": "cloudflare/cloudflare",
		"vercel":     "vercel/vercel",
	}
	provider := spec.InfraProvider
	providerBlock := provider
	switch provider {
	case "azure":
		providerBlock = "azurerm"
	case "gcp":
		providerBlock = "google"
	}
	providerConfig := fmt.Sprintf("provider %q {}", providerBlock)
	if provider == "azure" {
		providerConfig = `provider "azurerm" {
  features {}
}`
	}
	return []File{textFile("infra/main.tf", fmt.Sprintf(`terraform {
  required_providers {
    %s = {
      source = %q
    }
  }
}

%s`, providerBlock, sources[provider], providerConfig))}
}

func knipTemplate(spec Spec) string {
	ignored := []string{
		"@textlint-ja/textlint-rule-preset-ai-writing",
		"knip",
		"markdownlint-cli2",
		"oxfmt",
		"oxlint",
		"textlint",
		"typescript",
	}
	entries := []string{}
	dependencies, devDependencies := map[string]string{}, map[string]string{}
	if spec.Has(UseApp) {
		if spec.IsNext() {
			entries = append(entries, "src/app/**/*.{ts,tsx}")
		} else {
			entries = append(entries, "src/main.tsx")
		}
	}
	if spec.HasNodeCLI() {
		entries = append(entries, "cli/main.ts")
		ignored = append(ignored, "tsx")
	}
	if spec.Has(UseApp) && spec.AppKind == AppDesktop {
		ignored = append(ignored, "@tauri-apps/api", "@tauri-apps/cli")
	}
	addFeatureDependencies(spec, dependencies, devDependencies)
	for name := range dependencies {
		ignored = append(ignored, name)
	}
	for name := range devDependencies {
		if name == "shadcn" || name == "@svgr/core" {
			ignored = append(ignored, name)
		}
	}
	slices.Sort(ignored)
	quoted := make([]string, len(ignored))
	for i, name := range ignored {
		quoted[i] = fmt.Sprintf("%q", name)
	}
	quotedEntries := make([]string, len(entries))
	for i, entry := range entries {
		quotedEntries[i] = fmt.Sprintf("%q", entry)
	}
	return fmt.Sprintf(`{"entry":[%s],"ignoreDependencies":[%s]}`, strings.Join(quotedEntries, ","), strings.Join(quoted, ","))
}
