// @Title        s3obs
// @Description  s3_obs_manager
// @Create       dingshuai 2025/9/25 13:37

package s3obs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var contentTypes = map[string]string{
	// 图片类型
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".webp": "image/webp",
	".svg":  "image/svg+xml",
	".ico":  "image/x-icon",

	// 文档类型
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".txt":  "text/plain",
	".csv":  "text/csv",

	// 压缩文件
	".zip": "application/zip",
	".rar": "application/x-rar-compressed",
	".7z":  "application/x-7z-compressed",
	".tar": "application/x-tar",
	".gz":  "application/gzip",

	// 音频视频
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".mp4":  "video/mp4",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".webm": "video/webm",

	// 代码文件
	".html": "text/html",
	".htm":  "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",

	// 字体文件
	".ttf":   "font/ttf",
	".otf":   "font/otf",
	".woff":  "font/woff",
	".woff2": "font/woff2",

	// 其他
	".bin": "application/octet-stream",
	".exe": "application/octet-stream",
	".dmg": "application/octet-stream",
}

// VideoRecord 录像文件信息
type VideoRecord struct {
	Key          string    `json:"key"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	URL          string    `json:"url"`
	StorageClass string    `json:"storage_class"`
}

// S3Config S3配置
type S3Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	SmsNode         string
	Expires         time.Duration
	ForcePathStyle  bool // 用于MinIO等兼容S3的服务

}

// S3OBSManager S3对象存储管理模块
type S3OBSManager struct {
	client   *s3.Client
	uploader *manager.Uploader
	// downloader  *manager.Downloader
	config      *S3Config
	uploadQueue *UploadQueue
	queueConfig *UploadQueueConfig
}

// NewS3OBSManager 创建S3管理实例
func NewS3OBSManager(s3Config *S3Config, queueConfig *UploadQueueConfig) (*S3OBSManager, error) {
	if queueConfig == nil {
		queueConfig = &UploadQueueConfig{
			MaxConcurrentUploads: 6,
			MaxQueueSize:         1000,
			DefaultTimeout:       1 * time.Minute,
			DefaultRetries:       3,
			RetryDelay:           3 * time.Second,
			PriorityEnabled:      true,
			ResultBufferSize:     100,
		}
	}

	return &S3OBSManager{
		config:      s3Config,
		queueConfig: queueConfig,
		uploadQueue: NewUploadQueue(queueConfig),
	}, nil
}

// NewS3OBSManagerWithAWS 使用AWS默认配置创建管理器（用于AWS S3）
func NewS3OBSManagerWithAWS(bucketName, region string, queueConfig *UploadQueueConfig) (*S3OBSManager, error) {
	s3Config := &S3Config{
		BucketName: bucketName,
		Region:     region,
	}

	return NewS3OBSManager(s3Config, queueConfig)
}

func (m *S3OBSManager) Start() error {
	// 创建AWS配置
	var cfg aws.Config
	var err error

	if m.config.Endpoint != "" {
		// 使用自定义端点（MinIO等兼容S3的服务）
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               m.config.Endpoint,
				SigningRegion:     m.config.Region,
				HostnameImmutable: true,
			}, nil
		})

		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(m.config.Region),
			config.WithEndpointResolverWithOptions(customResolver),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				m.config.AccessKeyID,
				m.config.SecretAccessKey,
				"",
			)),
		)
	} else {
		// 使用AWS默认配置
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(m.config.Region),
		)
	}
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}

	// 创建S3客户端
	clientOptions := func(o *s3.Options) {
		o.UsePathStyle = m.config.ForcePathStyle
	}

	m.client = s3.NewFromConfig(cfg, clientOptions)
	m.uploader = manager.NewUploader(m.client)

	// 启动上传队列
	m.uploadQueue.Start(m)

	return nil
}

// uploadWithContext 带上下文的上传实现
func (m *S3OBSManager) uploadWithContext(ctx context.Context, localFilePath, s3Key string) (*VideoRecord, error) {
	// 打开文件
	// file, err := manager.ReadSeekCloser(localFilePath)
	file, err := os.Open(localFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	if m.config.Expires == 0 {
		m.config.Expires = 168 * time.Hour // 一周有168小时
	}

	var (
		ext         = strings.ToLower(path.Ext(localFilePath))
		contentType = aws.String(contentTypes[".bin"])
	)
	if v, ok := contentTypes[ext]; ok {
		contentType = aws.String(v)
	}

	// 执行上传
	result, err := m.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(m.config.BucketName),
		Key:         aws.String(s3Key),
		Body:        file,
		ContentType: contentType,
		// 7天后超时删除
		Expires: aws.Time(time.Now().Add(m.config.Expires)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload MP4 to S3: %v", err)
	}

	// 获取文件详细信息
	headOutput, err := m.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(m.config.BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object metadata: %v", err)
	}
	// if result.Location != "" {
	// 	fmt.Printf("Upload file[%s] result: UploadID=%s Location=%s", localFilePath, result.UploadID, result.Location)
	// }
	_ = result

	record := &VideoRecord{
		Key:          s3Key,
		Size:         *headOutput.ContentLength,
		ETag:         strings.Trim(*headOutput.ETag, "\""),
		LastModified: *headOutput.LastModified,
		StorageClass: string(headOutput.StorageClass),
	}

	// 生成预签名URL
	record.URL = m.GeneratePreSignedURL(ctx, s3Key, 24*time.Hour)

	return record, nil
}

// UploadMP4 上传MP4文件（同步接口）
func (m *S3OBSManager) UploadMP4(localFilePath, s3Key string, options ...UploadOption) (*UploadResult, error) {
	task := &UploadTask{
		LocalFilePath: localFilePath,
		S3Key:         s3Key,
	}

	for _, option := range options {
		option(task)
	}

	return m.uploadQueue.AddTaskSync(task, task.Timeout+30*time.Second)
}

// BatchUploadMP4 批量上传MP4文件
func (m *S3OBSManager) BatchUploadMP4(tasks []*UploadTask) ([]*UploadResult, error) {
	taskIDs, errors := m.uploadQueue.AddBatchTasks(tasks)
	if len(errors) > 0 {
		log.Printf("Some S3 tasks failed to add: %v", errors)
	}

	results := make([]*UploadResult, 0, len(taskIDs))
	for _, taskID := range taskIDs {
		result, err := m.uploadQueue.WaitForTaskResult(taskID, 30*time.Minute)
		if err != nil {
			results = append(results, &UploadResult{
				TaskID: taskID,
				Error:  err,
			})
		} else {
			results = append(results, result)
		}
	}

	return results, nil
}

// UploadMP4Async 上传MP4文件（异步接口）
func (m *S3OBSManager) UploadMP4Async(localFilePath, s3Key string, options ...UploadOption) error {
	task := &UploadTask{
		LocalFilePath: localFilePath,
		S3Key:         s3Key,
	}

	for _, option := range options {
		option(task)
	}

	return m.uploadQueue.AddTask(task)
}

// BatchUploadAsync 异步批量上传
func (m *S3OBSManager) BatchUploadAsync(tasks []*UploadTask, resultHandler func(*UploadResult)) error {
	m.uploadQueue.AddBatchTasksAsync(tasks, resultHandler)
	return nil
}

// UploadOption 上传选项
type UploadOption func(*UploadTask)

// WithTimeout 设置超时
func WithTimeout(timeout time.Duration) UploadOption {
	return func(task *UploadTask) {
		task.Timeout = timeout
	}
}

// WithRetries 设置重试次数
func WithRetries(retries int) UploadOption {
	return func(task *UploadTask) {
		task.MaxRetries = retries
	}
}

// WithPriority 设置优先级
func WithPriority(priority int) UploadOption {
	return func(task *UploadTask) {
		task.Priority = priority
	}
}

// WithContext 设置上下文
func WithContext(ctx context.Context) UploadOption {
	return func(task *UploadTask) {
		task.Context = ctx
	}
}

// GeneratePresignedURL 生成预签名URL
func (m *S3OBSManager) GeneratePreSignedURL(ctx context.Context, objectKey string, expiry time.Duration) string {
	preSignClient := s3.NewPresignClient(m.client)

	request, err := preSignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.config.BucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})

	if err != nil {
		log.Printf("Failed to generate presigned URL: %v", err)
		return ""
	}

	return request.URL
}

// QueryRecordsByDay 按天查询录像
func (m *S3OBSManager) QueryRecordsByDay(ctx context.Context, targetDate time.Time, prefix string) ([]VideoRecord, error) {
	// 构建日期前缀：prefix/YYYYMMDD/
	datePrefix := fmt.Sprintf("%s/%04d%02d%02d/",
		strings.TrimSuffix(prefix, "/"),
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day())

	return m.listObjectsWithPrefix(ctx, datePrefix)
}

// QueryRecordsByMonth 按月查询录像
func (m *S3OBSManager) QueryRecordsByMonth(ctx context.Context, year int, month time.Month, prefix string) ([]VideoRecord, error) {
	// TODO: 按月查询有待测试
	// 构建月份前缀：prefix/YYYYMM/
	monthPrefix := fmt.Sprintf("%s/%04d%02d/",
		strings.TrimSuffix(prefix, "/"),
		year, month)

	return m.listObjectsWithPrefix(ctx, monthPrefix)
}

// QueryRecordsByTimeRange 按开始时间结束时间查询录像
func (m *S3OBSManager) QueryRecordsByTimeRange(ctx context.Context, startTime, endTime time.Time, prefix string) ([]VideoRecord, error) {
	// 获取前缀下的所有对象
	allRecords, err := m.listObjectsWithPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// 过滤时间范围内的记录
	var filteredRecords []VideoRecord
	for _, record := range allRecords {
		if record.LastModified.After(startTime) && record.LastModified.Before(endTime) {
			filteredRecords = append(filteredRecords, record)
		}
	}

	return filteredRecords, nil
}

// 通用的前缀列表查询
func (m *S3OBSManager) listObjectsWithPrefix(ctx context.Context, prefix string) ([]VideoRecord, error) {
	var allRecords []VideoRecord
	var continuationToken *string = nil

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(m.config.BucketName),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		}

		result, err := m.client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %v", err)
		}

		// 转换响应结果为VideoRecord
		for _, object := range result.Contents {
			record := VideoRecord{
				Key:          *object.Key,
				LastModified: *object.LastModified,
				Size:         *object.Size,
				ETag:         strings.Trim(*object.ETag, "\""),
				StorageClass: string(object.StorageClass),
			}
			record.URL = m.GeneratePreSignedURL(ctx, record.Key, 24*time.Hour)
			allRecords = append(allRecords, record)
		}

		// 检查是否还有更多结果
		if !*result.IsTruncated {
			break
		}
		continuationToken = result.NextContinuationToken
	}

	return allRecords, nil
}

// QueryAllRecords 查询所有录像记录（用于调试）
func (m *S3OBSManager) QueryAllRecords(ctx context.Context, prefix string) ([]VideoRecord, error) {
	return m.listObjectsWithPrefix(ctx, prefix)
}

// DeleteRecordsByTimeRange 按指定时间范围删除录像
func (m *S3OBSManager) DeleteRecordsByTimeRange(ctx context.Context, startTime, endTime time.Time, prefix string) (int, error) {
	// 获取时间范围内的记录
	records, err := m.QueryRecordsByTimeRange(ctx, startTime, endTime, prefix)
	if err != nil {
		return 0, err
	}

	return m.batchDeleteRecords(ctx, records)
}

// DeleteRecordsByDay 按天删除录像
func (m *S3OBSManager) DeleteRecordsByDay(ctx context.Context, targetDate time.Time, prefix string) (int, error) {
	// 获取当天的记录
	records, err := m.QueryRecordsByDay(ctx, targetDate, prefix)
	if err != nil {
		return 0, err
	}

	return m.batchDeleteRecords(ctx, records)
}

// BatchDeleteRecords 录像批量删除
func (m *S3OBSManager) BatchDeleteRecords(ctx context.Context, objectKeys []string) (int, error) {
	if len(objectKeys) == 0 {
		return 0, nil
	}

	// S3批量删除最多支持1000个对象
	const maxBatchSize = 1000
	totalDeleted := 0

	for i := 0; i < len(objectKeys); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(objectKeys) {
			end = len(objectKeys)
		}

		batch := objectKeys[i:end]
		deleted, err := m.deleteBatch(ctx, batch)
		if err != nil {
			return totalDeleted, err
		}
		totalDeleted += deleted
	}

	return totalDeleted, nil
}

// 批量删除辅助函数
func (m *S3OBSManager) batchDeleteRecords(ctx context.Context, records []VideoRecord) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	objectKeys := make([]string, len(records))
	for i, record := range records {
		objectKeys[i] = record.Key
	}

	return m.BatchDeleteRecords(ctx, objectKeys)
}

// 执行批量删除
func (m *S3OBSManager) deleteBatch(ctx context.Context, objectKeys []string) (int, error) {
	objects := make([]types.ObjectIdentifier, len(objectKeys))
	for i, key := range objectKeys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(m.config.BucketName),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true), // 安静模式，不返回删除结果详情
		},
	}

	result, err := m.client.DeleteObjects(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to delete objects: %v", err)
	}

	// 计算成功删除的数量
	deletedCount := len(objectKeys)
	if len(result.Errors) > 0 {
		deletedCount = len(objectKeys) - len(result.Errors)
		// 记录删除错误
		for _, err := range result.Errors {
			log.Printf("Failed to delete %s: %s", *err.Key, *err.Message)
		}
	}

	return deletedCount, nil
}

// GetBucketUsage 获取存储桶使用情况
func (m *S3OBSManager) GetBucketUsage(ctx context.Context) (int64, error) {
	var totalSize int64
	var continuationToken *string = nil

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(m.config.BucketName),
			ContinuationToken: continuationToken,
		}

		result, err := m.client.ListObjectsV2(ctx, input)
		if err != nil {
			return 0, fmt.Errorf("failed to list objects: %v", err)
		}

		// 累加文件大小
		for _, object := range result.Contents {
			totalSize += *object.Size
		}

		// 检查是否还有更多结果
		if !*result.IsTruncated {
			break
		}
		continuationToken = result.NextContinuationToken
	}

	return totalSize, nil
}

// GetRecordCount 获取录像记录数量
func (m *S3OBSManager) GetRecordCount(ctx context.Context, prefix string) (int, error) {
	var count int
	var continuationToken *string = nil

	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            aws.String(m.config.BucketName),
			Prefix:            aws.String(prefix),
			ContinuationToken: continuationToken,
		}

		result, err := m.client.ListObjectsV2(ctx, input)
		if err != nil {
			return 0, fmt.Errorf("failed to list objects: %v", err)
		}

		count += len(result.Contents)

		if !*result.IsTruncated {
			break
		}
		continuationToken = result.NextContinuationToken
	}

	return count, nil
}

// GetQueueStats 获取队列统计
func (m *S3OBSManager) GetQueueStats() QueueStats {
	if m.uploadQueue != nil {
		return m.uploadQueue.GetQueueStats()
	}
	return QueueStats{}
}

func (m *S3OBSManager) GetSMSNode() string {
	return m.config.SmsNode
}

// CreateFolder 创建文件夹（S3中实际上是创建空对象）
func (m *S3OBSManager) CreateFolder(ctx context.Context, folderPath string) error {
	// 确保路径以/结尾
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	_, err := m.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(m.config.BucketName),
		Key:    aws.String(folderPath),
		Body:   strings.NewReader(""),
	})

	return err
}

// createBucket 创建一个指定名称和区域的 S3 存储桶 (一般用在桶不存在的时候去创建)
func (m *S3OBSManager) CreateBucket(bucketName string, region string) error {
	// 1. 加载默认配置，包括区域和认证凭证
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region), // 显式指定区域
	)

	if err != nil {
		return fmt.Errorf("无法加载 AWS 配置: %v", err)
	}

	// 2. 创建 S3 服务客户端
	client := s3.NewFromConfig(cfg)

	// 3. 准备创建存储桶的请求
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		// 注意：在除 us-east-1 以外的区域创建桶时，必须指定 CreateBucketConfiguration。
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}

	// 4. 发送请求
	_, err = client.CreateBucket(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("创建存储桶失败: %v", err)
	}

	// 5. 等待存储桶存在（可选但推荐）
	// 创建操作是异步的，此操作会阻塞直到桶真正创建成功并可访问。
	waiter := s3.NewBucketExistsWaiter(client)
	waitErr := waiter.Wait(context.TODO(),
		&s3.HeadBucketInput{Bucket: aws.String(bucketName)},
		2*time.Minute, // 最长等待时间
	)
	if waitErr != nil {
		return fmt.Errorf("等待存储桶就绪时发生错误: %v", waitErr)
	}

	fmt.Printf("存储桶 '%s' 在区域 '%s' 中创建成功！\n", bucketName, region)
	return nil
}

// Close 关闭管理器资源（S3客户端不需要显式关闭）
func (m *S3OBSManager) Close() {
	if m.uploadQueue != nil {
		m.uploadQueue.Stop()
	}
	// AWS SDK v2 不需要显式关闭客户端
}
