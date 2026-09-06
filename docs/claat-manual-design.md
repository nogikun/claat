# `claat-manual` 設計（Go版）

## 結論

`manual.md` の構文を検査してから、既存の `claat export` に渡す薄い Go CLI を作る。

利用者の入口は次の 2 つだけにする。

```console
# 検査だけ
go run github.com/nogikun/claat/cmd/claat-manual@v0.1.0 lint manual.md

# 検査に通った場合だけ HTML 化
go run github.com/nogikun/claat/cmd/claat-manual@v0.1.0 build -output output manual.md
```

`claat` の Markdown パーサーや HTML レンダラーは再実装しない。Go 側は「入力規約の検査」と「`claat` の安全な起動」だけを担当する。

## 背景と観察

`C:\Users\takah\Documents\git\try-claat` を確認した結果は次のとおり。

- `dev/template.md` は YAML front matter ではなく、ファイル先頭の `key: value` メタデータを使う。
- メタデータの後に `#` のページタイトルが 1 つある。
- `##` 見出しが手順の境界で、直後に `Duration: H:M:SS` がある。
- `###` 以降、コードフェンス、リンク、画像、表、チェックリストなどは手順本文としてそのまま扱う。
- 現在の変換処理は Taskfile の `claat export -o output <file>` だけである。
- `output/` は生成物であり、入力 Markdown と分離されている。
- 本体リポジトリには、まだ CLI 実装も `docs/` もない。

したがって、独自の変換フォーマット、Web UI、設定ファイル、常駐サーバーを先に作る必要はない。

## 対象範囲

### 対象にするもの

- `manual.md` の構造検査
- template.md にあるメタデータの必須性と値の検査
- 手順ごとの `Duration` の検査
- ローカル画像ファイルの存在検査
- 検査成功後の `claat export` 実行
- CI から扱える終了コードと、行番号付きの標準エラー出力

### 対象にしないもの

- Markdown 全般の文体・日本語校正
- 外部 URL へのアクセス確認
- Markdown の自動修正
- `claat` バイナリの自動ダウンロード
- `claat` の HTML 出力の再実装
- ファイル監視、プレビューサーバー、複数入力の一括ビルド

これらは入力規約の検査と HTML 化の最短経路を成立させた後に、実際の失敗例が出てから追加する。

## linter の位置づけ

linter は必要であり、`claat-manual lint` を初版の必須機能とする。ただし、次の 2 種類を混ぜない。

| 種類 | 目的 | 初版 |
| --- | --- | --- |
| claat 規約 linter | HTML 化に失敗する入力を事前に止める | Go CLI に内蔵 |
| 一般 Markdown linter | 見出しレベル、空行、リスト記法などの文体・形式を整える | 別途 CI で任意に実行 |

一般 Markdown linter だけでは、template.md 固有の次の関係を検査できない。

- `id`、`status`、`environments` の値が claat 用として妥当か
- `##` の直後に、その手順の `Duration` があるか
- metadata の重複や未知キーがないか
- 相対画像が入力ファイルから解決できるか
- コードフェンス内の見出しを誤って手順として数えていないか

したがって、`lint` は外部パッケージではなく Go 標準ライブラリの小さな規約 linter として実装する。外部 Markdown linter は、このツールの成功条件とは独立した品質チェックとして扱う。

## CLI

### `lint`

```console
go run github.com/nogikun/claat/cmd/claat-manual@v0.1.0 lint path/to/manual.md
```

- エラーを `path:line:column: error CODE message` 形式で標準エラー出力に出す。
- エラーがなければ終了コード `0`。
- 1 件でもエラーがあれば終了コード `1`。
- `claat` は起動しない。

### `build`

```console
go run github.com/nogikun/claat/cmd/claat-manual@v0.1.0 build -output output path/to/manual.md
```

処理順は固定する。

1. 入力ファイルを UTF-8 として読む。
2. `lint` と同じ検査を同じプロセス内で実行する。
3. エラーがあれば、出力ディレクトリを変更せず終了コード `1`。
4. `PATH` から `claat` を探す。見つからなければインストール方法を示して終了コード `2`。
5. `exec.Command(claat, "export", "-o", output, input)` で起動する。
6. `claat` の終了コードと標準出力・標準エラー出力を利用者へ返す。

`-output` の既定値は `output` とする。既存ファイルを Go 側で削除しない。生成物を残したくない場合の掃除は利用者または CI の責任にする。

### 既存環境での使い方

`go run` は Go CLI の取得・コンパイル・実行を担当し、`claat` は別途 PATH に用意する。

```console
go install github.com/googlecodelabs/tools/claat@v0.0.0-20240220115335-873fe39d02dc
go run github.com/nogikun/claat/cmd/claat-manual@v0.1.0 lint dev/manual.md
go run github.com/nogikun/claat/cmd/claat-manual@v0.1.0 build -output output dev/manual.md
```

