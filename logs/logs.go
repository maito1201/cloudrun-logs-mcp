package logs

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/api/run/v1"
)

// ServiceInfo はCloud Runのサービス情報を表します
type ServiceInfo struct {
	Name        string
	Description string
	URL         string
	Status      string
	Region      string
	CreateTime  time.Time
	UpdateTime  time.Time
}

// GetCloudRunServices はプロジェクトIDを指定してCloud Runのサービス一覧を取得します
func GetCloudRunServices(ctx context.Context, projectID string, region string) ([]ServiceInfo, error) {
	// Cloud Run APIクライアントを作成
	runService, err := run.NewService(ctx, option.WithScopes(run.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("run.NewService: %v", err)
	}

	// リージョンが指定されていない場合はデフォルトのリージョンを使用
	if region == "" {
		region = "us-central1" // デフォルトのリージョン
	}

	// サービス一覧を取得
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)
	resp, err := runService.Projects.Locations.Services.List(parent).Do()
	if err != nil {
		return nil, fmt.Errorf("runService.Projects.Locations.Services.List: %v", err)
	}

	// 結果を処理
	services := []ServiceInfo{}
	for _, item := range resp.Items {
		createTime, _ := time.Parse(time.RFC3339, item.Metadata.CreationTimestamp)
		// 更新時間が取得できない場合は作成時間を使用
		updateTime := createTime
		if item.Metadata.Annotations != nil && item.Metadata.Annotations["client.knative.dev/user-image"] != "" {
			// 最終デプロイ時間の代わりに現在時刻を使用
			updateTime = time.Now()
		}

		service := ServiceInfo{
			Name:        item.Metadata.Name,
			Description: item.Metadata.Annotations["description"],
			URL:         item.Status.Url,
			Status:      item.Status.Conditions[0].Status,
			Region:      region,
			CreateTime:  createTime,
			UpdateTime:  updateTime,
		}
		services = append(services, service)
	}

	return services, nil
}

// FilterOptions は、ログのフィルタリングオプションを定義します
type FilterOptions struct {
	ProjectID   string    // Google Cloudプロジェクトのプロジェクトid
	ServiceName string    // Cloud Runのサービス名
	StartTime   time.Time // ログの開始時間
	EndTime     time.Time // ログの終了時間
	LogLevel    string    // ログレベル（INFO, ERROR, WARNINGなど）
	Keywords    []string  // 検索キーワード
	Limit       int       // 取得するログエントリの最大数
}

// LogEntry はログエントリを表します
type LogEntry struct {
	Timestamp time.Time
	Severity  string
	Message   string
	Labels    map[string]string
}

// GetCloudRunLogs は指定されたフィルターオプションに基づいてCloud Runのログを取得します
func GetCloudRunLogs(ctx context.Context, opts FilterOptions) ([]LogEntry, error) {
	// Logging Adminクライアントを作成
	client, err := logadmin.NewClient(ctx, opts.ProjectID, option.WithScopes("https://www.googleapis.com/auth/logging.read"))
	if err != nil {
		return nil, fmt.Errorf("logadmin.NewClient: %v", err)
	}
	defer client.Close()

	// フィルター文字列を構築
	filter := buildFilter(opts)

	// ログエントリを取得
	entries := []LogEntry{}
	iter := client.Entries(ctx, logadmin.Filter(filter))

	// 結果を処理
	count := 0
	for {
		entry, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("iter.Next: %v", err)
		}

		// LogEntryに変換
		logEntry := LogEntry{
			Timestamp: entry.Timestamp,
			Severity:  entry.Severity.String(),
			Labels:    entry.Labels,
		}

		// メッセージを取得（ペイロードの種類によって処理が異なる）
		switch p := entry.Payload.(type) {
		case string:
			logEntry.Message = p
		case map[string]interface{}:
			if msg, ok := p["message"].(string); ok {
				logEntry.Message = msg
			} else {
				// JSONとしてフォーマット
				logEntry.Message = fmt.Sprintf("%v", p)
			}
		default:
			logEntry.Message = fmt.Sprintf("%v", p)
		}

		entries = append(entries, logEntry)

		count++
		if opts.Limit > 0 && count >= opts.Limit {
			break
		}
	}

	return entries, nil
}

// buildFilter はフィルターオプションからフィルター文字列を構築します
func buildFilter(opts FilterOptions) string {
	filter := fmt.Sprintf("resource.type=\"cloud_run_revision\"")

	// プロジェクトIDは既にクライアント作成時に指定されているため、フィルターには含めない

	// サービス名が指定されている場合
	if opts.ServiceName != "" {
		filter += fmt.Sprintf(" AND resource.labels.service_name=\"%s\"", opts.ServiceName)
	}

	// 時間範囲が指定されている場合
	if !opts.StartTime.IsZero() {
		filter += fmt.Sprintf(" AND timestamp>=\"%s\"", opts.StartTime.Format(time.RFC3339))
	}
	if !opts.EndTime.IsZero() {
		filter += fmt.Sprintf(" AND timestamp<=\"%s\"", opts.EndTime.Format(time.RFC3339))
	}

	// ログレベルが指定されている場合
	if opts.LogLevel != "" {
		filter += fmt.Sprintf(" AND severity>=\"%s\"", opts.LogLevel)
	}

	// キーワードが指定されている場合
	for _, keyword := range opts.Keywords {
		if keyword != "" {
			filter += fmt.Sprintf(" AND textPayload:\"%s\"", keyword)
		}
	}

	return filter
}
