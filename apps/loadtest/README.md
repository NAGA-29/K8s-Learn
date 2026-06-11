# Load Test - k6 負荷テストスクリプト

[k6](https://k6.io/) を使った HTTP および WebSocket の負荷テストスクリプト。

## k6 のインストール

```bash
# macOS
brew install k6

# Ubuntu/Debian
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
  --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D68
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" \
  | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update && sudo apt-get install k6

# Docker
docker run --rm -i grafana/k6 run - <script.js
```

## テストの実行

### HTTP 負荷テスト

```bash
# デフォルト (localhost:8080)
k6 run http-test.js

# ターゲット URL を指定
k6 run -e BASE_URL=http://my-service:8080 http-test.js
```

### WebSocket 負荷テスト

```bash
# デフォルト (ws://localhost:8080/ws)
k6 run ws-test.js

# ターゲット URL を指定
k6 run -e WS_URL=ws://my-service:8080/ws ws-test.js

# 同時接続数を変更（デフォルト100）
k6 run -e VUS=500 ws-test.js
```

## 環境変数

| 変数 | デフォルト | 説明 |
|------|-----------|------|
| `BASE_URL` | `http://localhost:8080` | HTTP テストのターゲット URL |
| `WS_URL` | `ws://localhost:8080/ws` | WebSocket テストのターゲット URL |
| `VUS` | `100` | WebSocket テストの最大同時接続数 |

## 結果の見方

テスト完了後に k6 が出力するサマリーの主な指標:

- **http_req_duration**: リクエストのレイテンシ (p50, p90, p95, p99)
- **http_req_failed**: エラー率
- **ws_connect_time**: WebSocket 接続にかかった時間
- **ws_messages_received**: 受信したメッセージ総数
- **ws_send_errors**: 送信エラーの数
- **iterations**: 完了したイテレーション数
- **vus**: 同時仮想ユーザー数

### しきい値 (Thresholds)

| テスト | 指標 | 条件 |
|--------|------|------|
| HTTP | `http_req_duration` | p95 < 500ms |
| HTTP | `http_req_failed` | < 1% |
| WebSocket | `ws_connect_time` | p95 < 2000ms |
| WebSocket | `ws_send_errors` | < 10 回 |

しきい値を超えると k6 が非ゼロの終了コードを返すため、CI/CD パイプラインに組み込むことができる。
