.PHONY: cluster-create cluster-delete build-images load-images step01 step11 step15 step17 clean help

CLUSTER_NAME := k8s-learning

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

cluster-create: ## kindクラスタを作成
	kind create cluster --name $(CLUSTER_NAME) --config step01-kind-cluster/kind-config.yaml
	@echo "クラスタ作成完了。kubectl get nodes で確認してください。"

cluster-delete: ## kindクラスタを削除
	kind delete cluster --name $(CLUSTER_NAME)

build-images: ## 全アプリのDockerイメージをビルド
	docker build -t simple-api:latest apps/simple-api/
	docker build -t simple-web:latest apps/simple-web/
	docker build -t realtime-api:latest apps/realtime-api/

load-images: ## kindクラスタにイメージをロード
	kind load docker-image simple-api:latest --name $(CLUSTER_NAME)
	kind load docker-image simple-web:latest --name $(CLUSTER_NAME)
	kind load docker-image realtime-api:latest --name $(CLUSTER_NAME)

step01: ## Step01: kindクラスタ構築
	kind create cluster --name $(CLUSTER_NAME) --config step01-kind-cluster/kind-config.yaml

step11: ## Step11: ミニ構成をデプロイ
	# namespace を先に適用する（ディレクトリ一括適用はファイル名順のため、
	# namespace より先に他リソースが適用されて初回実行が失敗する）
	kubectl apply -f step11-mini-architecture/namespace.yaml
	kubectl apply -f step11-mini-architecture/

step15: ## Step15: WebSocket基礎をデプロイ
	kubectl apply -f step11-mini-architecture/namespace.yaml
	kubectl apply -f step15-websocket-basics/

step17: ## Step17: WebSocket負荷試験を実行
	cd apps/loadtest && k6 run ws-test.js

clean: ## クラスタ削除とイメージ削除
	kind delete cluster --name $(CLUSTER_NAME) 2>/dev/null || true
	docker rmi simple-api:latest simple-web:latest realtime-api:latest 2>/dev/null || true
	@echo "クリーンアップ完了"
