# Step 01: kind でローカル Kubernetes クラスタを構築する

## 目的

ローカルマシン上に Kubernetes クラスタを作成し、`kubectl` の基本操作に慣れる。
本ステップが以降すべてのハンズオンの土台になる。

## 学ぶこと

- kind (Kubernetes in Docker) を使ったクラスタ構築
- control-plane / worker ノードの役割
- `kubectl` によるクラスタ情報の確認方法
- extraPortMappings によるホストとコンテナ間のポート転送

## ディレクトリ構成

```
step01-kind-cluster/
├── README.md
└── kind-config.yaml
```

## kind-config.yaml の解説

| 項目 | 説明 |
|---|---|
| `nodes[0].role: control-plane` | Kubernetes の管理コンポーネント (API Server, etcd, scheduler 等) が動くノード |
| `kubeadmConfigPatches` | control-plane ノードに `ingress-ready=true` ラベルを付与し、後の Step05 で Ingress Controller を配置できるようにする |
| `extraPortMappings` | コンテナ内のポート 80/443 をホストの同ポートにマッピング。Ingress 経由の HTTP/HTTPS アクセスに必要 |
| `nodes[1..2].role: worker` | 実際にアプリケーション Pod が動くノード。2 台用意することで Pod の分散配置を体験できる |

## 実行手順

```bash
# 1. クラスタを作成する
kind create cluster --name k8s-learning --config kind-config.yaml

# 2. クラスタ情報を確認する
kubectl cluster-info

# 3. ノード一覧を確認する
kubectl get nodes

# 4. デフォルトの Namespace を確認する
kubectl get namespaces
```

## 確認方法

`kubectl get nodes` を実行し、3 つのノード (control-plane x1, worker x2) すべてが **Ready** になっていれば成功。

```
NAME                          STATUS   ROLES           AGE   VERSION
k8s-learning-control-plane    Ready    control-plane   1m    v1.x.x
k8s-learning-worker           Ready    <none>          1m    v1.x.x
k8s-learning-worker2          Ready    <none>          1m    v1.x.x
```

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| `Cannot connect to the Docker daemon` | Docker Desktop / Docker Engine が起動していない | Docker を起動してから再実行 |
| `Bind for 0.0.0.0:80 failed: port is already allocated` | ホストのポート 80 が別プロセスで使用中 | `lsof -i :80` で確認し、該当プロセスを停止するか `hostPort` を変更 |
| ノードが 1 台しかない | worker ノードの定義を忘れた、または kind-config.yaml を指定せずに作成した | `kind delete cluster --name k8s-learning` で削除してから再作成 |
| worker 無しで「動いたから OK」と満足する | control-plane だけでも Pod は動くが、本番とかけ離れた構成になる | 必ず worker ノードを含む構成で学習を進める |

## 本番だとどう変わるか

- **マネージド Kubernetes**: AWS EKS / Google GKE / Azure AKS を使い、control-plane はクラウドプロバイダが管理する
- **CNI プラグイン**: kind はデフォルトで kindnet を使うが、本番では Calico, Cilium, AWS VPC CNI など要件に合わせて選択する
- **LoadBalancer**: kind では LoadBalancer タイプの Service がそのままでは使えないが、本番ではクラウド LB が自動で払い出される
- **StorageClass**: kind はローカルストレージだが、本番では EBS, Persistent Disk, Azure Disk 等を使う
- **ノード数・スペック**: 本番では数十〜数百ノードを運用し、オートスケーリングも設定する

---

次のステップでは、このクラスタ上に最小のワークロード単位である **Pod** をデプロイし、コンテナの動作確認やログ確認の方法を学ぶ。Step 02 に進もう。
