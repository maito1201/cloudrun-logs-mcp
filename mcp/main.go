package main

import (
	"context"
	"fmt"
	"time"

	"github.com/maito1201/cloudrun-logs-mcp/logs"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// MCPサーバーを作成
	mcpServer := server.NewMCPServer(
		"cloudrun-logs-mcp",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// ログ取得ツールを登録
	mcpServer.AddTool(getLogsTool(), getLogsHandler)

	// サービス一覧取得ツールを登録
	mcpServer.AddTool(getServicesTool(), getServicesHandler)

	// サーバー起動
	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// getLogsTool はログ取得ツールを定義します
func getLogsTool() mcp.Tool {
	return mcp.NewTool("get_logs",
		mcp.WithDescription("Google Cloud Runのログを取得します"),
		mcp.WithString("project_id", mcp.Description("Google Cloudプロジェクトのプロジェクトid"), mcp.Required()),
		mcp.WithString("service_name", mcp.Description("Cloud Runのサービス名（省略可）")),
		mcp.WithString("start_time", mcp.Description("ログの開始時間（RFC3339形式、例: 2023-01-01T00:00:00Z）（省略可）")),
		mcp.WithString("end_time", mcp.Description("ログの終了時間（RFC3339形式、例: 2023-01-01T00:00:00Z）（省略可）")),
		mcp.WithString("log_level", mcp.Description("ログレベル（INFO, ERROR, WARNINGなど）（省略可）")),
		mcp.WithArray("keywords", mcp.Description("検索キーワード（省略可）"), mcp.Items(map[string]interface{}{
			"type": "string",
		})),
		mcp.WithNumber("limit", mcp.Description("取得するログエントリの最大数（省略可、デフォルト: 100）")),
	)
}

// getLogsHandler はログ取得ツールのハンドラーです
func getLogsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	// パラメータを取得
	projectID, _ := args["project_id"].(string)
	serviceName, _ := args["service_name"].(string)
	logLevel, _ := args["log_level"].(string)
	limit := 100
	if limitVal, ok := args["limit"].(float64); ok {
		limit = int(limitVal)
	}

	// キーワードを取得
	var keywords []string
	if keywordsVal, ok := args["keywords"].([]interface{}); ok {
		for _, k := range keywordsVal {
			if keyword, ok := k.(string); ok {
				keywords = append(keywords, keyword)
			}
		}
	}

	// フィルターオプションを構築
	opts := logs.FilterOptions{
		ProjectID:   projectID,
		ServiceName: serviceName,
		LogLevel:    logLevel,
		Keywords:    keywords,
		Limit:       limit,
	}

	// 開始時間が指定されている場合
	if startTimeStr, ok := args["start_time"].(string); ok && startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			return nil, fmt.Errorf("開始時間の解析エラー: %v", err)
		}
		opts.StartTime = startTime
	}

	// 終了時間が指定されている場合
	if endTimeStr, ok := args["end_time"].(string); ok && endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			return nil, fmt.Errorf("終了時間の解析エラー: %v", err)
		}
		opts.EndTime = endTime
	}

	// ログを取得
	entries, err := logs.GetCloudRunLogs(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("ログの取得エラー: %v", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", entries)), nil
}

// getServicesTool はサービス一覧取得ツールを定義します
func getServicesTool() mcp.Tool {
	return mcp.NewTool("get_services",
		mcp.WithDescription("Google Cloud Runのサービス一覧を取得します"),
		mcp.WithString("project_id", mcp.Description("Google Cloudプロジェクトのプロジェクトid"), mcp.Required()),
		mcp.WithString("region", mcp.Description("Cloud Runのリージョン（省略可、デフォルト: us-central1）")),
	)
}

// getServicesHandler はサービス一覧取得ツールのハンドラーです
func getServicesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	// パラメータを取得
	projectID, _ := args["project_id"].(string)
	region := "us-central1"
	if regionVal, ok := args["region"].(string); ok && regionVal != "" {
		region = regionVal
	}

	// サービス一覧を取得
	services, err := logs.GetCloudRunServices(ctx, projectID, region)
	if err != nil {
		return nil, fmt.Errorf("サービス一覧の取得エラー: %v", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("%v", services)), nil
}
