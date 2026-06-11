# Step 02: Pod -- Kubernetes の最小実行単位

## 目的

Pod が Kubernetes における最小のデプロイ単位であることを理解し、Pod の作成・確認・操作・削除の一連のライフサイクルを体験する。

## 学ぶこと

- Pod とコンテナの違い
- Pod のライフサイクル (Pending -> Running -> Succeeded/Failed)
- `kubectl` による Pod の操作 (logs, exec, describe)
- なぜ Pod を直接運用してはいけないのか

### Pod とコンテナの違い

Docker のコンテナは「1 つのプロセスを隔離する単位」だが、Kubernetes の Pod は「1 つ以上のコンテナをまとめて管理する単位」である。同じ Pod 内のコンテナはネットワーク (localhost) とストレージ (Volume) を共有する。ほとんどのケースでは 1 Pod = 1 コンテナだが、サイドカーパターンなどで複数コンテナを含むこともある。

## ディレクトリ構成

```
step02-pod/
├── README.md
└── pod.yaml
```

## 実行手順

```bash
# 1. Pod を作成する
kubectl apply -f pod.yaml

# 2. Pod の状態を確認する
kubectl get pods

# 3. Pod のログを確認する
kubectl logs my-nginx

# 4. Pod 内に入ってシェル操作する
kubectl exec -it my-nginx -- /bin/bash

# (Pod 内で) nginx の設定や配信ファイルを確認
# ※ nginx 公式イメージには curl / wget が入っていないため、HTTP での確認は手順 4.5 で行う
cat /usr/share/nginx/html/index.html
exit

# 4.5. port-forward でホストから nginx の応答を確認する
kubectl port-forward pod/my-nginx 8080:80 &
curl http://localhost:8080
kill %1   # port-forward を終了する

# 5. Pod の詳細情報を確認する
kubectl describe pod my-nginx

# 6. Pod を削除する
kubectl delete pod my-nginx
```

## 確認方法

- `kubectl get pods` で STATUS が **Running** であること
- `kubectl exec -it my-nginx -- /bin/bash` でコンテナ内のシェルに入れること
- `kubectl logs my-nginx` でアクセスログが確認できること
- `kubectl describe pod my-nginx` で Events セクションにエラーがないこと

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| STATUS が `ImagePullBackOff` | イメージ名やタグの typo、レジストリへのアクセス不可 | `kubectl describe pod` で Events を確認し、イメージ名を修正 |
| Pod を直接本番運用する | Pod が死んでも自動復旧されない | 次の Step03 で学ぶ Deployment を使う |
| 「restart した=問題ない」と思い込む | RESTARTS が増えているのは異常。CrashLoopBackOff の前兆 | `kubectl logs --previous` で前回のログを確認し、根本原因を調査 |
| `exec` でシェルに入れない | コンテナに `/bin/bash` がない (Alpine 系など) | `/bin/sh` を試す |

## 本番だとどう変わるか

- **Pod を直接作成することはほぼない**: Deployment, StatefulSet, Job, DaemonSet といった上位リソース経由で Pod を管理する
- **リソース制限**: `resources.requests` / `resources.limits` で CPU・メモリを必ず設定する
- **ヘルスチェック**: `livenessProbe` / `readinessProbe` を設定し、異常時の自動再起動やトラフィック遮断を行う
- **セキュリティ**: `securityContext` で root 実行を禁止し、`readOnlyRootFilesystem` を有効にする

---

Pod の仕組みがわかったところで、次は Pod を「望ましい状態」に保ち続ける仕組みである **Deployment** を学ぶ。Pod が落ちても自動復旧する世界を Step 03 で体験しよう。
