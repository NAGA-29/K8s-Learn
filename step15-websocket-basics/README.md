# Step 15: WebSocket基礎 - Kubernetes上のリアルタイム通信

## 目的

Kubernetes上でリアルタイム通信を扱う最小構成を作る。WebSocketの基本概念を理解し、HTTPとの違いやIngressで必要な設定を学ぶ。

## 学ぶこと

- WebSocketの基本（双方向・長時間接続）
- HTTPとWebSocketの違い
- Ingressでの注意点（timeout設定）

## HTTPとWebSocketの違い（重要！）

| 項目 | HTTP | WebSocket |
|------|------|-----------|
| 接続 | リクエストごと | 長時間維持 |
| 方向 | クライアント→サーバー | 双方向 |
| ステートレス | はい | いいえ（接続状態あり） |
| スケール | Pod増やすだけ | 接続の再分配が必要 |
| Ingress設定 | デフォルトでOK | timeout延長必須 |

HTTPは「リクエスト→レスポンス」で完結する。WebSocketは一度接続を確立すると、サーバーからもクライアントからも自由にメッセージを送れる。チャット、通知、リアルタイムダッシュボードなどで使われる。

## なぜIngressのtimeoutが重要か

- デフォルトの`proxy-read-timeout`は60秒
- WebSocket接続は長時間維持する必要がある
- timeoutが短いと接続が勝手に切れる

`ingress.yaml`では3つのtimeoutを全て3600秒（1時間）に設定している:

```yaml
annotations:
  nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
  nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
  nginx.ingress.kubernetes.io/proxy-connect-timeout: "3600"
```

## ディレクトリ構成

```
step15-websocket-basics/
├── README.md          # このファイル
├── deployment.yaml    # realtime-api Deployment（1 Pod）
├── service.yaml       # ClusterIP Service
└── ingress.yaml       # WebSocket対応Ingress（timeout延長）
```

## 前提条件

- `realtime-api`イメージがビルド・ロード済みであること
- `mini-app` namespaceが存在すること
- NGINX Ingress Controllerが有効であること

## 実行手順

### 1. リソースのデプロイ

```bash
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml
```

### 2. hostsファイルの設定

```bash
# /etc/hosts に以下を追加
echo "127.0.0.1 ws.local" | sudo tee -a /etc/hosts
```

### 3. Podの起動確認

```bash
kubectl get pods -n mini-app -l app=realtime-api
```

`STATUS`が`Running`、`READY`が`1/1`になるまで待つ。

### 4. ブラウザでアクセス

`http://ws.local` にアクセスする。`realtime-api`の`index.html`が表示される。

### 5. メッセージ送受信の確認

テキスト入力欄にメッセージを入力して送信する。自分の画面にメッセージが表示されることを確認する。

### 6. broadcastの確認

別タブで`http://ws.local`を開く。片方のタブでメッセージを送信すると、もう片方のタブにもメッセージが表示されることを確認する。

## 確認方法

```bash
# Podが正常に動いていること
kubectl get pods -n mini-app -l app=realtime-api

# Ingressが設定されていること
kubectl get ingress -n mini-app realtime-ingress

# 接続数の確認（/metricsエンドポイント）
curl -s http://ws.local/metrics
```

複数タブでメッセージが同期されていれば成功。`/metrics`でアクティブ接続数が表示される。

## よくある失敗

### 単発リクエストの感覚で考える

WebSocketはHTTPと違い、接続を維持し続ける。1リクエスト1レスポンスの感覚で設計すると、接続管理ができない。

### 接続数・メモリ使用量を気にしない

各WebSocket接続はサーバー側にgoroutine（またはスレッド）とメモリを消費する。接続数が増えるとリソース消費が比例して増える。

### Ingressのtimeoutデフォルトで接続が切れる

60秒のデフォルトtimeoutのままだと、WebSocket接続が1分で切断される。`proxy-read-timeout`、`proxy-send-timeout`、`proxy-connect-timeout`の3つ全てを延長する必要がある。

## 本番だとどう変わるか

- **WebSocket専用Gateway**: APIゲートウェイとは別に、WebSocket専用のゲートウェイを設ける
- **専用Realtime基盤**: Ably、Pusher、Socket.ioクラスタなど、リアルタイム通信に特化した基盤を使う
- **接続数ベースのスケーリング**: CPU/Memoryだけでなく、接続数に基づいてPodをスケールする
- **heartbeat/ping-pong**: 接続の生存確認を定期的に行い、切断を検知する
- **再接続ロジック**: クライアント側に自動再接続の仕組みを実装する
- **ロードバランサーのWebSocketサポート**: ALB/NLBなど、WebSocketを正しく扱えるロードバランサーを選定する

---

次のステップでは、Step 16でPodを複数に増やしたときにWebSocketのbroadcastが揃わなくなる問題を体験する。ステートフルな接続をスケールすることの難しさと、Redis Pub/Subによる解決策を学ぶ。
