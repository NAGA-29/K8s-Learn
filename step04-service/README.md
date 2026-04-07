# Step 04: Service -- Pod への安定したアクセス経路

## 目的

Pod の IP は一時的なものであり、直接アクセスすべきではないことを理解する。Service を使って、Pod 群への安定したアクセス経路を確立する方法を学ぶ。

## 学ぶこと

- なぜ Pod IP に直接アクセスしてはいけないのか
- Service の仕組み (ラベルセレクタによる Pod の紐づけ)
- ClusterIP と NodePort の違い
- Endpoints の役割

### ClusterIP と NodePort の違い

| タイプ | アクセス範囲 | 用途 |
|---|---|---|
| **ClusterIP** (デフォルト) | クラスタ内部のみ | Pod 間通信。外部からはアクセスできない |
| **NodePort** | ノードの IP + 指定ポート | 開発・検証用。全ノードの指定ポートでアクセス可能 |
| **LoadBalancer** | 外部ロードバランサ経由 | 本番のクラウド環境で使用 |

## ディレクトリ構成

```
step04-service/
├── README.md
└── service.yaml
```

## 前提条件

Step 03 の Deployment (`nginx-deployment`) がクラスタ上で動作していること。

```bash
kubectl get deployments
# nginx-deployment が READY 3/3 であることを確認
```

## 実行手順

```bash
# 1. Service を作成する
kubectl apply -f service.yaml

# 2. Service の状態を確認する
kubectl get svc nginx-service

# 3. Endpoints を確認する (Pod の IP が紐づいているか)
kubectl get endpoints nginx-service

# 4. Service の詳細を確認する
kubectl describe svc nginx-service

# 5. クラスタ内からアクセスを確認する
kubectl run curl-test --image=curlimages/curl --rm -it --restart=Never -- curl http://nginx-service:80

# 6. NodePort 経由でアクセスを確認する (kind の場合)
# kind ではノードが Docker コンテナなので、Docker 経由で確認する
docker exec k8s-learning-worker curl -s http://localhost:30080
```

## 確認方法

- `kubectl get endpoints nginx-service` で、Deployment の Pod IP が Endpoints に表示されていること
- Pod を増減させると Endpoints も自動的に更新されること
- curl でレスポンスが返ってくること

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| Endpoints が `<none>` になる | Service の `selector` と Pod の `labels` が一致していない | `kubectl get pods --show-labels` でラベルを確認し、Service の selector と合わせる |
| Pod IP に直接 curl する | Pod 再作成で IP が変わるため、アクセスできなくなる | 必ず Service 名経由でアクセスする |
| NodePort にアクセスできない | kind の場合、ホストから直接 NodePort にアクセスできない場合がある | `docker exec` 経由で worker コンテナ内から確認する |
| Service を削除しても Pod が消えると勘違い | Service と Pod は独立したリソース | Service 削除は経路が消えるだけ。Pod は Deployment が管理している |

## 本番だとどう変わるか

- **ClusterIP が主役**: Pod 間通信は ClusterIP で行い、NodePort は基本的に使わない
- **外部公開は Ingress / LoadBalancer**: ユーザ向けのエンドポイントには次の Step で学ぶ Ingress か、LoadBalancer タイプの Service を使う
- **Service Mesh**: Istio や Linkerd を導入し、サービス間通信の暗号化 (mTLS)、トラフィック制御、可観測性を実現する
- **DNS**: クラスタ内 DNS により `<service-name>.<namespace>.svc.cluster.local` で名前解決できる
- **ExternalName**: クラスタ外のサービス (RDS 等) を Kubernetes の Service として抽象化することもある

---

Service で Pod への安定したアクセス経路が作れるようになった。しかし NodePort はポート番号を覚える必要があり、パスベースのルーティングもできない。次の Step 05 では、HTTP レイヤでの柔軟なルーティングを実現する **Ingress** を学ぶ。
