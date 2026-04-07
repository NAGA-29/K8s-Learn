# Step 14: EKS移行 — kindからクラウドへ

## 目的

kindで学んだKubernetesの構成をAWS EKSに持ち込むとき、何が変わるかを理解する。
ローカル環境とクラウド環境の差分を把握し、移行時に必要な作業を整理する。

## 学ぶこと

- EKS（Elastic Kubernetes Service）の基本概念
- Terraformによるインフラ定義の基礎
- クラウド固有の要素（IAM, VPC, ALB, EBS など）
- kindとEKSの構成差分

## kindとEKSの差分

| 項目 | kind | EKS |
|------|------|-----|
| Control Plane | ローカルコンテナ | AWSマネージド |
| Worker Nodes | Dockerコンテナ | EC2 or Fargate |
| Ingress | NGINX Ingress (kind用) | ALB Ingress Controller / AWS LB Controller |
| Storage | local-path provisioner | EBS CSI Driver / EFS CSI Driver |
| LoadBalancer | 利用不可 | NLB/ALB自動作成 |
| DNS | /etc/hosts手動 | Route53 + ExternalDNS |
| Secret管理 | base64 YAML | Secrets Manager + External Secrets |
| 認証認可 | なし | IAM / IRSA (IAM Roles for Service Accounts) |
| ネットワーク | kindnet | VPC CNI (aws-node) |
| 監視 | 自前Prometheus | CloudWatch / Managed Prometheus |
| ログ | kubectl logs | CloudWatch Logs / Fluentbit |
| コスト | 無料 | EC2/Fargate/NAT Gateway等 |

## Ingressの違い

kindではNGINX Ingress Controller for kindを使い、ホストマシンのポートに直接マッピングしていた。
EKSではAWS Load Balancer Controllerを使い、ALBやNLBが自動作成される。

```yaml
# kind環境でのIngress annotation
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /

# EKS環境でのIngress annotation
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
```

主な違い:
- kindではIngressClassとして`nginx`を指定していたが、EKSでは`alb`を使う
- EKSではALBが自動的にプロビジョニングされ、AWSのロードバランサーとして作成される
- SSL終端もALB側で行えるため、cert-managerが不要になるケースがある

## Storageの違い

- **kind**: local-path provisioner（ノードローカルにデータ保存、Podが別ノードに移動するとデータを失う）
- **EKS**:
  - **EBS CSI Driver**: ブロックストレージ。AZ固定のため、Podが別AZに移動するとアタッチできない
  - **EFS CSI Driver**: ファイルストレージ。マルチAZ対応で複数Podから同時マウント可能
  - いずれもCSI Driverのインストールが必要

```yaml
# EKS用のStorageClass例（EBS）
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ebs-sc
provisioner: ebs.csi.aws.com
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: gp3
```

## IAM/IRSA（IAM Roles for Service Accounts）の概念

PodがAWSリソース（S3, DynamoDB, Secrets Managerなど）にアクセスするための仕組み。

**なぜ必要か:**
- Pod内にAWSアクセスキーをハードコードするのは危険
- ノード全体に権限を与えるのは過剰（最小権限の原則に反する）
- IRSAを使えばPod単位で必要最小限のIAM権限を付与できる

**仕組み:**
1. IAM Roleを作成し、必要なPolicyをアタッチする
2. KubernetesのServiceAccountにIAM RoleのARNをannotationで紐付ける
3. Podがそのannotation付きのServiceAccountを使うことで、自動的にIAM Roleの権限を取得する

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-app
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/my-app-role
```

## ディレクトリ構成

```
step14-eks-migration/
├── README.md          # このファイル
└── terraform/
    ├── README.md      # Terraform使用上の注意
    ├── main.tf        # VPC + EKSクラスタ定義
    ├── variables.tf   # 変数定義
    └── outputs.tf     # 出力定義
```

## 実行手順

このステップのTerraformはリファレンス用であり、実際に実行するとAWSリソースが作成されコストが発生する。
まずはコードを読んで構成を理解することを目的とする。

### 1. Terraformコードを読む

```bash
cat terraform/main.tf
cat terraform/variables.tf
cat terraform/outputs.tf
```

### 2. （任意）実際にEKSクラスタを作成する場合

```bash
# AWS CLIの設定が完了していることを確認
aws sts get-caller-identity

# Terraformの初期化
cd terraform
terraform init

# 実行計画の確認（リソースは作成されない）
terraform plan

# リソースの作成（コストが発生する）
terraform apply

# kubectlの設定
aws eks update-kubeconfig --region ap-northeast-1 --name k8s-learning

