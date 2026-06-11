# 応用編 — Step17 のその先へ

Step01〜17 で学んだ基礎の上に、本番運用で必要になるテーマを演習形式で学ぶ。
Step13 のアーキテクチャレビューで「改善案」として挙がった項目を、実際に手を動かして体験することが目的である。

本編の Step とは異なり、各演習は独立しており順不同で取り組める。

## 前提条件

- Step01〜17 を完了していること（最低でも Step11 まで）
- kind クラスタ（`k8s-learning`）が動作していること
- ex04 は Step11 の mini-app 構成がデプロイ済みであること

## 演習一覧

| 演習 | タイトル | 学ぶこと | 関連する本編Step |
|------|----------|----------|-----------------|
| [ex01](ex01-rolling-update/) | ローリングアップデートとロールバック | 無停止デプロイ、`maxSurge`/`maxUnavailable`、`rollout undo` | Step03, Step08 |
| [ex02](ex02-job-cronjob/) | Job / CronJob | バッチ処理、リトライ制御、定期実行 | Step02 |
| [ex03](ex03-security-hardening/) | セキュリティ強化と PDB | 非root実行、readOnlyRootFilesystem、PodDisruptionBudget | Step13 |
| [ex04](ex04-networkpolicy/) | NetworkPolicy | Pod間通信の最小権限化、default-deny | Step11, Step13 |

## ディレクトリ構成

```
advanced/
├── README.md                    # このファイル
├── ex01-rolling-update/
│   ├── README.md
│   ├── deployment-v1.yaml
│   └── deployment-v2.yaml
├── ex02-job-cronjob/
│   ├── README.md
│   ├── job.yaml
│   └── cronjob.yaml
├── ex03-security-hardening/
│   ├── README.md
│   ├── deployment.yaml
│   └── pdb.yaml
└── ex04-networkpolicy/
    ├── README.md
    └── networkpolicy.yaml
```

## この先の自学テーマ

応用編を終えたら、Step17 の末尾でも紹介している以下のテーマに進むことを推奨する。

- **Helm / Kustomize**: 環境差分の管理とマニフェストのテンプレート化
- **GitOps (ArgoCD / Flux)**: Git を信頼の源とした宣言的デプロイ
- **Service Mesh (Istio / Linkerd)**: mTLS、トラフィック制御、可観測性
- **ポリシーエンジン (OPA / Kyverno)**: マニフェストのガバナンス自動化
- **Chaos Engineering**: 意図的な障害注入による耐障害性の検証
