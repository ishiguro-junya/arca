package project

import (
	"fmt"
	"slices"
	"strings"
)

func viteFiles(spec Spec) []File {
	imports := []string{
		`import React from "react";`,
		`import ReactDOM from "react-dom/client";`,
		`import App from "./App";`,
		`import "./index.css";`,
	}
	if spec.Features.FontEnabled {
		if slices.Contains(spec.Features.Fonts, "inter") {
			imports = append(imports, `import "@fontsource-variable/inter/wght.css";`)
		}
		if slices.Contains(spec.Features.Fonts, "noto-sans-jp") {
			imports = append(imports, `import "@fontsource-variable/noto-sans-jp/wght.css";`)
		}
	}
	if spec.Features.Localization {
		imports = append(imports, `import "./i18n";`)
	}
	mainSource := strings.Join(imports, "\n") + `

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode><App /></React.StrictMode>,
);`
	viteConfig := `import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: { outDir: "dist/app" },
});`
	if spec.AppKind == AppDesktop {
		viteConfig = `import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: { port: 1420, strictPort: true },
  build: { outDir: "dist/app" },
});`
	}
	files := []File{
		textFile("index.html", fmt.Sprintf(`<!doctype html>
<html lang="%s">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>%s</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>`, defaultLocale(spec), spec.Name)),
		textFile("vite.config.ts", viteConfig),
		textFile("tsconfig.json", tsconfigApp()),
		textFile("src/main.tsx", mainSource),
		textFile("src/App.tsx", fmt.Sprintf(`export default function App() {
  return (
    <main className="min-h-screen p-8 font-sans">
      <h1 className="text-3xl font-bold">%s</h1>
    </main>
  );
}`, spec.Name)),
		textFile("src/index.css", appCSS(spec)),
	}
	if spec.Features.Localization {
		files = append(files, viteLocalizationFiles(spec)...)
	}
	return files
}

func nextFiles(spec Spec) []File {
	files := []File{
		textFile("tsconfig.json", tsconfigNext()),
		textFile("next-env.d.ts", `/// <reference types="next" />
/// <reference types="next/image-types/global" />`),
		textFile("postcss.config.mjs", `export default { plugins: { "@tailwindcss/postcss": {} } };`),
		textFile("next.config.ts", nextConfig(spec)),
		textFile("src/app/globals.css", appCSS(spec)),
		textFile("src/app/layout.tsx", nextLayout(spec)),
		textFile("src/app/page.tsx", fmt.Sprintf(`export default function Home() {
  return (
    <main className="min-h-screen p-8 font-sans">
      <h1 className="text-3xl font-bold">%s</h1>
    </main>
  );
}`, spec.Name)),
	}
	if spec.Features.Localization {
		files = append(files,
			textFile("messages/ja.json", `{"title":"こんにちは"}`),
			textFile("messages/en.json", `{"title":"Hello"}`),
			textFile("src/i18n/request.ts", fmt.Sprintf(`import { getRequestConfig } from "next-intl/server";

export default getRequestConfig(async () => ({
  locale: %q,
  messages: (await import(%q)).default,
}));`, defaultLocale(spec), "../../messages/"+defaultLocale(spec)+".json")),
		)
	}
	return files
}

func nextConfig(spec Spec) string {
	if !spec.Features.Localization {
		return `import type { NextConfig } from "next";

const nextConfig: NextConfig = {};
export default nextConfig;`
	}
	return `import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");
export default withNextIntl({});`
}

func nextLayout(spec Spec) string {
	imports := []string{`import "./globals.css";`}
	classes := []string{}
	fontDefinitions := []string{}
	if spec.Features.FontEnabled {
		fontImports := []string{}
		if slices.Contains(spec.Features.Fonts, "inter") {
			fontImports = append(fontImports, "Inter")
			fontDefinitions = append(fontDefinitions, `const inter = Inter({ subsets: ["latin"], variable: "--font-inter" });`)
			classes = append(classes, "inter.variable")
		}
		if slices.Contains(spec.Features.Fonts, "noto-sans-jp") {
			fontImports = append(fontImports, "Noto_Sans_JP")
			fontDefinitions = append(fontDefinitions, `const notoSansJP = Noto_Sans_JP({ preload: false, variable: "--font-noto-sans-jp" });`)
			classes = append(classes, "notoSansJP.variable")
		}
		imports = append(imports, fmt.Sprintf(`import { %s } from "next/font/google";`, strings.Join(fontImports, ", ")))
	}
	providerImport := ""
	body := "{children}"
	if spec.Features.Localization {
		providerImport = `import { NextIntlClientProvider } from "next-intl";`
		body = "<NextIntlClientProvider>{children}</NextIntlClientProvider>"
	}
	classExpression := `""`
	if len(classes) > 0 {
		classExpression = "[" + strings.Join(classes, ", ") + `].join(" ")`
	}
	return fmt.Sprintf(`%s
%s

%s

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang=%q className={%s}>
      <body>%s</body>
    </html>
  );
}`, strings.Join(imports, "\n"), providerImport, strings.Join(fontDefinitions, "\n"), defaultLocale(spec), classExpression, body)
}

