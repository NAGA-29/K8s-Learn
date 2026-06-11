# simple-api

Go + Echo フレームワークで作成したシンプルな API サーバーです。
Kubernetes 学習用のバックエンドアプリケーションとして使用します。

## 概要

- ヘルスチェック、メッセージ返却、訪問カウンター、メトリクス公開を行う軽量 API
- ポート 8080 で起動
- カウンター機能 (`/api/count`) のみ Redis を使用する。Redis がない環境でも他のエンドポイントは動作する

## エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/health` | ヘルスチェック (`{"status": "ok"}`)。Redis には依存しない |
| GET | `/api/message` | メッセージとタイムスタンプを返却 |
| GET | `/api/count` | Redis の INCR で訪問回数をカウントして返却。Redis に接続できない場合は 503 を返す |
| GET | `/metrics` | Prometheus 形式のメトリクス (リクエスト数、goroutine 数、稼働時間) |

## 環境変数

| 変数 | デフォルト | 説明 |
|------|-----------|------|
| `REDIS_HOST` | `redis` | Redis のホスト名 (`/api/count` で使用) |
| `REDIS_PORT` | `6379` | Redis のポート番号 |

## ローカルでの実行

```bash
go run main.go
```

http://localhost:8080/health でアクセス確認できます。
`/api/count` を試す場合はローカルで Redis を起動し、`REDIS_HOST=localhost` を設定してください。

## Docker ビルド・実行

```bash
docker build -t simple-api .
docker run -p 8080:8080 simple-api
```
