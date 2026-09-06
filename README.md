# claat-skill

`manual.md` を lint してから、Google Codelabs の `claat` で HTML に変換する Go CLI です。

## スキルの導入

このリポジトリの手順書作成・検証・HTML生成を行うには、3つのスキルを導入してください。

```console
npx skills add nogikun/claat
```

すべて一括で導入したい場合、下記コマンドを実行してください  
`npx skills add nogikun/claat --skill claat-creator --skill claat-build --skill claat-writer`

## 前提

- Go 1.24+
- PATH 上の `claat`

`claat` が未導入なら、検証済みのバージョンを指定してインストールします。

```console
go install github.com/googlecodelabs/tools/claat@v0.0.0-20240220115335-873fe39d02dc
```

## 使い方

入力規約だけを検査します。

```console
go run ./cmd/claat-tools lint path/to/manual.md
```

検査に通った場合だけ HTML を生成します。

```console
go run ./cmd/claat-tools build -output output path/to/manual.md
```

公開済みバージョンは、リポジトリから直接実行できます。

```console
go run github.com/nogikun/claat/cmd/claat-tools@v0.1.0 lint manual.md
```

詳しい入力規約、lint ルール、テスト方針は [docs/claat-tools-design.md](docs/claat-tools-design.md) を参照してください。