func appCSS(spec Spec) string {
	fonts := []string{"sans-serif"}
	if spec.Features.FontEnabled {
		fonts = nil
		if slices.Contains(spec.Features.Fonts, "inter") {
			if spec.IsNext() {
				fonts = append(fonts, "var(--font-inter)")
			} else {
				fonts = append(fonts, `"Inter Variable"`)
			}
		}
		if slices.Contains(spec.Features.Fonts, "noto-sans-jp") {
			if spec.IsNext() {
				fonts = append(fonts, "var(--font-noto-sans-jp)")
			} else {
				fonts = append(fonts, `"Noto Sans JP Variable"`)
			}
		}
		fonts = append(fonts, "sans-serif")
	}
	return fmt.Sprintf(`@import "tailwindcss";

@theme inline {
  --font-sans: %s;
}

body {
  margin: 0;
}`, strings.Join(fonts, ", "))
}

func viteLocalizationFiles(spec Spec) []File {
	return []File{
		textFile("src/i18n.ts", fmt.Sprintf(`import i18n from "i18next";
import { initReactI18next } from "react-i18next";
import en from "./locales/en.json";
import ja from "./locales/ja.json";

void i18n.use(initReactI18next).init({
  resources: { en: { translation: en }, ja: { translation: ja } },
  lng: %q,
  fallbackLng: "ja",
  interpolation: { escapeValue: false },
});`, defaultLocale(spec))),
		textFile("src/locales/ja.json", `{"title":"こんにちは"}`),
		textFile("src/locales/en.json", `{"title":"Hello"}`),
	}
}

func defaultLocale(spec Spec) string {
	if spec.ReadmeLanguage == ReadmeEnglish {
		return "en"
	}
	return "ja"
}

func tsconfigApp() string {
	return `{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["DOM", "DOM.Iterable", "ES2022"],
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "jsx": "react-jsx",
    "strict": true,
    "noEmit": true,
    "skipLibCheck": true,
    "resolveJsonModule": true,
    "baseUrl": ".",
    "paths": { "@/*": ["./src/*"] }
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "vite.config.ts", "next.config.ts"]
}`
}

func tsconfigNext() string {
	return `{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["DOM", "DOM.Iterable", "ES2022"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "baseUrl": ".",
    "paths": { "@/*": ["./src/*"] }
  },
  "include": ["next-env.d.ts", ".next/types/**/*.ts", "**/*.ts", "**/*.tsx"],
  "exclude": ["node_modules", "cli"]
}`
}

func tauriFiles(spec Spec) []File {
	crateName := strings.ReplaceAll(spec.Name, "-", "_")
	return []File{
		textFile("src-tauri/Cargo.toml", fmt.Sprintf(`[package]
name = %q
version = "0.1.0"
edition = "2024"

[lib]
name = %q
crate-type = ["staticlib", "cdylib", "rlib"]

[build-dependencies]
tauri-build = { version = "2", features = [] }

[dependencies]
tauri = { version = "2", features = [] }`, crateName, crateName+"_lib")),
		textFile("src-tauri/build.rs", `fn main() { tauri_build::build() }`),
		textFile("src-tauri/src/lib.rs", `#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .run(tauri::generate_context!())
        .expect("Tauriアプリケーションの実行に失敗しました");
}`),
		textFile("src-tauri/src/main.rs", fmt.Sprintf(`fn main() {
    %s_lib::run();
}`, crateName)),
		textFile("src-tauri/tauri.conf.json", fmt.Sprintf(`{
  "$schema": "https://schema.tauri.app/config/2",
  "productName": %q,
  "version": "0.1.0",
  "identifier": "com.ishiguro.%s",
  "build": {
    "beforeDevCommand": "mise run app:dev-web",
    "beforeBuildCommand": "mise run app:build-web",
    "devUrl": "http://localhost:1420",
    "frontendDist": "../dist/app"
  },
  "app": {
    "windows": [{ "title": %q, "width": 1000, "height": 700 }]
  },
  "bundle": { "active": false }
}`, spec.Name, strings.ReplaceAll(spec.Name, "-", "."), spec.Name)),
		textFile("src-tauri/capabilities/default.json", `{
  "identifier": "default",
  "description": "Default capability",
  "windows": ["main"],
  "permissions": ["core:default"]
}`),
	}
}
