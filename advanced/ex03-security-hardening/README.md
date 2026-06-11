# ex03: セキュリティ強化と PodDisruptionBudget

## 目的

Step13 のレビューで指摘した「コンテナが root で実行されている」「PDB 未設定」を実際に解消する。
セキュリティ強化された Pod の作り方と、メンテナンス時の可用性保証を学ぶ。

## 学ぶこと

- `securityContext` による非root実行・権限の最小化
- `readOnlyRootFilesystem` と「書き込みが必要な場所だけ emptyDir」のパターン
- 非root化に伴うイメージ側の制約（1024未満のポートを bind できない等）
- PodDisruptionBudget（PDB）による退避時の可用性保証

### この演習で設定するセキュリティ項目

| 設定 | 効果 |
|---|---|
| `runAsNonRoot: true` | root での実行を禁止。コンテナ脱出時の被害を抑える |
| `readOnlyRootFilesystem: true` | ファイル改ざん・マルウェア設置を防ぐ |
| `allowPrivilegeEscalation: false` | setuid 等による権限昇格を防ぐ |
| `capabilities.drop: ["ALL"]` | 不要な Linux capability を全て剥奪する |

## 実行手順

### セキュリティ強化された Deployment

```bash
# 1. デプロイする
kubectl apply -f deployment.yaml
kubectl get pods -l app=hardened-web

# 2. 非root で動いていることを確認する（uid=101）
kubectl exec deploy/hardened-web -- id

# 3. ルートファイルシステムが読み取り専用であることを確認する
kubectl exec deploy/hardened-web -- touch /test-file
# → "Read-only file system" エラーになれば成功

# 4. /tmp（emptyDir）には書き込めることを確認する
kubectl exec deploy/hardened-web -- sh -c 'touch /tmp/test-file && echo "OK: /tmp は書き込み可能"'
```

### あえて違反させてみる

`runAsNonRoot: true` のまま root で動くイメージを使うとどうなるかを観察する。

```bash
# 通常の nginx（rootで起動する）を runAsNonRoot 環境で動かしてみる
kubectl run root-test --image=nginx:1.27 \
  --overrides='{"spec":{"securityContext":{"runAsNonRoot":true}}}'

# CreateContainerConfigError になることを確認する
kubectl get pod root-test
kubectl describe pod root-test | grep -A 3 "Events" | tail -3
# → "container has runAsNonRoot and image will run as root" というエラーが見える

kubectl delete pod root-test
```

### PodDisruptionBudget

```bash
# 1. PDB を作成する
kubectl apply -f pdb.yaml
kubectl get pdb hardened-web
# ALLOWED DISRUPTIONS が 1 になっている（2台中、同時に止めてよいのは1台）

# 2. Pod が動いているノードを確認する
kubectl get pods -l app=hardened-web -o wide

# 3. どちらかの Pod がいるノードを drain してみる（<node名> を置き換える）
kubectl drain <node名> --ignore-daemonsets --delete-emptydir-data

# 4. Pod が1台ずつ順番に退避され、常に1台は Ready が保たれることを観察する
kubectl get pods -l app=hardened-web -o wide -w

# 5. ノードを戻す
kubectl uncordon <node名>

# 6. クリーンアップ
kubectl delete -f pdb.yaml -f deployment.yaml
```

## 確認方法

1. `id` の結果が `uid=101` であること（rootでない）
2. `/` への書き込みが拒否され、`/tmp` には書き込めること
3. `runAsNonRoot` 違反の Pod が `CreateContainerConfigError` になること
4. drain 中も Ready な Pod が常に1台以上維持されること

## よくある失敗

| 症状 | 原因 | 対処 |
|---|---|---|
| `CreateContainerConfigError` | イメージが root で動く設計なのに `runAsNonRoot: true` を指定 | 非root対応イメージ（`-unprivileged` 等）を使うか、イメージを修正する |
| 起動直後にクラッシュする | `readOnlyRootFilesystem` でアプリの書き込み先が塞がれた | アプリが書き込むパスを特定し、emptyDir をマウントする |
| ポート80で listen できない | 非root ユーザーは 1024 未満のポートを bind できない | アプリを 8080 等で listen させ、Service で 80 → 8080 に変換する |
| drain が進まない | PDB の `minAvailable` がレプリカ数と同じで、1台も止められない | レプリカ数に対して退避の余地がある値を設定する（例: 2台なら minAvailable: 1） |
| PDB を設定して安心する | PDB は「自発的退避」しか守らない | ノード障害にはレプリカ数と分散配置（topologySpreadConstraints）で備える |

## 本番だとどう変わるか

- **Pod Security Admission**: namespace に `restricted` プロファイルを強制し、危険な Pod の作成自体を拒否する
- **ポリシーエンジン**: OPA Gatekeeper / Kyverno で「root禁止」「latestタグ禁止」等を組織ポリシーとして自動検査する
- **イメージスキャン**: Trivy 等を CI に組み込み、脆弱なイメージのデプロイを防ぐ
- **seccomp / AppArmor**: システムコールレベルでさらに制限を加える（`seccompProfile: RuntimeDefault`）
- **PDB はサービス全体で設計**: 全 Deployment に PDB を設定し、クラスタアップグレードを無停止で実施できるようにする
