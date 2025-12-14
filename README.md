# 安心手帳 バックエンド

安心手帳のバックエンド

## Requirements

最低限[Docker](https://www.docker.com/)と[Docker Compose](https://docs.docker.com/compose/)が必要です。
[Compose Watch](https://docs.docker.com/compose/file-watch/)を使うため、Docker Compose のバージョンは 2.22 以上にしてください。

Linter, Formatter には[golangci-lint](https://golangci-lint.run/)を使っています。
VSCode を使用する場合は`.vscode/settings.json`で Linter の設定を行ってください

```json
{
  "go.lintTool": "golangci-lint"
}
```

## Directory structure

[Organizing a Go module - The Go Programming Language](https://go.dev/doc/modules/layout#server-project) などを参考にしています。

```bash
$ tree | manual-explain
.
├── main.go # エントリーポイント
├── infrastructure # 統合テスト用に公開するDBやDI等の設定
├── internal # ロジック (結合テストに公開する必要がないもの)
│   ├── handler # APIハンドラ
│   ├── repository # DBアクセス
│   └── services # 外部サービス, 複雑なビジネスロジック
└── integration_tests # 結合テスト
```

特に重要なものは以下の通りです。

### `main.go`

アプリケーションのエントリーポイントを配置します。

**Tips**: 複数のエントリーポイントを実装する場合は、`cmd` ディレクトリを作成し、各エントリーポイントを `cmd/{app name}/main.go` に書くと見通しが良くなります。

### `infrastructure/database`

DB スキーマの定義、DB の初期化、マイグレーションを行っています。

マイグレーションツールは[pressly/goose](https://github.com/pressly/goose)を使っています。

### `internal/`

アプリケーション本体のロジックを配置します。
主に 2 つのパッケージに分かれています。

- `handler/`: ルーティング
  - 飛んできたリクエストを裁いてレスポンスを生成する
  - DB アクセスは`repository/`で実装したメソッドを呼び出す
  - **Tips**: リクエストのバリデーションがしたい場合は ↓ のどちらかを使うと良い
    - [go-playground/validator](https://github.com/go-playground/validator): タグベースのバリデーション
    - [go-ozzo/ozzo-validation](https://github.com/go-ozzo/ozzo-validation): コードベースのバリデーション
- `repository/`: ストレージ操作
  - DB や外部ストレージなどのストレージにアクセスする
    - 引数のバリデーションは`handler/`に任せる

**Tips**: `internal`パッケージは他モジュールから参照されません（参考: [Go 1.4 Release Notes](https://go.dev/doc/go1.4#internalpackages)）。
依存性注入や外部ライブラリの初期化のみを`core/`や`pkg/`で公開し、アプリケーションのロジックは`internal/`に閉じることで、後述の`integration_tests/go.mod`などの外部モジュールからの参照を最小限にすることができ、開発の効率を上げることができます。

### `integration_tests/`

結合テストを配置します。
API エンドポイントに対してリクエストを送り、レスポンスを検証します。
短期開発段階では時間があれば書く程度で良いですが、長期開発に向けては書いておくと良いでしょう。

```go
package integration_tests

import (
  "testing"
  "gotest.tools/v3/assert"
)

func TestUser(t *testing.T) {
  t.Run("get users", func(t *testing.T) {
    t.Run("success", func(t *testing.T) {
      t.Parallel()
      rec := doRequest(t, "GET", "/api/v1/users", "")

      expectedStatus := `200 OK`
      expectedBody := `[{"id":"[UUID]","name":"test","email":"test@example.com"}]`
      assert.Equal(t, rec.Result().Status, expectedStatus)
      assert.Equal(t, escapeSnapshot(t, rec.Body.String()), expectedBody)
    })
  })
}
```

**Tips**: DB コンテナの立ち上げには[ory/dockertest](https://github.com/ory/dockertest)を使っています。

**Tips**: アサーションには[gotest.tools](https://github.com/gotestyourself/gotest.tools)を使っています。
`go test -update`を実行することで、`expectedXXX`のスナップショットを更新することができます（参考: [gotest.tools を使う - 詩と創作・思索のひろば](https://motemen.hatenablog.com/entry/2022/03/gotest-tools)）。

外部サービス（traQ, Twitter など）へのアクセスが発生する場合は Test Doubles でアクセスを置き換えると良いでしょう。

## Tasks

開発に用いるコマンド一覧

> [!TIP] > `xc` を使うことでこれらのコマンドを簡単に実行できます。
> 詳細は以下のページをご覧ください。
>
> - [xc](https://xcfile.dev)
> - [Markdown ベースの Go 製タスクランナー「xc」のススメ](https://zenn.dev/trap/articles/af32614c07214d)
>
> ```bash
> go install github.com/joerdav/xc/cmd/xc@latest
> ```

### Build

アプリをビルドします。

```sh

CMD=server
go mod download
go build -o ./bin/${CMD} ./main.go
```

### Dev

ホットリロードの開発環境を構築します。

```sh
docker compose watch
```

API、DB、DB 管理画面が起動します。
各コンテナが起動したら、以下の URL にアクセスすることができます。
Compose Watch により、ソースコードの変更を検知して自動で再起動します。

- <http://localhost:8080/> (API)
- <http://localhost:8081/> (DB の管理画面)

### Test

全てのテストを実行します。

Requires: Test-Unit, Test-Integration

RunDeps: async

```sh
echo hello
```

### Test-Unit

単体テストを実行します。

```sh
go test -v -cover -race -shuffle=on ./internal/...
```

### Test-Integration

結合テストを実行します。

```sh
[ ! -e ./go.work ] && go work init . ./integration_tests
go test -v -cover -race -shuffle=on ./integration_tests/...
```

### Test-Integration:Update

結合テストのスナップショットを更新します。

```sh
[ ! -e ./go.work ] && go work init . ./integration_tests
go test -v -cover -race -shuffle=on ./integration_tests/... -update
```

### Lint

Linter (golangci-lint) を実行します。

```sh
golangci-lint run --timeout=5m --fix ./...
```

## Improvements

長期開発に向けた改善点をいくつか挙げておきます。

- ドメインを書く (`internal/domain/`など)
  - 現在は簡単のために API スキーマと DB スキーマのみを書きこれらを直接やり取りしている
  - 本来はアプリの仕様や概念をドメインとして書き、スキーマの変換にはドメインを経由させるべき
- クライアント API スキーマを共通化させる
  - OpenAPI や GraphQL を使い、そこから Go のファイルを生成する
- 単体テスト・結合テストのカバレッジを上げる
  - カバレッジの可視化には[Codecov](https://codecov.io)(traP だと主流)や[Coveralls](https://coveralls.io)が便利
- ログの出力を整備する
  - ロギングライブラリは好みに合ったものを使うと良い

## Deploy

GitHub Actions でビルドした Docker image を Heroku や NeoShowcase のような PaaS にデプロイするための Job を用意しています。

詳しくは [./.github/workflows/image.yaml](./.github/workflows/image.yaml) を参照してください。

## Troubleshooting

### Docker image のビルドが失敗する

不要なファイルが Docker image に混入するのを防ぐために `.dockerignore` を allowlist 方式にしています。
ビルドに必要なファイルやディレクトリを追加した場合、 `.dockerignore` も編集してください。
