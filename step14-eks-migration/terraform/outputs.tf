# =============================================================================
# 出力定義
# =============================================================================
# terraform apply 後に表示される値。
# クラスタへの接続情報などを確認するために使用する。
# =============================================================================

output "cluster_endpoint" {
  description = "EKSクラスタのエンドポイントURL"
  value       = module.eks.cluster_endpoint
}

output "cluster_name" {
  description = "EKSクラスタ名"
  value       = module.eks.cluster_name
}

output "region" {
  description = "AWSリージョン"
  value       = var.region
}

output "configure_kubectl" {
  description = "kubectlを設定するためのコマンド"
  value       = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name}"
}
