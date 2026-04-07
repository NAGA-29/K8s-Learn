# Step 03: Deployment -- Pod の自動復旧とレプリカ管理

## 目的

Deployment を使って Pod を宣言的に管理し、自動復旧 (self-healing) とスケーリングの仕組みを理解する。

## 学ぶこと

- Desired State (望ましい状態) と Reconciliation Loop の概念
- ReplicaSet と Deployment の関係
- Pod の自動復旧の動作
- スケールアウト / スケールインの方法

### Desired State とは

Kubernetes は「今の状態」を「望ましい状態 (desired state)」に近づけ続ける。Deployment で `replicas: 3` と宣言すれば、Pod が何らかの理由で減っても、コントローラが自動的に Pod を再作成して 3 台を維持する。これが Kubernetes の根幹となる考え方である。

## ディレクトリ構成

```
step03-deployment/
├── README.md
└── deployment.yaml
```

## 実行手順

```bash
# 1. Deployment を作成する
kubectl apply -f deployment.yaml

# 2. Deployment の状態を確認する
kubectl get deployments

# 3. 作成された Pod を確認する
kubectl get pods -l app=nginx

# 4. 自動復旧を体験する -- Pod を 1 つ削除してみる
kubectl delete pod <Pod名>   # kubectl get pods で表示された Pod 名を指定

# 5. すぐに Pod 一覧を確認する (自動復旧を観察)
kubectl get pods -l app=nginx -w

# 6. ロールアウト状態を確認する
kubectl rollout status deployment/nginx-deployment

# 7. スケールアウトする
kubectl scale deployment/nginx-deployment --replicas=5
kubectl get pods -l app=nginx

# 8. スケールインする
kubectl scale deployment/nginx-deployment --replicas=3
kubectl get pods -l app=nginx
```

## 確認方法

- `kubectl get deployments` で READY が `3/3` であること
- Pod を 1 つ削除しても、数秒で新しい Pod が自動作成され、再び 3 台になること
- `kubectl rollout status` が `successfully rolled out` を返すこと

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| Pod が作成されない | `selector.matchLabels` と `template.metadata.labels` が一致していない | 両方のラベルが完全に一致しているか確認する |
| `replicas: 1` で満足する | 1 台だと Pod 再起動中にダウンタイムが発生する | 最低 2 台、できれば 3 台以上に設定する |
| Deployment を削除せずに Pod だけ削除し続ける | Deployment が存在する限り Pod は再作成される | クリーンアップは `kubectl delete deployment` で行う |
| ReplicaSet を直接操作する | Deployment が管理する ReplicaSet を手動変更すると不整合が起きる | 常に Deployment 経由で操作する |

## 本番だとどう変わるか

- **RollingUpdate 戦略**: デフォルトの更新戦略。`maxUnavailable` と `maxSurge` を調整して無停止デプロイを実現する
- **Canary デプロイ**: 新バージョンを一部の Pod にだけ展開し、問題がなければ全体に広げる
- **Blue/Green デプロイ**: 旧バージョンと新バージョンを並行稼働させ、切り替える
- **HPA (Horizontal Pod Autoscaler)**: CPU やメモリの使用率に応じて自動でレプリカ数を増減させる
- **PodDisruptionBudget**: ノードメンテナンス時に同時に停止できる Pod 数を制限する

---

Deployment で Pod が安定的に動くようになった。しかし、Pod の IP は再作成のたびに変わるため、直接 IP を指定してアクセスするのは現実的ではない。次の Step 04 では、Pod へのアクセスを抽象化する **Service** を学ぶ。
