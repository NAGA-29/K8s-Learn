# Step 07: Volume — Pod が消えても残したいデータの考え方

## 目的

Pod が消えても残したいデータの考え方を学ぶ。
コンテナ内のファイルはコンテナが再起動すると消える。この事実を体験し、永続化の仕組みを理解する。

## 学ぶこと

- コンテナ内ファイルのライフサイクル（Pod 削除で消える）
- `emptyDir` と `PersistentVolumeClaim (PVC)` の違い
- PVC を使った永続ストレージの基本
- Volume のマウント方法

### emptyDir と PVC の違い

| 項目 | emptyDir | PVC |
|---|---|---|
| ライフサイクル | Pod と同じ（Pod 削除で消える） | Pod とは独立（Pod 削除後もデータが残る） |
| 用途 | 一時キャッシュ、Pod 内コンテナ間のデータ共有 | データベースファイル、アップロードデータなど永続化が必要なもの |
| ストレージ | ノードのディスク（またはメモリ） | クラスタの StorageClass に基づくプロビジョニング |

## ディレクトリ構成

```
step07-volume/
├── README.md
├── pvc.yaml
└── deployment.yaml
```

## 各ファイルの解説

### pvc.yaml

| 項目 | 説明 |
|---|---|
| `accessModes: ReadWriteOnce` | 単一ノードからの読み書きを許可。最も一般的なモード |
| `resources.requests.storage: 1Gi` | 1GiB のストレージを要求 |

### deployment.yaml

| 項目 | 説明 |
|---|---|
| `volumes[0]: persistentVolumeClaim` | PVC `data-pvc` を `data` ボリュームとしてマウント。Pod が再起動してもデータが残る |
| `volumes[1]: emptyDir` | 一時ボリューム `tmp` を作成。Pod が削除されるとデータも消える |
| `volumeMounts` | `/data` に PVC、`/tmp/scratch` に emptyDir をマウント |
| `command / args` | 起動時に `/data/log.txt` へタイムスタンプを追記し、nginx を起動 |

## 実行手順

```bash
# 1. PVC を作成する
kubectl apply -f pvc.yaml

# 2. PVC の状態を確認する（Bound になるまで待つ）
kubectl get pvc data-pvc

# 3. Deployment を作成する
kubectl apply -f deployment.yaml

# 4. Pod が Running になるまで待つ
kubectl get pods -l app=volume-demo -w

# 5. Pod に入って /data/log.txt を確認する
kubectl exec -it $(kubectl get pod -l app=volume-demo -o jsonpath='{.items[0].metadata.name}') -- cat /data/log.txt

# 6. /tmp/scratch にもファイルを作る（比較用）
kubectl exec -it $(kubectl get pod -l app=volume-demo -o jsonpath='{.items[0].metadata.name}') -- sh -c 'echo "temp data" > /tmp/scratch/temp.txt && cat /tmp/scratch/temp.txt'

# 7. Pod を削除して再作成させる
kubectl delete pod -l app=volume-demo

# 8. 新しい Pod が起動したら再度確認
kubectl get pods -l app=volume-demo -w

# 9. /data/log.txt を確認 → 前回の記録 + 新しい記録が残っている
kubectl exec -it $(kubectl get pod -l app=volume-demo -o jsonpath='{.items[0].metadata.name}') -- cat /data/log.txt

# 10. /tmp/scratch を確認 → temp.txt は消えている
kubectl exec -it $(kubectl get pod -l app=volume-demo -o jsonpath='{.items[0].metadata.name}') -- ls /tmp/scratch/
```

## 確認方法

1. 手順 9 で `/data/log.txt` に **2 行分**のタイムスタンプ（最初の起動 + 再起動）が記録されていれば PVC による永続化が成功。
2. 手順 10 で `/tmp/scratch/` が空であれば emptyDir の揮発性を確認できている。

```
# /data/log.txt の期待出力
Pod started at Tue Apr  7 12:00:00 UTC 2026
Pod started at Tue Apr  7 12:01:30 UTC 2026
```

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| PVC が `Pending` のまま | StorageClass が存在しない、またはプロビジョナーが未設定 | `kubectl get sc` で StorageClass を確認。kind はデフォルトで `standard` が使える |
| Pod 内ファイルが永続化されると思い込む | Volume をマウントしていないパスのファイルはコンテナ再起動で消える | 永続化したいデータは必ず PVC マウント先に書く |
| `local-path` で本番も OK と思い込む | kind のデフォルト StorageClass はノードローカル。ノード障害でデータが消失する | 本番ではクラウドプロバイダのストレージを使う |
| emptyDir にデータを置いて「永続化できた」と思う | emptyDir は Pod のライフサイクルに縛られる | emptyDir はあくまで一時ストレージ。永続化には PVC を使う |

## 本番だとどう変わるか

- **CSI Driver**: クラウドプロバイダ固有のストレージ（AWS EBS, EFS, GCP Persistent Disk）を CSI Driver 経由で利用する
- **StorageClass**: `gp3`, `io2` など性能要件に合わせた StorageClass を定義する
- **StatefulSet**: データベースなど、各 Pod に固有の永続ボリュームが必要な場合は StatefulSet + volumeClaimTemplates を使う
- **バックアップ**: Velero などのツールで PV のスナップショットを定期取得する
- **ReadWriteMany**: 複数 Pod から同時に書き込む場合は EFS (NFS) のような共有ストレージが必要

---

次のステップでは、Pod が正しく起動しているか・生きているかを Kubernetes に伝える **Probe（ヘルスチェック）** と、CPU/メモリの **リソース制限** を学ぶ。Step 08 に進もう。
