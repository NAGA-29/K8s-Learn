# Terraform リファレンス構成

## 注意事項

このTerraformコードはリファレンス用である。
`terraform apply` を実行すると実際にAWSリソースが作成され、コストが発生する。

学習目的であればコードを読んで構成を理解するだけで十分である。

## 構成概要

| ファイル | 内容 |
|---------|------|
| `main.tf` | VPCとEKSクラスタの定義 |
| `variables.tf` | リージョン、プロジェクト名、タグなどの変数 |
| `outputs.tf` | クラスタ接続情報の出力 |

## 作成されるリソース

- **VPC**: 10.0.0.0/16 のCIDRブロック
  - パブリックサブネット x 2（ALB用）
  - プライベートサブネット x 2（ワーカーノード用）
  - NAT Gateway x 1（コスト削減のため1つのみ）
- **EKSクラスタ**: Kubernetes 1.29
  - マネージドノードグループ: t3.medium x 2台

## 使い方

```bash
# 1. AWS CLIの認証設定が完了していることを確認
aws sts get-caller-identity

# 2. Terraformの初期化（プロバイダとモジュールのダウンロード）
terraform init

# 3. 実行計画の確認（この時点ではリソースは作成されない）
terraform plan

# 4. リソースの作成（ここからコストが発生する）
terraform apply

# 5. kubectlの設定
aws eks update-kubeconfig --region ap-northeast-1 --name k8s-learning

# 6. 動作確認
kubectl get nodes

# 7. 作業完了後は必ず削除すること
terraform destroy
```

## 推定コスト（ap-northeast-1）

| リソース | 概算月額 |
|---------|---------|
| EKS Control Plane | ~$73 |
| t3.medium x 2 | ~$67 |
| NAT Gateway | ~$45 |
| その他（EBS等） | ~$10 |
| **合計** | **~$195/月** |

不要になったら速やかに `terraform destroy` でリソースを削除すること。
