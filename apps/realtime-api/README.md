# Realtime API - WebSocket サーバー

Go で実装されたリアルタイムチャット WebSocket サーバー。接続中の全クライアントにメッセージをブロードキャストする。

## エンドポイント

| パス | 説明 |
|------|------|
| `GET /` | ブラウザ用 WebSocket クライアント (index.html) |
| `GET /health` | ヘルスチェック |
| `GET /ws` | WebSocket 接続エンドポイント |
| `GET /metrics` | 接続数・メッセージ数・goroutine 数を JSON で返す |

## メッセージフォーマット

```json
{
  "type": "message",
  "data": "Hello!",
  "from": "abc12345",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## ローカル実行

```bash
go run .
# ブラウザで http://localhost:8080 を開く
```

## Docker で実行

```bash
docker build -t realtime-api .
docker run -p 8080:8080 realtime-api
```

## Redis Pub/Sub モード

環境変数 `REDIS_URL` を設定すると、Redis Pub/Sub を使って複数 Pod 間でメッセージを配信できる。設定しない場合はインメモリのブロードキャストのみ。

```bash
# Redis を使う場合
export REDIS_URL=redis://localhost:6379
go run .
```

## 環境変数

| 変数 | デフォルト | 説明 |
|------|-----------|------|
| `PORT` | `8080` | サーバーのリッスンポート |
| `REDIS_URL` | (なし) | Redis 接続 URL。設定するとPub/Subモードが有効になる |
