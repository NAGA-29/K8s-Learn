# ex01: ローリングアップデートとロールバック

## 目的

Deployment の更新戦略を理解し、無停止デプロイと「やらかした時の切り戻し」を体験する。
Step03 では Pod の自動復旧を学んだが、ここでは「動いているものを安全に入れ替える」方法を学ぶ。

## 学ぶこと

- `RollingUpdate` 戦略と `maxSurge` / `maxUnavailable` の意味
- readinessProbe がローリングアップデートの安全性を支えていること
- `kubectl rollout history` / `undo` によるロールバック
- `kubernetes.io/change-cause` によるリビジョンの記録

### maxSurge / maxUnavailable

| 設定 | 意味 | この演習の値 |
|---|---|---|
| `maxSurge` | 更新中に `replicas` を超えて作ってよい Pod 数 | 1（一時的に4台まで） |
| `maxUnavailable` | 更新中に Ready でなくなってよい Pod 数 | 0（常に3台 Ready を維持） |

`maxUnavailable: 0` + readinessProbe の組み合わせで、「新しい Pod が Ready になってから古い Pod を消す」が保証され、無停止デプロイになる。

## 実行手順

```bash
# 1. v1 (nginx:1.25) をデプロイする
kubectl apply -f deployment-v1.yaml
kubectl rollout status deployment/rollout-demo

# 2. 別ターミナルで Pod の入れ替わりを監視しておく
kubectl get pods -l app=rollout-demo -w

# 3. v2 (nginx:1.27) に更新する
kubectl apply -f deployment-v2.yaml

# 4. ローリングアップデートの進行を確認する
kubectl rollout status deployment/rollout-demo
# → 監視ターミナルで「新 Pod が Ready になってから旧 Pod が Terminating になる」ことを観察する

# 5. リビジョン履歴を確認する（change-cause が表示される）
kubectl rollout history deployment/rollout-demo

# 6. イメージが v2 になっていることを確認する
kubectl get deployment rollout-demo -o jsonpath='{.spec.template.spec.containers[0].image}'

# 7. v1 にロールバックする
kubectl rollout undo deployment/rollout-demo
kubectl rollout status deployment/rollout-demo

# 8. イメージが v1 に戻っていることを確認する
kubectl get deployment rollout-demo -o jsonpath='{.spec.template.spec.containers[0].image}'

# 9. クリーンアップ
kubectl delete deployment rollout-demo
```

### 失敗するデプロイを体験する

存在しないイメージタグに更新して、ロールアウトが「進まなくなる」状況を観察する。

```bash
# 存在しないイメージに更新する
kubectl set image deployment/rollout-demo nginx=nginx:does-not-exist

# 新しい Pod が ImagePullBackOff になり、ロールアウトが止まる
# ただし maxUnavailable: 0 のため、既存の3台は Ready のままサービス継続している
kubectl get pods -l app=rollout-demo
kubectl rollout status deployment/rollout-demo --timeout=30s
# → "timed out waiting for ..." と表示される

# ロールバックして復旧する
kubectl rollout undo deployment/rollout-demo
kubectl rollout status deployment/rollout-demo
```

## 確認方法

1. 手順 4 で、Ready な Pod が常に 3 台を下回らずに入れ替わること
2. `kubectl rollout history` に change-cause 付きでリビジョンが記録されていること
3. `rollout undo` 後にイメージが `nginx:1.25` に戻ること
4. 失敗デプロイ中も既存 Pod がサービスを継続していること

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| 更新中に一瞬アクセスできなくなる | readinessProbe がない、または `maxUnavailable` が大きい | readinessProbe を設定し、`maxUnavailable: 0` にする |
| `rollout history` の CHANGE-CAUSE が `<none>` | change-cause アノテーションを付けていない | マニフェストの `kubernetes.io/change-cause` を更新内容に合わせて書き換える |
| ロールバックしたのに直らない | アプリの問題ではなく ConfigMap など外部リソースの問題 | `rollout undo` は Pod テンプレートしか戻さない。設定変更も合わせて戻す |
| 失敗に気づかず放置する | `kubectl apply` は非同期で、成功したように見える | デプロイ後は必ず `rollout status` で完了を確認する。CI/CD に組み込む |

## 本番だとどう変わるか

- **progressDeadlineSeconds**: ロールアウトが進まない場合に Failed と判定するまでの時間を設定し、CI/CD で自動ロールバックのトリガーにする
- **Canary / Blue-Green**: Argo Rollouts や Flagger を使い、一部トラフィックだけ新バージョンに流して検証してから全体に展開する
- **PodDisruptionBudget との併用**: ノードメンテナンスとデプロイが重なっても可用性を維持する（ex03 参照）
- **メトリクス連動**: デプロイ後のエラー率・レイテンシを監視し、閾値超過で自動ロールバックする
