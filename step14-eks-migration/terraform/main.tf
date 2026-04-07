# =============================================================================
# EKS クラスタ構成（リファレンス用）
# =============================================================================
# このファイルはkindで学んだ構成をEKSに移行する際の参考例です。
# 実際に terraform apply するとAWSリソースが作成され、コストが発生します。
# =============================================================================

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region
}

# -----------------------------------------------------------------------------
# VPC
# -----------------------------------------------------------------------------
# EKSクラスタを配置するVPCを作成する。
# kindではDockerネットワーク内で完結していたが、EKSではVPC設計が必要になる。
# -----------------------------------------------------------------------------
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${var.project_name}-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["${var.region}a", "${var.region}c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]

  # NAT Gatewayの設定
  # dev環境ではコスト削減のため1つだけ作成する
  # 本番環境では各AZに配置すること（single_nat_gateway = false）
  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  # サブネットタグ
  # AWS Load Balancer Controllerがサブネットを自動検出するために必要
  public_subnet_tags = {
    "kubernetes.io/role/elb" = 1
  }
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
  }

  tags = var.tags
}

# -----------------------------------------------------------------------------
# EKS クラスタ
# -----------------------------------------------------------------------------
# kindではローカルのDockerコンテナがControl Planeだったが、
# EKSではAWSがControl Planeをマネージドで提供する。
# ワーカーノードのみを管理すればよい。
# -----------------------------------------------------------------------------
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = var.project_name
  cluster_version = "1.29"

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  # クラスタエンドポイントへのパブリックアクセスを許可
  # 本番環境ではfalseにしてVPN/DirectConnect経由でアクセスすること
  cluster_endpoint_public_access = true

  # マネージドノードグループ
  # kindではDockerコンテナがワーカーノードだったが、
  # EKSではEC2インスタンスがワーカーノードになる
  eks_managed_node_groups = {
    default = {
      instance_types = ["t3.medium"]
      min_size       = 1
      max_size       = 3
      desired_size   = 2
    }
  }

  tags = var.tags
}
