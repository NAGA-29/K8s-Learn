# Step 09: HPA -- メトリクスに応じた自動スケーリング

## 目的

メトリクスに応じた自動スケーリングの入口を学ぶ。CPU 使用率が上がったら Pod を自動で増やし、下がったら減らす。この仕組みを HPA（Horizontal Pod Autoscaler）で体験する。

---

## 学ぶこと

- HPA の仕組みと動作フロー
- metrics-server の役割と導入方法
- CPU ベースのスケーリング設定
- `resources.requests` が HPA に必須である理由
- 負荷テストによるスケールアウト・スケールインの観察

### HPA の動作フロー

```
metrics-server ──(Pod の CPU/メモリ使用量を収集)──▶ Kubernetes API
                                                        │
HPA コントローラ ◀──(定期的にメトリクスを取得)──────────┘
  │
  ├── 現在の使用率 > target → レプリカ数を増やす（スケールアウト）
  └── 現在の使用率 < target → レプリカ数を減らす（スケールイン）
```

### hpa.yaml の設定値

| 項目 | 値 | 説明 |
|---|---|---|
| `minReplicas` | 1 | 最小 Pod 数。負荷がゼロでも 1 つは維持する |
| `maxReplicas` | 5 | 最大 Pod 数。これ以上は増えない |
| `averageUtilization` | 50 | 全 Pod の平均 CPU 使用率が 50% を超えたらスケールアウト |

---

## ディレクトリ構成

```
step09-hpa/
├── README.md          # このファイル
├── deployment.yaml    # resources.requests 付きの Deployment
└── hpa.yaml           # HPA の定義
```

---

## 前提条件

- Step 08 の `resources` 設定（requests / limits）を理解していること
- simple-api イメージがビルド済みであること

まだの場合はプロジェクトルートで以下を実行してください。

```bash
make build-images && make load-images
```

---

## 実行手順

### 1. metrics-server を導入する（kind 用）

HPA はメトリクスを元に判断するため、metrics-server が必須です。kind クラスタにはデフォルトで入っていないので手動で導入します。

```bash
# metrics-server をインストールする
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# kind 環境では kubelet の証明書検証をスキップする必要がある
kubectl patch -n kube-system deployment metrics-server \
  --type=json \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'

# metrics-server が起動するまで待つ（1〜2分）
kubectl -n kube-system rollout status deployment/metrics-server

# 動作確認（ノードのメトリクスが取れれば OK）
kubectl top nodes
```

### 2. Deployment と HPA を作成する

```bash
# Deployment を作成する
kubectl apply -f deployment.yaml

# Pod が Running になるまで待つ
kubectl get pods -l app=hpa-demo -w

# HPA を作成する
kubectl apply -f hpa.yaml

# HPA の状態を確認する（TARGETS が <unknown> でなくなるまで待つ）
kubectl get hpa hpa-demo -w
```

### 3. 負荷をかける

別のターミナルを開いて、busybox から連続リクエストを送ります。

```bash
# Service が無い場合は先に作成する
kubectl expose deployment hpa-demo --port=8080 --target-port=8080

# 負荷生成用の Pod を起動する
kubectl run -i --tty load-generator --rm --image=busybox --restart=Never -- \
  /bin/sh -c "while sleep 0.01; do wget -q -O- http://hpa-demo:8080/health; done"
```

### 4. スケールアウトを観察する

元のターミナルで HPA の変化を観察します。

```bash
# HPA の状態をリアルタイムで監視する
kubectl get hpa hpa-demo --watch
```

TARGETS 列の CPU 使用率が 50% を超えると、REPLICAS が増加していくのが確認できます。

### 5. 負荷を止めてスケールインを確認する

負荷生成用のターミナルで Ctrl+C を押して停止します（`--rm` オプションにより Pod は自動削除されます）。

```bash
# スケールインを観察する（5〜10分かかる）
kubectl get hpa hpa-demo --watch
```

---

## 確認方法

以下の全てが確認できれば、このステップは完了です。

1. `kubectl get hpa` の TARGETS 列に CPU 使用率が表示されること（`<unknown>` ではないこと）
2. 負荷をかけた後、REPLICAS 数が 1 から増加すること（最大 5）
3. 負荷を止めた後、しばらくして REPLICAS が 1 に戻ること

### 期待する出力例

負荷中:
```
NAME       REFERENCE             TARGETS    MINPODS   MAXPODS   REPLICAS   AGE
hpa-demo   Deployment/hpa-demo   120%/50%   1         5         3          5m
```

負荷停止後:
```
NAME       REFERENCE             TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
hpa-demo   Deployment/hpa-demo   2%/50%    1         5         1          15m
```

---

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| TARGETS が `<unknown>/50%` のまま | metrics-server が未導入またはまだ起動していない | metrics-server を導入し、`kubectl top nodes` で動作確認する |
| TARGETS が `<unknown>/50%` のまま（metrics-server は動いている） | Deployment に `resources.requests.cpu` が設定されていない | requests がないと HPA は使用率を計算できない。必ず設定する |
| スケールインが遅い | HPA のデフォルトでは安定化ウィンドウが 5 分ある | 意図的な仕様。本番では急なスケールインによるサービス断を防ぐため |
| 負荷をかけても Pod が増えない | Service が作成されていない、または負荷の宛先が間違っている | `kubectl expose` で Service を作成し、正しいエンドポイントに負荷をかける |
| metrics-server が CrashLoopBackOff になる | kind 環境で `--kubelet-insecure-tls` パッチを適用していない | 上記の patch コマンドを実行する |

---

## 本番だとどう変わるか

| 項目 | kind（この教材） | 本番環境 |
|---|---|---|
| **メトリクス** | CPU 使用率のみ | カスタムメトリクス（RPS、レイテンシ、キュー長など） |
| **スケーリングツール** | HPA のみ | HPA + KEDA + VPA の組み合わせ |
| **ノードスケール** | なし（固定） | Cluster Autoscaler / Karpenter でノード自体も自動追加 |
| **スケーリングポリシー** | デフォルト | `behavior` フィールドでスケールアウト/スケールインの速度を細かく制御 |
| **メトリクスソース** | metrics-server | Prometheus Adapter、Datadog Cluster Agent など |

- **KEDA（Kubernetes Event-Driven Autoscaler）**: SQS のキュー長、Kafka の lag、Prometheus のメトリクスなど外部イベントソースに基づくスケーリングに対応
- **VPA（Vertical Pod Autoscaler）**: Pod 数ではなく、Pod あたりの CPU/メモリ割り当てを自動調整する
- **Cluster Autoscaler / Karpenter**: Pod が増えてノードが足りなくなった場合、ノード自体を自動追加する

---

## まとめ

- HPA は CPU 使用率などのメトリクスに基づいて Pod 数を自動調整する
- metrics-server が必須であり、kind では手動導入が必要
- `resources.requests` の設定がないと HPA は動作しない
- スケールインには安定化ウィンドウ（デフォルト5分）があり、意図的に遅い

---

## 次のステップ

Step 10 では「動いている」ではなく「どう壊れているか」を見るための **Observability（可観測性）** を学びます。Prometheus と Grafana を使ってメトリクスの収集と可視化を体験しましょう。
