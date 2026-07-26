# Kubernetes（k8s）主要周辺エコシステム・ツール一覧

Kubernetesを本番環境で運用するために広く使われている、代表的なデファクトスタンダード（標準ツール）のまとめです。

---

## 1. デプロイ・GitOps（パッケージングと自動同期）
* **Helm**：Kubernetesの「App Store」。マニフェストをテンプレート化し、環境ごとの差異を管理。
* **Argo CD**：GitOpsを実現するCDツール。Gitリポジトリとクラスターの状態を常に自動同期。
* **Flux (Flux CD)**：Argo CDのライバル。GUIを持たない、シンプルかつKubernetesネイティブな設計。
* **Kustomize**：テンプレートを使わず、ベースのYAMLに対する「差分（Overlay）」で環境差分を管理。

## 2. ネットワークと通信（インフラ・セキュリティ）
* **NGINX Ingress Controller / Traefik**：外部からのアクセス（L7ルーティング）を受け付ける窓口。
* **Istio / Linkerd**：サービスメッシュ。マイクロサービス間の通信暗号化、リトライ、可視化を担当。
* **Cilium / Calico**：CNI（コンテナネットワークインターフェース）。Pod間の高速・安全な通信やポリシー制御。

## 3. オブザーバビリティ（監視・ログ・追跡）
* **Prometheus & Grafana**：メトリクス監視の王道。リソース使用量を収集し、ダッシュボードで視覚化。
* **Loki / Fluent Bit / Fluentd**：ログ収集・管理。コンテナから出力される膨大なログを集約。
* **Jaeger / OpenTelemetry**：分散トレーシング。リクエストがどのサービスをどう通過したかを追跡。

## 4. セキュリティ・ガバナンス
* **Kyverno / OPA Gatekeeper**：ポリシー管理。「rootユーザーでのコンテナ起動禁止」などを強制。
* **Trivy / Aqua Security**：脆弱性スキャン。イメージやマニフェストのセキュリティ欠陥を自動検査。
* **Cert-Manager**：TLS/SSL証明書の自動発行・更新（Let's Encryptなどと連携）。

## 5. ストレージ・バックアップ
* **Rook (Ceph) / Longhorn**：永続ストレージ（CSI）。データベースなどのデータを安全に保存。
* **Velero**：バックアップ・リストア。クラスターの状態や永続データをS3などに退避。

## 6. シークレット管理・CI/CD
* **External Secrets Operator (ESO)**：AWS/GCP等のシークレットマネージャーとk8s Secretを安全に同期。
* **Tekton / GitHub Actions**：CI/CD（ビルド・テスト）。Argo CD（デプロイ）の前段となるイメージ作成を担当。

## 7. 開発者・運用者支援（CLI/GUIツール）
* **K9s**：ターミナル用UIツール。コマンド（kubectl）不要で、ログ確認やPod削除が爆速。
* **Lens**：デスクトップ用GUI。クラスター全体の稼働状況を視覚的に確認・操作。
