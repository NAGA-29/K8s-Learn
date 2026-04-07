# =============================================================================
# 変数定義
# =============================================================================

variable "region" {
  description = "AWSリージョン"
  type        = string
  default     = "ap-northeast-1"
}

variable "project_name" {
  description = "リソース命名に使用するプロジェクト名"
  type        = string
  default     = "k8s-learning"
}

variable "tags" {
  description = "全リソースに付与する共通タグ"
  type        = map(string)
  default = {
    Environment = "dev"
    Project     = "k8s-learning"
    ManagedBy   = "terraform"
  }
}
