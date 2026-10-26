# 道具と実行方法

## 実行環境

同梱スクリプトは Python 3.10 以上の標準ライブラリで動く。
まずインタープリターのバージョンを確認する。
macOS の `/usr/bin/python3` は要件を満たさない場合があるため、利用可能な Homebrew・mise・Codex 同梱ランタイムも確認する。
対応する Python がなければ実行前に必要な環境を報告する。
追加パッケージの自動インストールは不要。

下記の `PYTHON` は確認済みインタープリター、`SKILL` はこのスキルの絶対パスへ、`TOPIC` と `ROOT` は調査保存先へ置き換える。
これは実行例で、永続的なシェル設定やコマンド定義を作る指示ではない。

```sh
"$PYTHON" "$SKILL/scripts/fetch.py" "$TOPIC/sources" author-year-title https://example.org/article
"$PYTHON" "$SKILL/scripts/chain.py" find "paper title" --log "$TOPIC" --round 1
"$PYTHON" "$SKILL/scripts/chain.py" refs "10.example/doi" --log "$TOPIC" --round 1
"$PYTHON" "$SKILL/scripts/chain.py" cites "10.example/doi" --log "$TOPIC" --round 1
"$PYTHON" "$SKILL/scripts/chain.py" author "author name" --log "$TOPIC" --round 1
"$PYTHON" "$SKILL/scripts/report.py" "$TOPIC" --check
"$PYTHON" "$SKILL/scripts/report.py" "$TOPIC"
"$PYTHON" "$SKILL/scripts/digest.py" "$TOPIC" --check
"$PYTHON" "$SKILL/scripts/digest.py" "$TOPIC"
"$PYTHON" "$SKILL/scripts/dashboard.py" "$ROOT"
"$PYTHON" "$SKILL/scripts/watch.py" add "$TOPIC" https://example.org/feed.xml
"$PYTHON" "$SKILL/scripts/watch.py" list "$TOPIC"
"$PYTHON" "$SKILL/scripts/watch.py" check "$TOPIC" --days 7
```

OpenAlex の候補は題名・年・DOI・所属を確認する。
新しい論文や同姓同名の著者では欠落・混同があるため、0件を不存在の証拠にしない。
API キー・課金設定を追加せず、利用制限に当たれば公式ページや参考文献に切り替える。

## 検索記録

検索ツールが実際に返した各ヒットを、検索語ごとに次の形で標準入力へ渡す。

```json
[{"query":"実際の検索語","method":"実際の検索ツール","hits":[{"title":"題名","url":"https://example.org/article"}]}]
```

```sh
"$PYTHON" "$SKILL/scripts/digest.py" add "$TOPIC" --round 0 --by japanese
```

空振りは `hits: []` として残す。
`by` で検索方針または実際の担当を分ける。
分担先が返した記録は再入力で書き直さず、その JSON を渡す。
`chain.py --log` の検索は自動記録されるため二重登録しない。

## 取得と任意コマンド

| 資料 | 同梱の処理 | 任意コマンドがない場合 |
| --- | --- | --- |
| HTML・テキスト | 公開先を検査して取得し、本文を保存 | 標準ライブラリで処理する |
| PDF | PDF と `pdftotext -layout` の文字起こし | `pdftotext` 不在の理由を返す |
| スキャン PDF | `pdftoppm` と `tesseract` で OCR | OCR 道具不在・品質不足は失敗として残す |
| YouTube | `yt-dlp` で字幕だけ取得 | コマンド不在の理由を返す |
| HTML の整形 | `ax` があれば取得済み HTML を変換 | 組み込みの本文抽出を使う |

OCR はインストール済みの日本語言語データを検出する。
`IMOZURU_OCR_LANG` は必要な言語データがある場合に指定できる。
取得名は著者・年・短い題、または arXiv 番号から作り、資料中の文字列をそのままシェルへ埋め込まない。
`fetch.py` は成功したものだけを保存し、失敗時は既存の原文を保持する。
本文長による検査は真偽・学術的品質の保証ではないため、内容を実際に読む。
`IMOZURU_ALLOW_PRIVATE` による宛先制限解除はユーザーの明示判断なしに行わない。

## 配信と持ち出し

静的な閲覧は `dashboard.py <research_dir>` だけで行える。
配信を依頼された場合は `dashboard.py <research_dir> --serve 8765` が `127.0.0.1` だけで待ち受け、合言葉付き URL を返す。
固定する場合は16文字以上の `IMOZURU_SERVE_TOKEN` を使い、トークンを外部検索や公開ログへ出さない。
スマートフォンからの利用にはユーザーの Tailscale 等の中継環境が別途必要。
中継の設定・インストール・バックグラウンド起動は依頼された範囲でのみ行う。

`dashboard.py <research_dir> --bundle <dir>` は原文を除き、ダッシュボードとレポート類をコピーする。
出力先とその親ディレクトリにシンボリックリンクがあると拒否するため、macOS では `/tmp` や `/var` の別名を避けて実パスを指定する。
引用は含まれるため、公開してよい束とは扱わない。
この移植版では感想フォームの既定の送り先を空にしている。
利用者が `IMOZURU_FEEDBACK_FORM` へ自身の Google フォームの事前入力 URL を設定した場合だけフォームを表示する。
エージェントからのフォーム送信は利用者の明示依頼がある場合だけ行う。
