# ex02: Job / CronJob — バッチ処理と定期実行

## 目的

「動き続ける」Deployment とは異なる、「実行して終わる」ワークロードを学ぶ。
バッチ処理（Job）と定期実行（CronJob）の管理方法、リトライと多重起動の制御を理解する。

## 学ぶこと

- Job の `completions` / `parallelism` / `backoffLimit`
- `restartPolicy` が Deployment と異なる理由（`Never` / `OnFailure`）
- CronJob の `schedule` と `concurrencyPolicy`
- 完了した Pod の後片付け（`ttlSecondsAfterFinished`、履歴上限）

### Deployment と Job の違い

| | Deployment | Job |
|---|---|---|
| 目的 | サービスを動かし続ける | 処理を完了させる |
| 正常な終了 | ない（終了 = 異常） | ある（exit 0 = 成功） |
| restartPolicy | Always | Never / OnFailure |
| 失敗時 | 同じ Pod を再起動 | リトライ上限まで Pod を作り直す |

## 実行手順

### Job

```bash
# 1. Job を作成する
kubectl apply -f job.yaml

# 2. Job と Pod の状態を観察する
#    completions: 3, parallelism: 2 のため、まず2つ並列で動き、
#    完了したら3つ目が起動する
kubectl get jobs,pods -w

# 3. 完了した Pod のログを確認する
kubectl logs -l job-name=batch-demo --tail=20

# 4. Job の詳細を確認する（Completions が 3/3 になっている）
kubectl describe job batch-demo
```

### 失敗するJobを体験する

```bash
# わざと失敗するJobを作成する（exit 1 で終了する）
kubectl create job fail-demo --image=busybox:1.36 -- /bin/sh -c "echo failing; exit 1"

# backoffLimit（デフォルト6）までPodが作り直されるのを観察する
kubectl get pods -l job-name=fail-demo -w

# Job が Failed になっていることを確認する
kubectl describe job fail-demo | grep -A 5 Conditions

# クリーンアップ
kubectl delete job fail-demo
```

### CronJob

```bash
# 1. CronJob を作成する
kubectl apply -f cronjob.yaml

# 2. 毎分 Job が作られるのを観察する（2〜3分待つ）
kubectl get cronjobs,jobs -w

# 3. 実行結果のログを確認する
kubectl logs -l job-name --tail=20 --prefix | grep report

# 4. 手動で即時実行する（スケジュールを待たずにテストできる）
kubectl create job report-manual --from=cronjob/report-demo
kubectl logs -l job-name=report-manual

# 5. クリーンアップ（CronJob を消すと配下の Job も消える）
kubectl delete cronjob report-demo
kubectl delete job report-manual batch-demo --ignore-not-found
```

## 確認方法

1. `batch-demo` の COMPLETIONS が `3/3` になること
2. 並列実行中に Running の Pod が同時に2つまでしか存在しないこと
3. `fail-demo` がリトライの末に Failed になること
4. CronJob から毎分 Job が生成され、履歴が `successfulJobsHistoryLimit: 3` 件に保たれること

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| `restartPolicy: Always` でエラーになる | Job では Always を指定できない | `Never` または `OnFailure` を指定する |
| 失敗した Job が無限に Pod を作る | `backoffLimit` が大きすぎる | 適切なリトライ回数を設定し、失敗時のアラートを用意する |
| バッチが多重起動してデータが壊れる | `concurrencyPolicy` がデフォルト（Allow）のまま | 多重起動できない処理には `Forbid` を設定する |
| 完了した Pod が大量に残る | 後片付けの設定がない | `ttlSecondsAfterFinished` や履歴上限を設定する |
| CronJob が動いた形跡がない | schedule の書式ミス、またはクラスタのタイムゾーンの誤解 | `kubectl describe cronjob` でイベントを確認。スケジュールは UTC 基準（`timeZone` フィールドで変更可能） |

## 本番だとどう変わるか

- **ワークフローエンジン**: 依存関係のある複数ジョブは Argo Workflows などで DAG として管理する
- **キュー連動**: SQS や Kafka のメッセージをトリガーに Job を起動する（KEDA の ScaledJob 等）
- **冪等性の設計**: リトライ・多重起動されても安全なように、バッチ処理自体を冪等に実装する
- **監視**: Job の失敗を Prometheus（kube-state-metrics）+ Alertmanager で通知する
- **タイムゾーン**: `spec.timeZone`（K8s 1.27+）で明示し、深夜バッチのズレを防ぐ
