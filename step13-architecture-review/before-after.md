# Before / After 比較表

Step11のミニ構成（Before）と、本番想定の改善構成（After）を比較します。

---

## 構成比較

| 観点 | Before（kind学習構成） | After（本番想定） |
|------|----------------------|------------------|
| **クラスタ** | kind（ローカル） | EKS（マネージド） |
| **ノード数** | 1 control-plane + 2 worker | 3+ worker nodes（マルチAZ） |
| **Ingress** | NGINX Ingress (kind用) | AWS Load Balancer Controller + ALB |
| **DNS** | /etc/hosts 手動設定 | Route53 + ExternalDNS |
| **TLS** | なし | ACM証明書 + ALBでTLS終端 |
| **Web replicas** | 2 | 3+（HPA付き） |
| **API replicas** | 2 | 3+（HPA付き） |
| **Redis** | 単一Pod | Redis Cluster or ElastiCache |
| **DB** | なし | RDS（Multi-AZ） |
| **Secret管理** | base64 YAML | Secrets Manager + External Secrets |
| **ConfigMap管理** | 手動apply | GitOps (ArgoCD) |
| **ログ** | kubectl logs | Fluentbit → CloudWatch Logs |
| **メトリクス** | 自前Prometheus | Managed Prometheus / CloudWatch |
| **アラート** | なし | Alertmanager → Slack/PagerDuty |
| **トレース** | なし | X-Ray / Tempo |
| **デプロイ** | 手動 kubectl apply | GitHub Actions → ArgoCD |
| **ロールバック** | 手動 rollout undo | ArgoCD自動ロールバック |
| **Network Policy** | なし | Pod間通信を制限 |
| **Security Context** | デフォルト（root） | runAsNonRoot + readOnlyRootFilesystem |
| **イメージスキャン** | なし | Trivy / ECRスキャン |
| **RBAC** | 未設定 | 最小権限原則 |
| **バックアップ** | なし | Velero + RDSスナップショット |
| **コスト** | 無料 | ~$200-500/月（dev環境） |
| **可用性** | 保証なし | 99.9%+ SLA |
| **障害対応** | 手動調査 | Runbook + 自動復旧 |
| **負荷試験** | 未実施 | 定期的に実施 |

---

## 改善の優先順位

本番化に向けて改善すべき項目を優先度順に並べます。

### 優先度: 高（本番リリース前に必須）

1. **TLS/HTTPS対応** — 通信の暗号化は最低要件
2. **Secret暗号化** — base64は暗号化ではない
3. **Security Context** — rootでの実行を禁止
4. **ログ集約** — 障害時に調査できない
5. **アラート設定** — 障害に気づけない

### 優先度: 中（本番リリース後早期に対応）

6. **Redis冗長化** — SPOFの排除
7. **Network Policy** — 不要な通信を遮断
8. **HPA設定** — 負荷に応じた自動スケール
9. **CI/CDパイプライン** — デプロイの自動化
10. **バックアップ** — データ保護

### 優先度: 低（段階的に改善）

11. **分散トレーシング** — マイクロサービス間のボトルネック特定
12. **カオスエンジニアリング** — 障害耐性の検証
13. **コスト最適化** — Spot Instances, Savings Plans
14. **マルチクラスタ** — DR対策
15. **Service Mesh** — 通信の可観測性・制御

---

## Before構成の弱点まとめ

```
問題の深刻度:
🔴 致命的  → Secret平文、TLSなし、アラートなし
🟡 重要    → Redis SPOF、ログ集約なし、手動デプロイ
🟢 改善推奨 → Network Policy、トレーシング、カオスエンジニアリング
```

---

## 教訓

- **「動く」は出発点に過ぎない** — 本番運用には可用性・セキュリティ・運用性が必須
- **全部を一度に改善しない** — 優先度を付けて段階的に進める
- **コストと効果のバランス** — 過剰な冗長化はコストに跳ね返る
- **自動化が鍵** — 手動操作は事故のもと
