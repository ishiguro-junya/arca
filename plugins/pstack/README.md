# pstack for Codex

Lauren Tan（poteto）の [pstack](https://github.com/cursor/plugins/tree/main/pstack) を Codex 向けに移植した非公式プラグインです。
調査、設計、実装、検証、レビュー、PR 管理を、54 スキルと 23 の手順集で支援します。
作者や Cursor、OpenAI による公式版ではありません。

## 導入

このリポジトリを取得し、そのルートで実行します。

```sh
codex plugin marketplace add .
codex plugin add pstack@arca
```

マーケットプレイス名は `arca`、プラグイン名は `pstack` です。
Codex App ではマーケットプレイスからプラグインを選び、必要ならアプリを再起動して新しいチャットで確認してください。
このリポジトリへのファイル配置だけでは、普段の Codex にインストールされません。
確認には `codex plugin list --available --marketplace arca --json` を使用できます。
CLI のプラグイン機能がない場合は、対応する Codex App から導入してください。

## 使い方

Codex のスキル一覧から `pstack:poteto-mode` を選ぶか、スキル名を明示して依頼します。

```text
pstack の poteto-mode を使って、この不具合を再現し、修正して検証してください。
```

個別の調査なら `how`、設計なら `architect`、レビューなら `interrogate` を指定できます。
原典から残している `/name` 表記はスキル名の略記であり、Codex の同名スラッシュコマンドを保証するものではありません。
内部のスキル呼び出しは同梱ファイルを読む方式なので、別途入っている同名スキルを上書きしません。
起動時フックや常時適用モードはありません。
`setup-pstack` 以外のスキルは原典どおり明示呼び出し用で、poteto-mode から必要な手順を参照します。

## モデルと利用環境

設定なしでは、すべての役割が親チャットのモデルと推論設定を継承します。
`setup-pstack` で利用可能なモデルを確認し、役割別のモデル・推論量・レビュー人数を変更できます。
保存先は既存の CODEX_HOME 配下の `pstack/models.md` で、CODEX_HOME が未設定なら `~/.codex/pstack/models.md` です。
Codex のグローバルなモデル設定や AGENTS.md は変更しません。

並列実行は、現在の Codex が提供するサブエージェント機能と利用先の指示に従います。
利用できない場合は工程を順次実行し、独立レビューを実施したとは報告しません。
別モデルによる検証が明示的な完了条件なら、その条件は未達として扱います。
実行手段、履歴の範囲、継続実行、権限の扱いは [Codex runtime contract](references/codex-runtime.md) に定義しています。

通常のスキル文書の利用に、補助ツールのインストールは不要です。
補助ツールの対応環境は macOS／Linux のローカル環境で、Windows ネイティブ実行は検証対象外です。

| 使用する機能 | 必要なもの |
| --- | --- |
| PR 状態の取得と監視 | Bun、Git、認証済みの GitHub CLI |
| 長期作業の状態保存 | Bun |
| 計画書の構造チェック | Node.js または Bun |
| 作業ツリーの監査 | Bash、Git、jq、Python 3、PR 状態を取得する GitHub CLI |
| ローカル会話履歴の照合 | Python 3 |
| 実行記録 | Bash |

Bun と JavaScript 依存関係の検証対象は [package.json](skills/poteto-mode/scripts/package.json) とロックファイルに記録しています。
CLI 補助ツールの初回起動では、固定された依存関係をユーザーキャッシュへ取得します。
保存先は XDG_CACHE_HOME または `~/.cache` 配下の `pstack/tools/` で、PSTACK_CACHE_DIR により検証用の別ディレクトリを指定できます。
インストール時スクリプトは実行せず、プラグイン本体には依存関係を書き込みません。
継続作業の記録は Codex のユーザーディレクトリ内の `pstack/runs/<run-id>/` に分離し、再開時は記録に残した同じ場所を使います。

## PR 監視と継続実行

一度だけ状態を確認する依頼では `watch-pr --status-only` を使います。
取得成功時は終了コード 0、取得失敗時は再試行や待機をせず終了コード 7 を返します。
継続監視が依頼されている場合に限り、通常の監視処理を使用します。
GitHub がブランチの更新待ちやマージ可否の判定未完了を返した場合は、`READY` にせず待機します。
JSON の待機理由は `merge-state` で、`mergeStateStatus` に `BEHIND` または `UNKNOWN` を返します。
期限を指定した監視では、期限まで解消しなければ終了コード 5 を返します。
キュー監視では、更新待ちや判定未完了が解消した後も、その PR のマージを待ちます。
後で再開する依頼は利用可能な Codex 自動化機能を使い、機能がない場合は予約できていないことと再開方法を報告します。
プラグインは独自の常駐プロセスや cron を登録しません。
PR の作成、コメント投稿、マージ、強制 push は、利用者の依頼と利用先の権限に従います。

## 対応外

- Cursor のクラウドマシンと Custom Modes。
- Grok Bot の専用 Webhook UI を作成する `make-bot-ui`。
- Benny の Cursor Automations サービス。
- Cursor 専用機能を再現する独自バックエンド。

通常の並列作業はローカルのサブエージェントと作業ツリーに置き換えます。
既存の Bugbot コメントは GitHub のレビューとして確認できますが、Cursor サービスの新規設定は行いません。
UI や CLI の検証には、対象プロジェクトで利用できるハーネス、コネクター、操作ツールを使用します。

## 検証

実行結果と未検証事項は [検証記録](docs/validation.md) にまとめています。
補助ツールのテストは、一時ディレクトリへ `plugins/pstack/` 全体をコピーし、その中の `skills/poteto-mode/scripts/` で実行できます。
計画書のテストは同梱の手順書も読み込みます。

```sh
bun install --frozen-lockfile --ignore-scripts
bun test
bun run typecheck
```

テストは一時リポジトリと偽の GitHub 応答を使用し、実際の PR を変更しません。
キャッシュの初回作成テストではパッケージ取得のためにネットワークを使用します。
スキルの実行シナリオは [評価シナリオ](evals/scenarios.md) を参照してください。
メタデータやリンクの検証だけで、実際のエージェント動作を検証済みとはしません。

## 上流の更新

移植元コミット、対象ファイルのハッシュ、除外範囲は [upstream.json](upstream.json) に記録しています。
更新時はそのコミットと更新先の公式ソースを比較し、スキル、参照、手順集、補助ツール、ライセンスの差分を確認します。
Codex 用の実行契約、モデル設定、明示呼び出し設定、履歴照合、キャッシュ、権限の扱いを維持しながら取り込みます。
追加・削除されたスキルと参照の整合を確認し、影響するテストと実行シナリオを再検証してから来歴とプラグインのバージョンを更新します。
この手順は手動で行い、自動同期ツールは同梱していません。

## 出典とライセンス

原典 pstack は Lauren Tan による MIT ライセンスです。
[LICENSE](LICENSE) の著作権表示を保持しています。
コード整理用の [deslop](references/deslop.md) は同じ公式リポジトリの cursor-team-kit から取り込んでおり、[そのライセンス](references/LICENSE-cursor-team-kit) を同梱しています。
原典の Cursor 向け導入ガイドは同梱せず、この README と Codex 用のヘルプを提供します。

## Skills

| スキル | 入口 |
| --- | --- |
| architect | [SKILL.md](skills/architect/SKILL.md) |
| arena | [SKILL.md](skills/arena/SKILL.md) |
| automate-me | [SKILL.md](skills/automate-me/SKILL.md) |
| benchmark-checklist | [SKILL.md](skills/benchmark-checklist/SKILL.md) |
| blast-radius | [SKILL.md](skills/blast-radius/SKILL.md) |
| bro | [SKILL.md](skills/bro/SKILL.md) |
| correct | [SKILL.md](skills/correct/SKILL.md) |
| create-formal-verification-skill | [SKILL.md](skills/create-formal-verification-skill/SKILL.md) |
| create-verification-skill | [SKILL.md](skills/create-verification-skill/SKILL.md) |
| figure-it-out | [SKILL.md](skills/figure-it-out/SKILL.md) |
| how | [SKILL.md](skills/how/SKILL.md) |
| in-your-own-words | [SKILL.md](skills/in-your-own-words/SKILL.md) |
| interrogate | [SKILL.md](skills/interrogate/SKILL.md) |
| ketchup | [SKILL.md](skills/ketchup/SKILL.md) |
| maintain-verification-skill | [SKILL.md](skills/maintain-verification-skill/SKILL.md) |
| no-comments | [SKILL.md](skills/no-comments/SKILL.md) |
| poteto-help | [SKILL.md](skills/poteto-help/SKILL.md) |
| poteto-mode | [SKILL.md](skills/poteto-mode/SKILL.md) |
| recall | [SKILL.md](skills/recall/SKILL.md) |
| reflect | [SKILL.md](skills/reflect/SKILL.md) |
| setup-pstack | [SKILL.md](skills/setup-pstack/SKILL.md) |
| sharpen | [SKILL.md](skills/sharpen/SKILL.md) |
| show-me-your-work | [SKILL.md](skills/show-me-your-work/SKILL.md) |
| swarm | [SKILL.md](skills/swarm/SKILL.md) |
| tdd | [SKILL.md](skills/tdd/SKILL.md) |
| teach | [SKILL.md](skills/teach/SKILL.md) |
| technical-writing | [SKILL.md](skills/technical-writing/SKILL.md) |
| typescript-best-practices | [SKILL.md](skills/typescript-best-practices/SKILL.md) |
| unslop | [SKILL.md](skills/unslop/SKILL.md) |
| why | [SKILL.md](skills/why/SKILL.md) |

24 の `principle-*` スキルは、[poteto-mode](skills/poteto-mode/SKILL.md) の原則一覧から参照します。
