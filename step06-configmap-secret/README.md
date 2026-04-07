# Step 06: ConfigMap と Secret -- 設定をコンテナイメージから分離する

## 目的

アプリケーションの設定値や機密情報をコンテナイメージに焼き込まず、ConfigMap と Secret を使って外部から注入する方法を学ぶ。「同じイメージを環境ごとに異なる設定で動かす」という原則を実践する。

---

## 学ぶこと

- ConfigMap の作成と利用方法
- Secret の作成と利用方法
- 環境変数として注入するパターン（`env`）
- ファイルとしてマウントするパターン（`volumeMounts`）
- **Secret の base64 は暗号化ではない** という重要な事実

### ConfigMap と Secret の違い

| | ConfigMap | Secret |
|---|---|---|
| 用途 | 機密性のない設定値 | パスワード、APIキー等の機密情報 |
| 格納形式 | 平文 | base64 エンコード |
| 暗号化 | なし | **なし**（base64 は暗号化ではない!） |
| サイズ上限 | 1 MiB | 1 MiB |

> **重要**: Secret の値は base64 エンコードされているだけで、誰でもデコードできます。
> ```bash
> echo "cGFzc3dvcmQxMjM=" | base64 -d
> # => password123
> ```
> Secret は etcd 上で暗号化する設定（EncryptionConfiguration）や外部シークレット管理ツールと組み合わせて初めて安全になります。base64 = 安全と思い込まないでください。

### 注入パターンの比較

| パターン | 方法 | 適するケース |
|---|---|---|
| **環境変数** | `env.valueFrom.configMapKeyRef` / `secretKeyRef` | 個別の設定値（DB_HOST, LOG_LEVEL 等） |
| **ボリュームマウント** | `volumes` + `volumeMounts` | 設定ファイル丸ごと（config.json, nginx.conf 等） |

---

## ディレクトリ構成

```
step06-configmap-secret/
├── README.md          # このファイル
├── configmap.yaml     # ConfigMap の定義
├── secret.yaml        # Secret の定義
└── deployment.yaml    # ConfigMap/Secret を参照する Deployment
```

---

## 実行手順

### 1. ConfigMap を作成する

```bash
kubectl apply -f configmap.yaml
```

### 2. Secret を作成する

```bash
kubectl apply -f secret.yaml
```

### 3. ConfigMap と Secret を参照する Deployment を作成する

```bash
kubectl apply -f deployment.yaml
```

### 4. 環境変数が注入されているか確認する

```bash
kubectl exec -it deploy/config-demo -- env | grep -E "APP_ENV|DB_PASSWORD"
```

期待する出力:
```
APP_ENV=development
DB_PASSWORD=password123
```

### 5. ボリュームマウントされた設定ファイルを確認する

```bash
# マウントされたファイルの一覧を確認
kubectl exec -it deploy/config-demo -- ls /etc/app-config/

# 設定ファイルの内容を確認
kubectl exec -it deploy/config-demo -- cat /etc/app-config/config.json
```

---

## 確認方法

以下の全てが確認できれば、このステップは完了です。

1. `kubectl exec` で `APP_ENV=development` が環境変数に設定されていること
2. `kubectl exec` で `DB_PASSWORD=password123` が環境変数に設定されていること（base64 デコードされた値）
3. `/etc/app-config/` ディレクトリにファイルがマウントされていること
4. `/etc/app-config/config.json` の内容が読めること

---

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| Secret を平文で Git にコミットしてしまう | base64 を暗号化と勘違いしている | Secret の YAML は `.gitignore` に追加するか、Sealed Secrets / External Secrets を使う |
| コンテナイメージに `.env` を焼き込む | 設定変更のたびにイメージのリビルドが必要になる | ConfigMap / Secret で外部から注入する |
| `kubectl apply` で Secret の値が文字化けする | base64 エンコードが正しくない | `echo -n "平文" \| base64` で正しくエンコードする。`-n` を忘れると改行が含まれてしまう |
| ConfigMap/Secret が見つからず Pod が起動しない | ConfigMap/Secret の名前が Deployment の参照と一致していない | `configMapKeyRef.name` と `kubectl get configmap` の名前を照合する |
| ConfigMap を更新したのに環境変数が変わらない | 環境変数は Pod 起動時に読み込まれ、以降更新されない | `kubectl rollout restart deployment` で Pod を再起動する。ボリュームマウントの場合は自動反映される（最大約1分） |

---

## 本番だとどう変わるか

ローカルの kind 環境では Secret をそのまま YAML で管理しても問題ありませんが、本番環境では以下のツール・サービスを組み合わせます。

| ツール | 説明 |
|---|---|
| **External Secrets Operator** | AWS Secrets Manager, GCP Secret Manager, Azure Key Vault 等の外部ストアと連携し、Secret を自動同期する |
| **Sealed Secrets** | Bitnami の暗号化 Secret。暗号化された Secret を Git で安全に管理できる（GitOps 対応） |
| **HashiCorp Vault** | 動的シークレット、自動ローテーション、監査ログなど高度なシークレット管理を実現する |
| **AWS SSM Parameter Store** | 設定値を AWS Systems Manager で一元管理し、External Secrets 経由で Kubernetes に同期する |
| **AWS Secrets Manager** | 機密情報の保存・ローテーション・アクセス制御を提供するマネージドサービス |
| **etcd 暗号化** | EncryptionConfiguration で etcd に保存される Secret を暗号化する |

---

## まとめ

- ConfigMap は機密性のない設定値、Secret は機密情報に使う
- 環境変数注入とボリュームマウントの2つの方法がある
- **Secret の base64 は暗号化ではない** -- これを忘れないこと
- 本番では External Secrets, Vault, Secrets Manager 等と組み合わせる

---

## 次のステップ

Step 07 では Pod が消えてもデータを残す方法を学びます。PersistentVolume と PersistentVolumeClaim を使って、ステートフルなアプリケーションのデータ永続化に進みましょう。
