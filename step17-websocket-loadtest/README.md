# Step 17: WebSocket負荷試験 - リアルタイム接続のボトルネック体験

## 目的

リアルタイム接続のボトルネックを体験する。接続数を段階的に増やし、CPU、メモリ、goroutine数などのメトリクスを計測して、スケーリングの限界点と判断基準を学ぶ。

## 学ぶこと

- WebSocket負荷試験の方法（k6を使用）
- 計測すべきメトリクスとその意味
- 接続数増加に伴うリソース消費の変化
- Pod数・Redisの有無がパフォーマンスに与える影響
- Pod再起動時の影響分析

## 試験項目

1. 接続数を100 → 500 → 1000へ段階的に増やす
2. broadcast頻度を上げる
3. Pod数を変える（1 → 3 → 5）
4. Redisあり/なしを比較する

## 計測すべきメトリクス

| メトリクス | 確認方法 | なぜ重要か |
|-----------|---------|-----------|
| CPU使用率 | `kubectl top pods` | broadcastのループ処理でCPUを消費する |
| Memory使用率 | `kubectl top pods` | 各接続がバッファメモリを消費する |
| goroutine数 | `/metrics`エンドポイント | 接続ごとにread/writeの2 goroutineが生成される |
| アクティブ接続数 | `/metrics`エンドポイント | 現在の同時接続数を把握する |
| メッセージ配信遅延 | k6のレスポンスタイム | 負荷が高まると配信遅延が増加する |
| Pod再起動時の影響 | 接続切断数、再接続時間 | 本番での障害影響を見積もる |

## ディレクトリ構成

```
step17-websocket-loadtest/
└── README.md    # このファイル（手順と分析ガイド）
```

負荷試験スクリプト（`apps/loadtest/`配下の`http-test.js`、`ws-test.js`）を使用する。

## 前提条件

- Step 15, Step 16が完了していること
- `realtime-api`が`mini-app` namespaceで動作していること
- k6がインストールされていること
- `apps/loadtest/`配下にテストスクリプトが存在すること

## 実行手順

### 1. k6のインストール確認

```bash
k6 version
```

インストールされていない場合：

```bash
# macOS
brew install k6

# Linux (snap)
sudo snap install k6
```

### 2. port-forwardでアクセス可能にする

```bash
kubectl port-forward -n mini-app svc/realtime-api 8080:8080
```

このターミナルは開いたままにしておく。

### 3. HTTP healthcheckテスト

別ターミナルで実行する：

```bash
cd apps/loadtest
k6 run http-test.js
```

まずHTTPエンドポイントが正常に応答することを確認する。

### 4. WebSocketテスト（接続数を段階的に増やす）

#### 100接続

```bash
K6_WS_URL=ws://localhost:8080/ws k6 run --vus 100 --duration 1m ws-test.js
```

#### 500接続

```bash
K6_WS_URL=ws://localhost:8080/ws k6 run --vus 500 --duration 1m ws-test.js
```

#### 1000接続

```bash
K6_WS_URL=ws://localhost:8080/ws k6 run --vus 1000 --duration 1m ws-test.js
```

### 5. 別ターミナルでメトリクス監視

負荷試験中に別ターミナルで以下を実行して、リアルタイムにリソース消費を監視する：

```bash
# metricsエンドポイントの監視（2秒間隔）
watch -n 2 'curl -s http://localhost:8080/metrics'
```

```bash
# Pod単位のCPU/Memory監視
watch -n 5 'kubectl top pods -n mini-app'
```

### 6. Pod数を変えてテスト

```bash
# 1 Pod
kubectl scale deployment/realtime-api -n mini-app --replicas=1
# → 上記の接続数テストを繰り返す

# 3 Pod
kubectl scale deployment/realtime-api -n mini-app --replicas=3
# → 上記の接続数テストを繰り返す

# 5 Pod
kubectl scale deployment/realtime-api -n mini-app --replicas=5
# → 上記の接続数テストを繰り返す
```

### 7. Redisあり/なしの比較

```bash
# Redisなし
kubectl apply -f ../step16-websocket-scale/deployment-no-redis.yaml
# → 接続数テストを実行し、メトリクスを記録

# Redisあり
kubectl apply -f ../step16-websocket-scale/deployment-with-redis.yaml
# → 同じ接続数テストを実行し、メトリクスを記録
```

## 計測結果サンプル（テーブル形式）

| 接続数 | Pod数 | Redis | CPU | Memory | goroutines | 配信遅延 |
|--------|-------|-------|-----|--------|------------|----------|
| 100 | 1 | なし | 15% | 24Mi | 205 | <5ms |
| 500 | 1 | なし | 65% | 89Mi | 1005 | 15ms |
| 1000 | 1 | なし | 95% | 156Mi | 2005 | 120ms |
| 1000 | 3 | あり | 35% | 62Mi | 670 | 25ms |

**注意:** これはサンプル値。実際の値は環境（CPU性能、メモリ量、ネットワーク帯域）に依存する。自分の環境で計測して記録すること。

## ボトルネック分析

### CPU