# クラスタの確認
kubectl get nodes
```

### 3. （任意）リソースの削除

```bash
# 必ず削除すること。放置するとコストが発生し続ける
terraform destroy
```

## 確認方法

### Terraformコードの理解度チェック

```bash
# main.tfを読んで以下の質問に答えられるか確認する

# Q1: VPCのCIDRは何か？
# A1: 10.0.0.0/16

# Q2: NAT Gatewayはいくつ作成されるか？
# A2: 1つ（single_nat_gateway = true）

# Q3: EKSのワーカーノードのインスタンスタイプは？
# A3: t3.medium

# Q4: ワーカーノードのdesired_sizeは？
# A4: 2
```

### 実際にクラスタを作成した場合

```bash
# ノードの確認
kubectl get nodes -o wide

# システムPodの確認（VPC CNIなどが動いているか）
kubectl get pods -n kube-system

# StorageClassの確認
kubectl get storageclass

# IngressClassの確認
kubectl get ingressclass
```

## 移行チェックリスト

kindで作ったアプリケーションをEKSに移行する際に確認すべき項目:

- [ ] VPC/サブネット設計
- [ ] EKSクラスタ作成
- [ ] ノードグループ設定
- [ ] IAM Role/Policy設計
- [ ] Ingress Controller差し替え（NGINX → AWS LB Controller）
- [ ] Storage Class設定（local-path → EBS/EFS）
- [ ] Secret管理方法変更（YAML → Secrets Manager + External Secrets）
- [ ] 監視/ログ基盤構築（CloudWatch / Fluentbit）
- [ ] CI/CDパイプライン構築
- [ ] コスト見積もり

## dev環境想定の軽量構成ポイント

学習やdev環境であれば、以下のようにコストを抑えられる:

- **インスタンスタイプ**: t3.medium x 2 で十分（本番ではm5.largeなどを検討）
- **NAT Gateway**: 1AZのみに配置（本番では各AZに配置）
- **Spot Instances**: ステートレスなワークロードにはSpot Instancesを活用し最大90%コスト削減
- **EKSクラスタ**: 1クラスタのみ（本番ではdev/staging/prodを分離）

## よくある失敗

| 失敗 | 原因 | 対処 |
|------|------|------|
| `terraform apply`が途中で失敗する | IAM権限不足 | AdministratorAccessまたは必要な権限を付与する |
| Podが`Pending`のまま | ノードグループのリソース不足 | `kubectl describe pod`でイベントを確認し、ノード数やインスタンスタイプを調整 |
| ALBが作成されない | AWS LB Controllerが未インストール | Helmでaws-load-balancer-controllerをインストールする |
| PodからS3にアクセスできない | IRSAの設定漏れ | ServiceAccountにIAM RoleのARN annotationを付与する |
| EBSボリュームがアタッチできない | PodとEBSが異なるAZにある | `volumeBindingMode: WaitForFirstConsumer`を設定する |
| `terraform destroy`し忘れ | 作業後の削除忘れ | 必ず削除する。CloudWatchで請求アラートを設定しておく |
| kubectlが繋がらない | kubeconfigが未設定 | `aws eks update-kubeconfig`を実行する |

## 本番だとどう変わるか

| 観点 | dev/学習環境 | 本番環境 |
|------|-------------|---------|
| ノード数 | 2台 | 3台以上（マルチAZ） |
| NAT Gateway | 1AZ | 各AZに配置（高可用性） |
| インスタンス | t3.medium | m5.large以上 |
| Spot Instance | 積極利用 | On-Demandをベースに一部Spot |
| クラスタ分離 | 1クラスタ | dev/staging/prod分離 |
| Secret管理 | K8s Secret | Secrets Manager + External Secrets Operator |
| 監視 | 最低限 | CloudWatch + Managed Prometheus + Grafana |
| ログ | kubectl logs | Fluentbit → CloudWatch Logs / S3 |
| バックアップ | なし | Velero等でクラスタバックアップ |
| CI/CD | 手動デプロイ | ArgoCD / Flux によるGitOps |
| ネットワーク | パブリックエンドポイント | プライベートエンドポイント + VPN/DirectConnect |
| コスト | 月額$100-200程度 | 月額$1,000以上（構成次第） |

## 次のステップ

ここまででHTTPベースのアプリケーションをKubernetes上で動かす方法を学んできた。Step 15ではHTTPとは異なるWebSocket通信をKubernetes上で扱う方法を学ぶ。WebSocketはリアルタイム通信（チャット、通知、ライブ更新など）で広く使われており、Kubernetes上での扱いにはHTTPとは異なる考慮点がある。
