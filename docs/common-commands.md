# よく使うコマンド集

Kubernetes学習で頻繁に使うコマンドをまとめています。すべてコピペして実行できます。

---

## kubectl get - リソース一覧の取得

### Pod一覧

```bash
kubectl get pods
```

詳細情報（IPアドレス、ノード名）も表示:

```bash
kubectl get pods -o wide
```

### Deployment一覧

```bash
kubectl get deployments
```

### Service一覧

```bash
kubectl get services
```

### Node一覧

```bash
kubectl get nodes
```

### すべてのリソースを一括表示

```bash
kubectl get all
```

### ラベルで絞り込み

```bash
kubectl get pods -l app=simple-api
```

### YAML形式で出力（設定確認に便利）

```bash
kubectl get deployment simple-api -o yaml
```

---

## kubectl describe - リソースの詳細情報

### Podの詳細（イベント、状態、エラーの確認に最も使う）

```bash
kubectl describe pod <pod名>
```

### Deploymentの詳細

```bash
kubectl describe deployment <deployment名>
```

### Serviceの詳細（EndpointsでPodとの紐付けを確認できる）

```bash
kubectl describe service <service名>
```

### Nodeの詳細（リソース使用量、条件を確認）

```bash
kubectl describe node <node名>
```

---

## kubectl logs - ログの確認

### Podのログを表示

```bash
kubectl logs <pod名>
```

### リアルタイムでログを追跡（-f: follow）

```bash
kubectl logs -f <pod名>
```

### クラッシュ前のコンテナのログを確認（--previous）

CrashLoopBackOffの原因調査に重要:

```bash
kubectl logs --previous <pod名>
```

### 複数コンテナがあるPodで特定コンテナのログ

```bash
kubectl logs <pod名> -c <コンテナ名>
```

### 直近のログだけ表示

```bash
kubectl logs --tail=50 <pod名>
```

---

## kubectl exec - コンテナ内でコマンド実行

### コンテナ内でシェルを起動（対話モード）

```bash
kubectl exec -it <pod名> -- /bin/sh
```

bashが使える場合:

```bash
kubectl exec -it <pod名> -- /bin/bash
```

### 単発コマンドを実行

```bash
kubectl exec <pod名> -- env
kubectl exec <pod名> -- cat /etc/config/app.conf
kubectl exec <pod名> -- wget -qO- http://localhost:8080/health
```

---

## kubectl rollout - デプロイ管理

### ロールアウトの状態確認

```bash
kubectl rollout status deployment/<deployment名>
```

### デプロイ履歴の確認

```bash
kubectl rollout history deployment/<deployment名>
```

### 前のバージョンにロールバック

```bash
kubectl rollout undo deployment/<deployment名>
```

### 特定リビジョンにロールバック

```bash
kubectl rollout undo deployment/<deployment名> --to-revision=2
```

---

## kubectl top - リソース使用量の確認

> **注意:** metrics-serverがインストールされている必要があります。

### ノードのリソース使用量

```bash
kubectl top nodes
```

### Podのリソース使用量

```bash
kubectl top pods
```

### 特定NamespaceのPod使用量

```bash
kubectl top pods -n monitoring
```

---

## kubectl port-forward - ポートフォワード

### PodへのポートフォワードPort

```bash
kubectl port-forward pod/<pod名> 8080:8080
```

### Serviceへのポートフォワード

```bash
kubectl port-forward svc/<service名> 8080:80
```

ブラウザで http://localhost:8080 にアクセスして確認できる。Ctrl+Cで終了。

---

## kubectl apply / delete - リソースの作成・削除

### マニフェストを適用（作成または更新）

```bash
kubectl apply -f deployment.yaml
```

### ディレクトリ内の全マニフェストを適用

```bash
kubectl apply -f step11-mini-architecture/
```

### リソースの削除

```bash
kubectl delete -f deployment.yaml
```

### Pod名を指定して削除

```bash
kubectl delete pod <pod名>
```

### Deploymentを削除（配下のPodも削除される）

```bash
kubectl delete deployment <deployment名>
```

---

## kind load docker-image - kindへイメージ読み込み

kindクラスタはDockerレジストリに直接アクセスしないため、ローカルでビルドしたイメージをクラスタに読み込む必要がある。

```bash
kind load docker-image simple-api:latest --name k8s-learning
```

複数イメージを一括:

```bash
kind load docker-image simple-api:latest simple-web:latest --name k8s-learning
```

> **よくあるミス:** `kind load` を忘れて `ImagePullBackOff` になるケースが非常に多い。ローカルイメージを使う場合は必ず `imagePullPolicy: Never` または `IfNotPresent` を設定すること。

---

## Namespace指定 (-n)

### 特定NamespaceのPodを取得

```bash
kubectl get pods -n kube-system
```

### 全NamespaceのPodを取得

```bash
kubectl get pods --all-namespaces
# または短縮形
kubectl get pods -A
```

### Namespaceを作成

```bash
kubectl create namespace monitoring
```

### マニフェスト適用時にNamespace指定

```bash
kubectl apply -f deployment.yaml -n staging
```

---

## Context切り替え

### 現在のContextを確認

```bash
kubectl config current-context
```

### 利用可能なContext一覧

```bash
kubectl config get-contexts
```

### Contextを切り替え

```bash
kubectl config use-context kind-k8s-learning
```

### EKSクラスタのContextを追加（Step14で使用）

```bash
aws eks update-kubeconfig --name my-cluster --region ap-northeast-1
```

### デフォルトNamespaceを設定（毎回 -n を付けなくて済む）

```bash
kubectl config set-context --current --namespace=monitoring
```