- 1 Podに接続が集中するとCPU使用率が急上昇する
- broadcast処理は全接続に対してループするため、接続数に比例してCPU消費が増える
- broadcast頻度が高いとCPU消費がさらに急増する

### メモリ

- goroutine数がほぼ「接続数 x 2」（read goroutine + write goroutine）になる
- 各接続のバッファがメモリを消費する
- 1000接続で150Mi前後、接続数に比例して増加する

### スケールアウト効果

- Redisありで3 Podに分散すると、1 Podあたりの負荷が約1/3になる
- ただしRedis自体がボトルネックになる可能性がある（大量メッセージ時）
- Pod数を増やしすぎるとRedisへのPub/Sub接続が増え、Redis側の負荷が上がる

### 配信遅延

- 低負荷時は5ms未満
- 高負荷時（1000接続、1 Pod）は100ms以上に悪化
- Redisを経由すると、低負荷でもRedisのRTT分（数ms）が加算される

## Pod再起動テスト

負荷試験中にPodを意図的に削除して、影響を計測する：

```bash
# 負荷中にPodを1つ削除
kubectl delete pod -n mini-app $(kubectl get pods -n mini-app -l app=realtime-api -o jsonpath='{.items[0].metadata.name}')
```

計測すべきポイント：

- 削除されたPodに接続していたクライアントの切断数
- 再接続にかかる時間
- 切断から再接続までの間にロストしたメッセージ数
- 残りのPodへの負荷の変化（接続が再分配される）

## 確認方法

以下の観点で計測結果を評価する：

```bash
# 負荷試験の結果サマリ（k6が出力する）
# - ws_connecting: 接続確立にかかった時間
# - ws_msgs_received: 受信メッセージ数
# - ws_msgs_sent: 送信メッセージ数

# Pod単位のリソース使用状況
kubectl top pods -n mini-app

# metricsエンドポイントの値
curl -s http://localhost:8080/metrics
```

接続数を増やした際にCPU/Memory/goroutine数がどう変化するかを記録し、自環境でのスケーリング限界点を把握する。

## よくある失敗

### 速いコードを書けば十分と思う

どれだけコードを最適化しても、1 Podで扱える接続数には物理的な上限がある。負荷試験でその上限を事前に把握し、スケール戦略を立てることが重要。

### 接続維持のメモリコストを軽視する

各WebSocket接続はgoroutine、バッファ、接続メタデータ分のメモリを消費する。1接続あたりのコストは小さくても、1万接続になると無視できない量になる。メモリのlimitsを超えるとOOM Killされる。

### 再接続ロジックが未考慮

Pod再起動やスケールダウンで接続は切断される。クライアント側に再接続ロジックがなければ、ユーザーは手動でページをリロードするしかない。再接続時のメッセージ復旧も考慮が必要。

## 本番だとどう変わるか

- **専用のWebSocketゲートウェイ**: 接続管理に特化したレイヤーを設け、アプリケーションサーバーと分離する
- **接続数ベースのオートスケール**: HPAのカスタムメトリクスとしてアクティブ接続数を使い、自動スケールする
- **メッセージキューイング**: 切断中のメッセージをキューに保存し、再接続時に配信する（at-least-once保証）
- **段階的な負荷試験**: ステージング環境で本番相当の負荷を定期的にかけ、キャパシティプランニングを行う
- **接続数の上限設定**: 1 Podあたりの最大接続数を設定し、超過時は新規接続を拒否する
- **モニタリングとアラート**: 接続数、メモリ使用率、配信遅延に対してアラートを設定する

---

## カリキュラム完了

ここまでの全17ステップで学んだことのまとめ：

- **K8sの基本リソース**: Pod, Deployment, Service, Ingress, Namespace等の基本操作
- **設定管理とストレージ**: ConfigMap, Secret, PersistentVolume
- **ヘルスチェックとオートスケール**: Liveness/Readiness Probe, HPA
- **可観測性の基礎**: ログ、メトリクス、Prometheus, Grafana
- **マイクロサービス構成**: 複数サービスの連携、Service Discovery
- **障害対応**: Pod障害、リソース不足、デバッグ手法
- **アーキテクチャレビュー**: 設計の見直し、改善ポイントの洗い出し
- **EKS移行の視点**: ローカルK8sからクラウドへの移行で変わること
- **WebSocketのリアルタイム処理**: 基礎、スケール問題、負荷試験

## 次のステップ（自学）の提案

このカリキュラムで基礎を固めた上で、以下のトピックに進むことを推奨する：

- **Service Mesh (Istio / Linkerd)**: サービス間通信の可視化、トラフィック制御、mTLS
- **GitOps (ArgoCD / Flux)**: Gitリポジトリを信頼の源とした宣言的デプロイ
- **セキュリティ強化 (OPA / Kyverno, Falco)**: ポリシーエンジンによるガバナンス、ランタイムセキュリティ
- **CI/CD (GitHub Actions + ArgoCD)**: コードプッシュからデプロイまでの自動化パイプライン
- **マルチクラスタ管理**: 複数クラスタの統合管理、フェデレーション
