# Step 08: Probes & Resources — 起動確認・生存確認・リソース制限

## 目的

起動確認・生存確認・リソース制限の基本を学ぶ。
Kubernetes がコンテナの「準備ができたか」「まだ生きているか」をどう判断するかを理解し、適切なリソース制限の設定方法を身につける。

## 学ぶこと

- readinessProbe / livenessProbe / startupProbe の違いと使い分け
- resources の requests と limits の意味
- Probe 失敗時に Kubernetes が取る挙動

### 3 種類の Probe の違い

```
Pod 起動
  │
  ▼
[startupProbe] ─── 成功するまで他の Probe は実行されない
  │                  失敗し続ける → コンテナ再起動
  │ 成功
  ▼
[readinessProbe] ── 成功 → Service のエンドポイントに追加（トラフィックが来る）
  │                  失敗 → エンドポイントから除外（トラフィックが来ない）
  │                  ※コンテナは再起動しない
  │
[livenessProbe] ─── 成功 → 何もしない
                     失敗 → コンテナを再起動
```

| Probe | 目的 | 失敗時の挙動 |
|---|---|---|
| `startupProbe` | 起動が完了したか確認 | コンテナを再起動 |
| `readinessProbe` | トラフィックを受ける準備ができたか | Service のエンドポイントから除外（再起動はしない） |
| `livenessProbe` | コンテナが生きているか（デッドロック検知等） | コンテナを再起動 |

### requests と limits の意味

| 項目 | 説明 |
|---|---|
| `requests` | Pod をスケジュールする際に「最低限これだけのリソースが必要」と宣言する値。スケジューラはこの値を見てノードを選ぶ |
| `limits` | コンテナが使えるリソースの上限。CPU は throttle、メモリは OOMKill |

## ディレクトリ構成

```
step08-probes-resources/
├── README.md
└── deployment.yaml
```

## deployment.yaml の解説

| 項目 | 説明 |
|---|---|
| `readinessProbe` | `/health` に HTTP GET。5 秒後に開始、10 秒間隔で確認 |
| `livenessProbe` | `/health` に HTTP GET。15 秒後に開始、20 秒間隔で確認 |
| `resources.requests` | CPU 100m (0.1 コア), メモリ 64Mi を要求 |
| `resources.limits` | CPU 200m (0.2 コア), メモリ 128Mi を上限に設定 |

## 前提条件

simple-api イメージが必要。プロジェクトルートで以下を実行する。

```bash
make build-images && make load-images
```

## 実行手順

```bash
# 1. Deployment を作成する
kubectl apply -f deployment.yaml

# 2. Pod の状態を確認する
kubectl get pods -l app=probe-demo -w

# 3. Pod の詳細を確認する（Probe の設定・状態が見える）
kubectl describe pod -l app=probe-demo

# 4. Pod のリソース使用量を確認する（metrics-server が必要）
kubectl top pods -l app=probe-demo

# 5. Probe の動作を Events セクションで確認する
kubectl get events --sort-by='.lastTimestamp' | grep probe-demo
```

### Probe 失敗を体験する

Probe のパスをわざと間違えて、失敗時の挙動を観察する。

```bash
# 6. Probe を失敗させる修正を適用する
#    deployment.yaml の readinessProbe.path を /health から /wrong に変更して apply

# 7. Pod の状態を確認する → READY が 0/1 のままになる
kubectl get pods -l app=probe-demo -w

# 8. describe で Readiness probe failed を確認する
kubectl describe pod -l app=probe-demo

# 9. 元に戻す
#    path を /health に戻して apply
```

## 確認方法

1. `kubectl get pods` で全 Pod が `READY 1/1` かつ `Running` であること。
2. `kubectl describe pod` の Events に `Unhealthy` 等のエラーがないこと。
3. Probe パスを `/wrong` に変更した場合、`READY 0/1` となり Service からトラフィックが来なくなることを確認。

```
NAME                          READY   STATUS    RESTARTS   AGE
probe-demo-xxxxx-yyyyy        1/1     Running   0          30s
probe-demo-xxxxx-zzzzz        1/1     Running   0          30s
```

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| readiness と liveness の違いが分からず両方同じ設定にする | 役割が異なる。readiness はトラフィック制御、liveness は再起動判断 | 上の図を参照。readiness は「まだ準備中なら外す」、liveness は「壊れたら再起動」 |
| resources を未設定のままデプロイする | Pod が無制限にリソースを消費し、ノード全体に影響する | 必ず requests と limits を設定する。LimitRange で namespace にデフォルト値を設定するのも有効 |
| liveness の `initialDelaySeconds` が短すぎる | アプリの起動が完了する前に liveness が失敗し、無限再起動ループに陥る | 起動時間を考慮して十分な delay を設定する。本番では startupProbe を使う |
| limits を低く設定しすぎて OOMKill が発生する | `memory limits` を超えるとコンテナが強制終了される | `kubectl describe pod` で `OOMKilled` を確認し、limits を調整する |

## 本番だとどう変わるか

- **startupProbe が重要**: Java アプリなど起動に時間がかかるコンテナでは startupProbe を設定し、起動完了まで liveness/readiness を遅延させる
- **過剰な limits は性能劣化を招く**: CPU limits が低すぎると throttle が発生する。requests を適切に設定し、limits は余裕を持たせるか、CPU limits を外す運用もある
- **VPA (Vertical Pod Autoscaler)**: 実際のリソース使用量に基づいて requests/limits を自動調整するツール
- **PodDisruptionBudget (PDB)**: readinessProbe と組み合わせて、ローリングアップデート中のサービス断を防ぐ

---

次のステップでは、負荷に応じて Pod 数を自動で増減する **HPA (Horizontal Pod Autoscaler)** を学ぶ。Step 09 に進もう。
