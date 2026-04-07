# Step 10: Observability — 「動いているか」ではなく「どう壊れているか」を見る

## 目的

「動いているか」ではなく「どう壊れているか」を見る入口を学ぶ。
ログだけに頼る運用から脱却し、メトリクスによる可観測性（Observability）の第一歩を体験する。

## 学ぶこと

- Observability の 3 本柱：ログ、メトリクス、トレース
- Prometheus によるメトリクス収集の仕組み
- Grafana によるメトリクスの可視化
- Kubernetes の RBAC（ServiceAccount / ClusterRole）の実例
- Pod アノテーションによる scrape 対象の制御

### Observability の 3 本柱

```
                Observability
         ┌──────────┼──────────┐
         │          │          │
       ログ     メトリクス    トレース
     (Logs)    (Metrics)   (Traces)
         │          │          │
   「何が起きたか」 「今どうか」 「どこで遅いか」
    テキスト情報   数値の時系列  リクエストの流れ
         │          │          │
    例: stderr    例: CPU 使用率  例: API → DB → Cache
    Loki 等で収集  Prometheus    Tempo / Jaeger
```

このステップでは **メトリクス** に焦点を当てる。Prometheus がメトリクスを収集し、Grafana で可視化する構成を作る。

### Prometheus の仕組み（Pull モデル）

```
Prometheus ──(HTTP GET /metrics)──▶ Pod A
    │                                 │
    ├──(HTTP GET /metrics)──▶ Pod B   │  ← Pod が自分のメトリクスを公開
    │                                 │
    └── 収集したデータを時系列 DB に保存
              │
              ▼
         Grafana で可視化
```

Prometheus は **Pull モデル** でメトリクスを収集する。各 Pod がメトリクスエンドポイント（通常 `/metrics`）を公開し、Prometheus が定期的に取りに行く。

## ディレクトリ構成

```
step10-observability/
├── README.md
├── namespace.yaml
├── prometheus-rbac.yaml
├── prometheus-config.yaml
├── prometheus-deployment.yaml
└── grafana-deployment.yaml
```

## 各ファイルの解説

### namespace.yaml

監視系コンポーネントを `monitoring` namespace に分離する。アプリケーションとインフラを namespace で分けるのは一般的なプラクティス。

### prometheus-rbac.yaml

| リソース | 説明 |
|---|---|
| `ServiceAccount` | Prometheus Pod が Kubernetes API にアクセスするためのアカウント |
| `ClusterRole` | Pod、Node、Service、Endpoints の情報を取得する権限 |
| `ClusterRoleBinding` | ServiceAccount に ClusterRole を紐づける |

Prometheus は Kubernetes API を使って「どの Pod が存在するか」を動的に発見する（Service Discovery）。そのため API へのアクセス権限が必要。

### prometheus-config.yaml

| 設定 | 説明 |
|---|---|
| `scrape_interval: 15s` | 15 秒ごとにメトリクスを収集する |
| `kubernetes_sd_configs` | Kubernetes API から Pod 一覧を取得して scrape 対象を自動発見する |
| `relabel_configs` | `prometheus.io/scrape: "true"` アノテーションがある Pod だけを対象にする |

### prometheus-deployment.yaml

Prometheus 本体の Deployment と NodePort Service。ConfigMap をボリュームとしてマウントし、設定ファイルを渡している。

### grafana-deployment.yaml

Grafana の Deployment と NodePort Service。初期パスワードは環境変数で `admin` に設定している。

## 実行手順

```bash
# 1. monitoring namespace を作成する
kubectl apply -f namespace.yaml

# 2. RBAC リソースを作成する
kubectl apply -f prometheus-rbac.yaml

# 3. Prometheus の設定を作成する
kubectl apply -f prometheus-config.yaml

# 4. Prometheus をデプロイする
kubectl apply -f prometheus-deployment.yaml

# 5. Grafana をデプロイする
kubectl apply -f grafana-deployment.yaml

# 6. Pod が全て Running になるまで待つ
kubectl -n monitoring get pods -w

# 7. Prometheus UI にアクセスする
#    kind の場合、NodePort でアクセスするか port-forward を使う
kubectl -n monitoring port-forward svc/prometheus 9090:9090 &

#    ブラウザで http://localhost:9090 を開く

# 8. Grafana にアクセスする
kubectl -n monitoring port-forward svc/grafana 3000:3000 &

#    ブラウザで http://localhost:3000 を開く
#    ログイン: admin / admin
```

