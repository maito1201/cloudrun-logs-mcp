package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/maito1201/cloudrun-logs-mcp/logs"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "cloudrun-logs",
		Usage: "Google Cloud Runのログを取得するCLIツール",
		Commands: []*cli.Command{
			{
				Name:   "logs",
				Usage:  "Cloud Runのログを取得",
				Flags:  getLogsFlags(),
				Action: getLogsAction,
			},
			{
				Name:  "services",
				Usage: "Cloud Runのサービス一覧を取得",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project",
						Aliases:  []string{"p"},
						Usage:    "Google Cloudプロジェクトのプロジェクトid",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "region",
						Aliases: []string{"r"},
						Usage:   "Cloud Runのリージョン（デフォルト: us-central1）",
						Value:   "us-central1",
					},
					&cli.BoolFlag{
						Name:    "json",
						Aliases: []string{"j"},
						Usage:   "サービス一覧をJSON形式で出力",
						Value:   false,
					},
				},
				Action: func(c *cli.Context) error {
					// サービス一覧を取得
					ctx := context.Background()
					services, err := logs.GetCloudRunServices(ctx, c.String("project"), c.String("region"))
					if err != nil {
						return fmt.Errorf("サービス一覧の取得エラー: %v", err)
					}

					// 結果を表示
					if c.Bool("json") {
						// JSON形式で出力
						jsonData, err := json.MarshalIndent(services, "", "  ")
						if err != nil {
							return fmt.Errorf("JSONエンコードエラー: %v", err)
						}
						fmt.Println(string(jsonData))
					} else {
						// テキスト形式で出力
						if len(services) == 0 {
							fmt.Println("サービスが見つかりませんでした。")
							return nil
						}

						fmt.Printf("プロジェクト %s のCloud Runサービス一覧（リージョン: %s）:\n\n", c.String("project"), c.String("region"))

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

						fmt.Printf("合計 %d 件のサービスを表示しました。\n", len(services))
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// getLogsFlags はログ取得コマンドのフラグを返します
func getLogsFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "project",
			Aliases:  []string{"p"},
			Usage:    "Google Cloudプロジェクトのプロジェクトid",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "service",
			Aliases: []string{"s"},
			Usage:   "Cloud Runのサービス名",
		},
		&cli.StringFlag{
			Name:    "start-time",
			Aliases: []string{"st"},
			Usage:   "ログの開始時間（RFC3339形式、例: 2023-01-01T00:00:00Z）",
		},
		&cli.StringFlag{
			Name:    "end-time",
			Aliases: []string{"et"},
			Usage:   "ログの終了時間（RFC3339形式、例: 2023-01-01T00:00:00Z）",
		},
		&cli.StringFlag{
			Name:    "level",
			Aliases: []string{"l"},
			Usage:   "ログレベル（INFO, ERROR, WARNINGなど）",
		},
		&cli.StringSliceFlag{
			Name:    "keyword",
			Aliases: []string{"k"},
			Usage:   "検索キーワード（複数指定可）",
		},
		&cli.IntFlag{
			Name:    "limit",
			Aliases: []string{"n"},
			Usage:   "取得するログエントリの最大数",
			Value:   100,
		},
		&cli.BoolFlag{
			Name:    "json",
			Aliases: []string{"j"},
			Usage:   "ログをJSON形式で出力",
			Value:   false,
		},
	}
}

// getLogsAction はログ取得コマンドのアクションを返します
var getLogsAction = func(c *cli.Context) error {
	// フィルターオプションを構築
	opts := logs.FilterOptions{
		ProjectID:   c.String("project"),
		ServiceName: c.String("service"),
		LogLevel:    c.String("level"),
		Keywords:    c.StringSlice("keyword"),
		Limit:       c.Int("limit"),
	}

	// 開始時間が指定されている場合
	if c.String("start-time") != "" {
		startTime, err := time.Parse(time.RFC3339, c.String("start-time"))
		if err != nil {
			return fmt.Errorf("開始時間の解析エラー: %v", err)
		}
		opts.StartTime = startTime
	}

	// 終了時間が指定されている場合
	if c.String("end-time") != "" {
		endTime, err := time.Parse(time.RFC3339, c.String("end-time"))
		if err != nil {
			return fmt.Errorf("終了時間の解析エラー: %v", err)
		}
		opts.EndTime = endTime
	}

	// ログを取得
	ctx := context.Background()
	entries, err := logs.GetCloudRunLogs(ctx, opts)
	if err != nil {
		return fmt.Errorf("ログの取得エラー: %v", err)
	}

	// 結果を表示
	if c.Bool("json") {
		// JSON形式で出力
		jsonData, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return fmt.Errorf("JSONエンコードエラー: %v", err)
		}
		fmt.Println(string(jsonData))
	} else {
		// テキスト形式で出力
		if len(entries) == 0 {
			fmt.Println("ログエントリが見つかりませんでした。")
			return nil
		}

		for _, entry := range entries {
			fmt.Printf("[%s] %s: %s\n", entry.Timestamp.Format(time.RFC3339), entry.Severity, entry.Message)

			// ラベルがある場合は表示
			if len(entry.Labels) > 0 {
				labelStrs := []string{}
				for k, v := range entry.Labels {
					labelStrs = append(labelStrs, fmt.Sprintf("%s=%s", k, v))
				}
				fmt.Printf("  Labels: %s\n", strings.Join(labelStrs, ", "))
			}

			fmt.Println()
		}

		fmt.Printf("合計 %d 件のログエントリを表示しました。\n", len(entries))
	}

	return nil
}
