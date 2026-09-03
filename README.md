# Arca

Arcaは、用途と技術スタックを対話形式で選び、新しい開発プロジェクトをセットアップするCLIです。
macOS Apple Siliconのみ対応しています。

## 📦 インストール

### Homebrew

```shell
brew tap ishiguro-junya/arca https://github.com/ishiguro-junya/arca
brew install ishiguro-junya/arca/arca
```

### 直接インストール

```shell
curl -fsSL https://raw.githubusercontent.com/ishiguro-junya/arca/main/scripts/install.sh | sh
```

`ARCA_INSTALL_DIR`を指定すると、既定の`$HOME/.local/bin`以外へインストールできます。

## 🚀 使い方

```shell
arca
```

引数なしで起動すると、App、Backend、CLI、Infraを複数選択できるウィザードを開始します。
AppではWebまたはDesktop、CLIではGoまたはNode.jsを選択できます。
確認画面では生成、プレビュー、キャンセルを選択できます。
プレビューではファイル作成、依存関係の取得、外部コマンド実行を行いません。

### ウィザードの選択肢

- AppのWebはReact + ViteまたはNext.js、DesktopはTauriと対象プラットフォームを選択します。
- BackendはFastAPI、CLIはGo + CobraまたはNode.js + Commander、Infraは選択したTerraformプロバイダーを設定します。
- AppではUI/UX、Localization、State、Form、Validate、Icon、Fontを任意で追加できます。
- Next.js、Backend、Go CLIでは対応するORMを任意で追加できます。
- コード品質とコードセキュリティの設定は選択不要で生成されます。

```shell
arca version
arca version --json
arca update --check
arca update --check --json
arca update --yes
```

直接インストールしたバイナリはチェックサム検証後に自己更新します。
Homebrewで管理されている場合は自己更新せず、`brew upgrade ishiguro-junya/arca/arca`を案内します。

## 🧰 生成プロジェクト

生成プロジェクトでは`mise.toml`へ完全なツール版を記録し、`mise.lock`は生成しません。
すべての操作はmiseタスクへ定義し、`package.json`へ`scripts`を追加しません。
Node.js依存関係には公開後24時間の待機期間と、依存関係ビルドの許可制を適用します。
生成プロジェクトにはGitHub Actions、CI/CD、デプロイ設定を追加しません。

## 🛠️ 開発

```shell
mise install
mise run fmt-check
mise run test
mise run lint
mise run lint-md
mise run build
mise run release-check
```

`release-check`はGoReleaserのスナップショットを生成して配布設定を検証します。

## 🧭 AI Guideline

See [AGENTS.md](./AGENTS.md) for the AI Guidelines.  
We recommend installing AGENTS.md globally via a symbolic link.  

```shell
# Codex
ln -sfn "$(pwd)/AGENTS.md" ~/.codex/AGENTS.md

# Claude Code
ln -sfn "$(pwd)/AGENTS.md" ~/.claude/CLAUDE.md
```

## 📚 Documents

- [my-stack](./docs/my-stack.md)

## 🔄 リリース

`main`ブランチからReleaseワークフローを手動実行し、`v1.2.3`または`v1.2.3-alpha.1`形式の版を指定します。
ワークフローは検査、注釈付きタグ、GitHub Release、チェックサム、インストーラーを生成します。
公開後はmacOS Apple Silicon上でRelease、チェックサム、直接インストール、版表示、更新確認を検証します。
Release後は確定チェックサムを反映するHomebrew Formula更新プルリクエストを作成します。
Formulaはプルリクエストの作成前に`brew audit`、インストール、版表示を検証します。

## 🔗 References

- [front-end-frameworks](https://2025.stateofjs.com/en-US/libraries/front-end-frameworks/)
- [skills.sh](https://www.skills.sh/)

## ⚖️ Licence

[MIT License](LICENSE)
