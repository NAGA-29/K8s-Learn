# simple-web

Nginx ベースの静的 Web フロントエンドです。
Kubernetes 学習用のフロントエンドアプリケーションとして使用します。

## 概要

- シンプルな HTML ページで `simple-api` からメッセージを取得・表示
- Nginx が静的ファイル配信と API へのリバースプロキシを担当
- `/api/*` へのリクエストは `simple-api:8080` に転送 (K8s Service 経由)

## ローカルでの実行

`index.html` をブラウザで直接開くか、任意の HTTP サーバーで配信してください。

```bash
# Python の簡易サーバーを使う例
python3 -m http.server 3000
```

※ ローカル単体では `/api/message` の取得に失敗します。API との結合は Kubernetes 環境で行います。

## Docker ビルド・実行

```bash
docker build -t simple-web .
docker run -p 3000:80 simple-web
```

http://localhost:3000 でアクセスできます。