一度インストールして PATH から使いたい場合は次の形にする。

```console
go install github.com/nogikun/claat/cmd/claat-manual@v0.1.0
claat-manual lint dev/manual.md
```

ローカル開発では `go run ./cmd/claat-manual lint dev/manual.md` とする。`go run package@version` は現在のプロジェクトの `go.mod` と分離して実行できるため、CI や利用者向けの例ではタグ付きバージョンを使う。`@latest` は試用時だけにする。

`claat` の自動取得を初版に含めないのは、OS ごとの実行ファイル配布とバージョン固定をこの小さなラッパーの責務に持ち込まないためである。公式リポジトリはアーカイブ済みなので、CI では `@latest` ではなく検証済みの配布物を固定する。

## 入力規約と lint ルール

検査対象は「template.md の記述規約」と「HTML 化前に失敗させると有益な構造エラー」に限定する。Markdown の意味を完全に解釈する linter にはしない。

### メタデータ

ファイル先頭から最初の `#` 見出しまでをメタデータ領域とする。空行は許可するが、メタデータでも空行でもない行があればエラーにする。

必須キーは template.md に合わせて次の 7 個とする。

| キー | 検査 |
| --- | --- |
| `summary` | 空でないこと |
| `id` | 空でなく、`[A-Za-z0-9][A-Za-z0-9_-]*` に一致すること。パス区切りや `..` は不可 |
| `categories` | 空でないカンマ区切りリストであること |
| `environments` | `Web` または `Kiosk` のカンマ区切りリストであること |
| `status` | `Draft`、`Published`、`Deprecated`、`Hidden` のいずれか 1 つであること |
| `feedback link` | 空でないこと。URL の疎通確認はしない |
| `analytics account` | 空でないこと。ID の形式は `claat` に委ねる |

同じキーの重複はエラーにする。初版では template.md にないキーもエラーにして、`summry` のような typo を見逃さない。将来、`tags` や `authors` などを使う必要が出た時点で明示的に許可する。

### 見出し

- ページタイトルとして `# タイトル` をちょうど 1 つ置く。
- ページタイトルはメタデータの後、最初の `##` より前に置く。
- `## 手順タイトル` を 1 個以上置く。
- 各 `##` は 1 手順として扱う。
- `###` と `####` は手順内の小見出しとして許可する。
- `#####`、`######` は初版ではエラーにする。
- ページタイトルの後に別の `#` が現れたらエラーにする。
- コードフェンス内の `#` は見出しとして数えない。

### 手順時間

各 `##` の後ろに、最初の非空行として次の形式を置く。

```text
Duration: H:M:SS
```

`H`、`M`、`SS` は非負整数、`M` と `SS` は `0` から `59` の範囲とする。`try-claat` の `0:5:00` と template.md の `0:05:00` の両方を受け付ける。

手順本文が `Duration` だけで終わる場合はエラーにする。最後の「最後に」手順を `0:00:00` にするかどうかは、公式ガイドの推奨に留まるため lint エラーにはしない。

### 本文とアセット

- 手順本文は空白だけでなければよい。操作方法や成功条件の文体までは検査しない。
- `![alt](relative/path.png)` のような相対ローカル画像だけ、入力ファイルから解決して存在を確認する。
- `https://`、`data:`、絶対パスの画像は存在検査しない。
- 通常のリンクは構文だけを壊さず通し、外部サイトへリクエストしない。
- コードフェンスは開始・終了の対応を検査する。対応しないフェンスはエラーにする。

### エラーコード

エラーコードは CI の除外や将来のエディター連携に使えるよう、短く固定する。

| コード | 意味 |
| --- | --- |
| `META001` | 必須メタデータがない |
| `META002` | 重複または未許可のメタデータキー |
| `META003` | メタデータ値が不正 |
| `DOC001` | ページタイトルがない、複数ある、または位置が不正 |
| `STEP001` | 手順見出しがない、または見出しレベルが不正 |
| `STEP002` | `Duration` がない、位置が不正、または値が不正 |
| `STEP003` | 手順本文が空 |
| `MD001` | コードフェンスが閉じていない |
| `ASSET001` | 相対画像ファイルが存在しない |

出力例:

```text
dev/manual.md:12:1: error STEP002 expected `Duration: H:M:SS` immediately after the step heading
dev/manual.md:48:5: error ASSET001 local image not found: images/setup.png
```

## 提供するプログラム

### コマンドの責務

| コマンド | 役割 | `claat` 起動 |
| --- | --- | --- |
| `lint <file>` | 入力規約だけを検査し、診断を出す | しない |
| `build <file> [-output dir]` | lint 成功後に HTML を生成する | 成功時だけ 1 回 |

