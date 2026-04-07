# simple-api

Go + Echo フレームワークで作成したシンプルな API サーバーです。
Kubernetes 学習用のバックエンドアプリケーションとして使用します。

## 概要

- ヘルスチェック、メッセージ返却、メトリクス公開を行う軽量 API
- ポート 8080 で起動

## エンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | `/health` | ヘルスチェック (`{"status": "ok"}`) |
| GET | `/api/message` | メッセージとタイムスタンプを返却 |
| GET | `/metrics` | Prometheus 形式のメトリクス (リクエスト数、goroutine 数、稼働時間) |

## ローカルでの実行

```bash
go run main.go
```

http://localhost:8080/health でアクセス確認できます。

## Docker ビルド・実行

```bash
docker build -t simple-api .
docker run -p 8080:8080 simple-api
```
