# Step 05: Ingress -- HTTP レイヤのルーティング

## 目的

NodePort のようなポート番号ベースではなく、ホスト名やパスによる HTTP レイヤ (L7) のルーティングを Ingress で実現する方法を学ぶ。

## 学ぶこと

- L4 (Transport) と L7 (Application) ロードバランシングの違い
- Ingress リソースと Ingress Controller の関係
- ホストベース / パスベースのルーティング
- kind 環境での Ingress Controller のセットアップ

### L4 と L7 の違い

| レイヤ | 対象 | 判断基準 | 例 |
|---|---|---|---|
| **L4 (Transport)** | TCP/UDP | IP アドレスとポート番号 | Service (NodePort, LoadBalancer) |
| **L7 (Application)** | HTTP/HTTPS | ホスト名、パス、ヘッダー | Ingress |

L4 は「どのポートに来たか」でしか振り分けられないが、L7 は「どの URL に来たか」で振り分けられる。1 つのポート (80/443) で複数のサービスにルーティングできるのが Ingress の強みである。

### Ingress Controller とは

Ingress リソースはあくまで「ルーティングルールの定義」であり、それ自体は何もしない。ルールを実際に処理するのが Ingress Controller である。代表的な実装として NGINX Ingress Controller, Traefik, HAProxy などがある。Ingress Controller がクラスタにデプロイされていなければ、Ingress リソースを作成しても機能しない。

## ディレクトリ構成

```
step05-ingress/
├── README.md
└── ingress.yaml
```

## 前提条件

- Step 03 の Deployment (`nginx-deployment`) がクラスタ上で動作していること
- Step 04 の Service (`nginx-service`) がクラスタ上で動作していること

```bash
kubectl get deployments
# nginx-deployment が READY 3/3 であることを確認

kubectl get svc nginx-service
# nginx-service が存在することを確認
```

## 実行手順

```bash
# 1. NGINX Ingress Controller を kind 用にインストールする
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# 2. Ingress Controller の Pod が Ready になるまで待つ
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

# 3. /etc/hosts にエントリを追加する (sudo が必要)
echo "127.0.0.1 k8s-learning.local" | sudo tee -a /etc/hosts

# 4. Ingress リソースを作成する
kubectl apply -f ingress.yaml

# 5. Ingress の状態を確認する
kubectl get ingress

# 6. Ingress の詳細を確認する
kubectl describe ingress nginx-ingress

# 7. ブラウザまたは curl でアクセスする
curl http://k8s-learning.local
```

## 確認方法

- `kubectl get ingress` で ADDRESS が設定されていること
- `kubectl describe ingress nginx-ingress` の Rules セクションに正しい host, path, backend が表示されていること
- `curl http://k8s-learning.local` で NGINX のデフォルトページが返ること
- ブラウザで `http://k8s-learning.local` にアクセスして NGINX のウェルカムページが表示されること

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| Ingress を作成しても何も起きない | Ingress Controller がインストールされていない | 手順 1 で NGINX Ingress Controller をインストールする |
| `curl: (6) Could not resolve host` | `/etc/hosts` にエントリが追加されていない | 手順 3 で `/etc/hosts` を編集する |
| `curl: (7) Failed to connect` | Ingress Controller の Pod がまだ起動していない | `kubectl get pods -n ingress-nginx` で状態を確認し、Ready になるまで待つ |
| 404 Not Found が返る | Ingress の `backend.service.name` と実際の Service 名が一致していない | `kubectl get svc` で Service 名を確認し、ingress.yaml を修正する |
| kind でポート 80 にアクセスできない | kind-config.yaml で `extraPortMappings` を設定していない | Step 01 の kind-config.yaml にポートマッピングが含まれているか確認する |

## 本番だとどう変わるか

- **クラウド用 Ingress Controller**: AWS では ALB Ingress Controller (AWS Load Balancer Controller)、GCP では GCE Ingress Controller を使い、クラウドネイティブな LB と統合する
- **TLS 終端**: cert-manager と組み合わせて Let's Encrypt 等の証明書を自動取得・更新し、HTTPS を実現する
- **Gateway API**: Ingress の後継として策定が進む Gateway API は、より柔軟で拡張性の高いルーティング定義ができる。今後の標準となる見込み
- **WAF / レート制限**: Ingress Controller の設定やアノテーションで、Web Application Firewall やレート制限を実装する
- **複数 Ingress の統合**: 本番では複数のサービスを 1 つの Ingress にまとめ、パスベースやホストベースで振り分けるのが一般的

---

次のステップでは、アプリケーションの設定情報や機密情報をコンテナイメージから分離する方法を学ぶ。**ConfigMap** と **Secret** を使った設定管理を Step 06 で体験しよう。