### Grafana で Prometheus データソースを追加する

1. Grafana にログインする（admin / admin）
2. 左メニュー → Connections → Data sources → Add data source
3. Prometheus を選択する
4. URL に `http://prometheus.monitoring.svc.cluster.local:9090` を入力する
5. 「Save & test」をクリックする

### scrape 対象の Pod を追加する

Prometheus が Pod のメトリクスを収集するには、Pod にアノテーションを追加する必要がある。

```bash
# simple-api の Deployment にアノテーションを追加する例
kubectl patch deployment probe-demo --type=merge -p '{
  "spec": {
    "template": {
      "metadata": {
        "annotations": {
          "prometheus.io/scrape": "true",
          "prometheus.io/port": "8080"
        }
      }
    }
  }
}'
```

## 確認方法

1. **Prometheus UI**: http://localhost:9090 → Status → Targets で scrape 対象が表示されること。
2. **Prometheus UI**: http://localhost:9090 → Graph で `up` を入力し Execute → scrape 対象の状態が確認できること。
3. **Grafana**: http://localhost:3000 でログインでき、Prometheus データソースの接続テストが成功すること。
4. **Grafana ダッシュボード**: Explore → Prometheus データソース → `container_cpu_usage_seconds_total` 等のメトリクスが表示されること。

```
# Prometheus Targets の期待出力例
Endpoint                    State   Labels
http://10.244.0.15:8080     UP      job="kubernetes-pods" ...
```

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| Prometheus の Targets が空 | scrape 対象の Pod にアノテーションがない | `prometheus.io/scrape: "true"` を Pod テンプレートに追加する |
| Grafana から Prometheus に接続できない | データソース URL が間違っている | `http://prometheus.monitoring.svc.cluster.local:9090` を使う（Grafana も monitoring namespace にいるなら `http://prometheus:9090` でも可） |
| ログだけで満足してしまう | ログは「何が起きたか」の記録。異常の兆候を事前にキャッチするにはメトリクスが必要 | メトリクスで「今どういう状態か」をリアルタイムに把握する習慣をつける |
| ダッシュボードを作って「監視した気」になる | ダッシュボードを見る人がいなければ意味がない | 閾値を超えたらアラートを飛ばす仕組み（Alertmanager）を入れる |
| RBAC エラーで Prometheus が Pod を発見できない | ServiceAccount や ClusterRoleBinding が正しく設定されていない | `kubectl logs -n monitoring` で Prometheus のログを確認し、権限エラーがないかチェックする |

## 本番だとどう変わるか

- **Alertmanager**: Prometheus と連携し、メトリクスが閾値を超えたら Slack や PagerDuty に通知する。ダッシュボードを眺めるだけでは障害に気づけない
- **Loki**: ログの収集・検索基盤。Grafana と統合してメトリクスとログを横断的に調査できる
- **Tempo / Jaeger**: 分散トレーシング。マイクロサービス間のリクエストの流れを可視化し、ボトルネックを特定する
- **OpenTelemetry Collector**: ログ・メトリクス・トレースを統一的に収集・変換・転送するコンポーネント。ベンダーロックインを避ける
- **Prometheus Operator / kube-prometheus-stack**: Helm チャートで Prometheus, Grafana, Alertmanager, node-exporter 等を一括デプロイ。本番ではゼロから構築せずこれを使うのが一般的
- **永続化**: Prometheus のデータはデフォルトでは Pod 内に保存される。本番では PVC を使うか、Thanos / Cortex で長期保存する

---

次のステップでは、これまで学んだリソースを組み合わせて **ミニアーキテクチャ** を構築する。複数のコンポーネントが連携する実践的な構成を体験する。Step 11 に進もう。