`help` は組み込みの usage で提供する。`serve`、`init`、`fmt`、`watch` は初版に入れない。

### 使うもの

```text
Go 1.24+
  ├─ 標準ライブラリ: flag, os, os/exec, path/filepath, regexp, strings, unicode/utf8
  └─ 外部依存: なし
```

Go 1.24.3 で動作確認する。`go run package@version` 自体は Go 1.17 以降で使えるが、初版の最低対応バージョンは実装時に CI で決める。

Typer のような CLI フレームワークは使わない。2 コマンドなら標準 `flag.FlagSet` と短い usage で十分であり、外部依存なしで `go run` の初回取得を速くできる。コマンドが増えて help の階層化や completion が本当に必要になった時だけ Cobra を検討する。

## 実装方針

### 最小の構成

```text
go.mod
cmd/
└── claat-manual/
    ├── main.go
    └── main_test.go
```

`cmd/claat-manual/main.go` に subcommand の dispatch、`flag.FlagSet`、小さな行スキャナー、診断出力、`claat` 起動を置く。複数の parser 層、プラグイン機構、設定ファイル、依存性注入は作らない。ファイルが大きくなった時だけ `manual.go` へ分割する。

行スキャナーは次の状態だけ持つ。

- メタデータ領域か本文領域か
- コードフェンス内かどうか
- 現在のページタイトルと手順
- 手順ごとの `Duration` と本文の有無

Markdown 全体の AST は不要である。`claat` が最終的な Markdown 解釈を行うため、ラッパーが同じ AST を持つと二重実装になる。

### `go.mod` の要点

```go
module github.com/nogikun/claat

go 1.24
```

このリポジトリのモジュールパスは `github.com/nogikun/claat` とする。main package を `cmd/claat-manual` に置くことで、Google の `claat` と衝突しない `claat-manual` バイナリを提供できる。利用者は `go run github.com/nogikun/claat/cmd/claat-manual@v0.1.0` と書ける。

### `claat` の起動

- `exec.LookPath("claat")` で PATH 上の実体を探す。
- `exec.Command` を使い、シェルを経由しない。
- 入力パスと出力パスは引数配列で渡し、シェル文字列を組み立てない。
- lint エラー時は `os/exec` を一度も呼ばない。
- `claat` が失敗した場合はその終了コードを返す。

これで Windows PowerShell と POSIX シェルの引用符差異をラッパー側が持たずに済む。

## テストと受け入れ条件

フレームワークを増やさず、次の最小テストを `go test ./...` で実行する。

1. `try-claat` の `dev/sample.md` が lint を通る。
2. 必須キー欠落、重複キー、不正な `status`、危険な `id` を検出する。
3. `#` タイトルの欠落・重複、`##` 手順の欠落を検出する。
4. `Duration` の欠落、位置違い、形式違い、範囲外を検出する。
5. コードフェンス内の見出しを誤検出しない。未閉鎖フェンスは検出する。
6. 相対画像の存在・欠落を検出する。
7. lint エラー時に fake `claat` が呼ばれない。
8. build 成功時に `claat export -o output manual.md` 相当の引数で一度だけ呼ばれる。
9. `claat` 未導入時に、入力エラーと区別できる終了コード `2` になる。

完成条件は次のとおり。

```console
go run ./cmd/claat-manual lint ..\try-claat\dev\sample.md
# exit 0

go run ./cmd/claat-manual build -output output ..\try-claat\dev\try-claat-guide.md
# lint 成功後に claat が実行され、output/<id>/ が生成される
```

## 採用しなかった案

### Python + Typer で作る

Typer は help や subcommand を整理しやすいが、このツールでは Go の `go run` と標準ライブラリで同じ目的を満たせる。Python 環境を増やす理由ができるまでは採用しない。

### `go run` 実行時に `claat` をダウンロードする

OS、CPU、配布物の署名・ハッシュ、キャッシュ、更新方針まで必要になる。初版は PATH 上の `claat` を使い、CI 側でバージョンを固定する。

### `--fix` と設定ファイル

入力を書き換えると、生成前のレビュー可能性が下がる。規約が変わるまで自動修正や YAML/TOML 設定は持たない。

### 一般的な Markdown linter を組み込む

既存の Markdown linter と責務が重なる。まずは `claat` が解釈する境界だけを検査し、文体ルールは必要になった時点で外部 linter の設定として追加する。

## 参照

- try-claat のローカル検証リポジトリ: `C:\Users\takah\Documents\git\try-claat`
- [Google Codelabs Tools](https://github.com/googlecodelabs/tools)
- [Codelab Formatting Guide](https://github.com/googlecodelabs/tools/blob/main/FORMAT-GUIDE.md)
- [claat README](https://github.com/googlecodelabs/tools/blob/main/claat/README.md)
