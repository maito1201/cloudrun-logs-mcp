# Cloud Run Logs CLI

Google Cloud RunのログをCLIから取得するためのツールです。

## 機能

- Google Cloudのアプリケーションデフォルトクレデンシャルを使用して認証
- プロジェクトID、サービス名、時間範囲、ログレベル、キーワードなどでフィルタリング
- テキスト形式またはJSON形式での出力

## インストール

```bash
go install github.com/ito-masahiko/cloudrun-logs-mcp@latest
```

または、リポジトリをクローンしてビルド：

```bash
git clone https://github.com/ito-masahiko/cloudrun-logs-mcp.git
cd cloudrun-logs-mcp
go build -o cloudrun-logs
```

## 使い方

このツールには、以下の2つのコマンドがあります：

- `logs`: Cloud Runのログを取得
- `services`: Cloud Runのサービス一覧を取得

### ログの取得

```bash
# 基本的な使い方（プロジェクトIDは必須）
./cloudrun-logs logs --project=your-project-id

# サービス名を指定
./cloudrun-logs logs --project=your-project-id --service=your-service-name

# 時間範囲を指定
./cloudrun-logs logs --project=your-project-id --start-time=2023-01-01T00:00:00Z --end-time=2023-01-02T00:00:00Z

# ログレベルを指定
./cloudrun-logs logs --project=your-project-id --level=ERROR

# キーワードで検索
./cloudrun-logs logs --project=your-project-id --keyword=error --keyword=exception

# 取得するログエントリの最大数を指定
./cloudrun-logs logs --project=your-project-id --limit=50

# JSON形式で出力
./cloudrun-logs logs --project=your-project-id --json
```

### サービス一覧の取得

```bash
# 基本的な使い方（プロジェクトIDは必須）
./cloudrun-logs services --project=your-project-id

# リージョンを指定
./cloudrun-logs services --project=your-project-id --region=us-central1

# JSON形式で出力
./cloudrun-logs services --project=your-project-id --json
```

### オプション一覧

#### logsコマンドのオプション

| オプション | 短縮形 | 説明 | 必須 | デフォルト値 |
|------------|--------|------|------|-------------|
| --project | -p | Google Cloudプロジェクトのプロジェクトid | はい | - |
| --service | -s | Cloud Runのサービス名 | いいえ | - |
| --start-time | -st | ログの開始時間（RFC3339形式） | いいえ | - |
| --end-time | -et | ログの終了時間（RFC3339形式） | いいえ | - |
| --level | -l | ログレベル（INFO, ERROR, WARNINGなど） | いいえ | - |
| --keyword | -k | 検索キーワード（複数指定可） | いいえ | - |
| --limit | -n | 取得するログエントリの最大数 | いいえ | 100 |
| --json | -j | ログをJSON形式で出力 | いいえ | false |

#### servicesコマンドのオプション

| オプション | 短縮形 | 説明 | 必須 | デフォルト値 |
|------------|--------|------|------|-------------|
| --project | -p | Google Cloudプロジェクトのプロジェクトid | はい | - |
| --region | -r | Cloud Runのリージョン | いいえ | us-central1 |
| --json | -j | サービス一覧をJSON形式で出力 | いいえ | false |

## 前提条件

- Go 1.16以上
- Google Cloudのアプリケーションデフォルトクレデンシャルが設定されていること

## 認証

このツールは、Google Cloudのアプリケーションデフォルトクレデンシャルを使用して認証を行います。以下のいずれかの方法で認証情報を設定してください：

1. `gcloud auth application-default login`コマンドを実行
2. GOOGLE_APPLICATION_CREDENTIALS環境変数にサービスアカウントキーのパスを設定
3. Google Cloud環境（Compute Engine、Cloud Run、GKEなど）で実行する場合は、自動的に認証情報が提供されます

## ライブラリとしての使用

このツールは、ライブラリとしても使用できます。以下は使用例です：

### ログの取得

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ito-masahiko/cloudrun-logs-mcp/pkg/logs"
)

func main() {
	// フィルターオプションを設定
	opts := logs.FilterOptions{
		ProjectID:   "your-project-id",
		ServiceName: "your-service-name",
		StartTime:   time.Now().Add(-24 * time.Hour), // 24時間前から
		LogLevel:    "ERROR",
		Keywords:    []string{"error", "exception"},
		Limit:       50,
	}

	// ログを取得
	ctx := context.Background()
	entries, err := logs.GetCloudRunLogs(ctx, opts)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		return
	}

	// 結果を処理
	for _, entry := range entries {
		fmt.Printf("[%s] %s: %s\n", entry.Timestamp.Format(time.RFC3339), entry.Severity, entry.Message)
	}
}
```

### サービス一覧の取得

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ito-masahiko/cloudrun-logs-mcp/pkg/logs"
)

func main() {
	// プロジェクトIDとリージョンを指定
	projectID := "your-project-id"
	region := "us-central1"

	// サービス一覧を取得
	ctx := context.Background()
	services, err := logs.GetCloudRunServices(ctx, projectID, region)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		return
	}

	// 結果を処理
	for _, service := range services {
		fmt.Printf("名前: %s\n", service.Name)
		if service.Description != "" {
			fmt.Printf("説明: %s\n", service.Description)
		}
		fmt.Printf("URL: %s\n", service.URL)
		fmt.Printf("ステータス: %s\n", service.Status)
		fmt.Printf("作成日時: %s\n", service.CreateTime.Format(time.RFC3339))
		fmt.Printf("更新日時: %s\n", service.UpdateTime.Format(time.RFC3339))
		fmt.Println()
	}
}
```

## 将来の拡張予定

- MCPサーバーとしての機能拡張
- より詳細なフィルタリングオプションの追加
- 出力形式のカスタマイズ

## ライセンス

MIT
