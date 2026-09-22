# claat-skill

`manual.md` を lint してから、Google Codelabs の `claat` で HTML に変換する Go CLI です。

## スキルの導入

手順書の執筆から HTML 生成までを 1 つのスキルで行います。

```console
npx skills add nogikun/claat
```

## 前提

- Go 1.24+
- PATH 上の `claat`（Google Codelabs）

CLI 本体は `skills/claat-creator/cmd/claat-tools/` に置いてあり、`npx skills add` でスキルごとコピーされます。**別途インストールは不要です。** 標準ライブラリしか使わないので `go.mod` も要りません。

`claat` だけは別に入れてください。未導入なら、検証済みのバージョンを指定します。

```console
go install github.com/googlecodelabs/tools/claat@v0.0.0-20240220115335-873fe39d02dc
```

## 使い方

入力規約だけを検査します。

```console
go run skills/claat-creator/cmd/claat-tools/main.go lint path/to/manual.md
```

検査に通った場合だけ HTML を生成します。

```console
go run skills/claat-creator/cmd/claat-tools/main.go build -output output path/to/manual.md
```

スキル経由で使う場合は、導入先（`.agents/skills/claat-creator/cmd/claat-tools/main.go`）を同じように `go run` に渡します。

## 開発

コミットテンプレートを使う場合は 1 回だけ設定します。

```console
git config --local commit.template .commit-template
```

詳しい入力規約、lint ルール、テスト方針は [docs/claat-tools-design.md](docs/claat-tools-design.md) を参照してください。
