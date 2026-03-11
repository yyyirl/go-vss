// @Title        s3obs
// @Description  s3_upload_queue
// @Create       dingshuai 2025/9/26 09:46

package s3obs

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// UploadTask 上传任务
type UploadTask struct {
	ID            string
	LocalFilePath string
	S3Key         string
	RetryCount    int
	MaxRetries    int
	Timeout       time.Duration
	Priority      int // 优先级，数值越小优先级越高
	CreatedAt     time.Time
	Context       context.Context
	CancelFunc    context.CancelFunc
}

// UploadResult 上传结果
type UploadResult struct {
	TaskID    string
	Task      *UploadTask
	Record    *VideoRecord
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// UploadQueueConfig 上传队列配置
type UploadQueueConfig struct {
	MaxConcurrentUploads int           // 最大并发上传数
	MaxQueueSize         int           // 最大队列大小
	DefaultTimeout       time.Duration // 默认超时时间
	DefaultRetries       int           // 默认重试次数
	RetryDelay           time.Duration // 重试延迟
	PriorityEnabled      bool          // 是否启用优先级
	ResultBufferSize     int           // 结果缓冲区大小
	DeleteLocalEnable    bool          // 是否在上传完后删除本地录像（不管成功与否）
}

// QueueStats 队列统计
type QueueStats struct {
	TotalTasks    int64
	SuccessTasks  int64
	FailedTasks   int64
	CurrentQueue  int
	ActiveWorkers int
	WaitingTasks  int
	mu            sync.RWMutex
}

// UploadQueue 上传队列
type UploadQueue struct {
	config       *UploadQueueConfig
	taskChan     chan *UploadTask   // 普通任务通道
	priorityChan chan *UploadTask   // 高优先级任务通道
	resultChan   chan *UploadResult // 结果通道
	stopChan     chan struct{}      // 停止信号
	working      chan *UploadTask   // 工作协程工作信号
	wg           sync.WaitGroup
	isRunning    bool
	mu           sync.RWMutex
	stats        *QueueStats
	taskMap      sync.Map // 任务映射，用于任务管理
}

// NewUploadQueue 创建上传队列
func NewUploadQueue(config *UploadQueueConfig) *UploadQueue {
	if config.MaxConcurrentUploads <= 0 {
		config.MaxConcurrentUploads = 5
	}
	if config.MaxQueueSize <= 0 {
		config.MaxQueueSize = 1000
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 30 * time.Minute
	}
	if config.DefaultRetries <= 0 {
		config.DefaultRetries = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.ResultBufferSize <= 0 {
		config.ResultBufferSize = 100
	}

	return &UploadQueue{
		config:       config,
		taskChan:     make(chan *UploadTask, config.MaxQueueSize),
		priorityChan: make(chan *UploadTask, config.MaxQueueSize/2),
		resultChan:   make(chan *UploadResult, config.ResultBufferSize),
		stopChan:     make(chan struct{}),
		working:      make(chan *UploadTask, config.MaxConcurrentUploads*10), // 缓冲工作信号
		stats:        &QueueStats{},
	}
}

// Start 启动上传队列
func (q *UploadQueue) Start(manager *S3OBSManager) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.isRunning {
		return
	}

	q.isRunning = true

	// 启动任务分发器
	q.wg.Add(1)
	go q.taskDispatcher(manager)

	// 启动工作协程池
	for i := 0; i < q.config.MaxConcurrentUploads; i++ {
		q.wg.Add(1)
		go q.worker(i, manager)
	}

	// 启动统计监控协程
	// q.wg.Add(1)
	// go q.monitor()

	// log.Printf("S3 Upload queue started with %d workers", q.config.MaxConcurrentUploads)
}

// taskDispatcher 任务分发器
func (q *UploadQueue) taskDispatcher(manager *S3OBSManager) {
	defer q.wg.Done()

	taskBuffer := make([]*UploadTask, 0, 100)
	priorityBuffer := make([]*UploadTask, 0, 50)

	flushTimer := time.NewTicker(100 * time.Millisecond)
	defer flushTimer.Stop()

	for {
		select {
		case task := <-q.taskChan:
			taskBuffer = append(taskBuffer, task)
			q.updateWaitingTasks(1)

		case task := <-q.priorityChan:
			priorityBuffer = append(priorityBuffer, task)
			q.updateWaitingTasks(1)

		case <-flushTimer.C:
			q.flushTasks(taskBuffer, priorityBuffer)
			taskBuffer = taskBuffer[:0]
			priorityBuffer = priorityBuffer[:0]

		case <-q.stopChan:
			q.flushTasks(taskBuffer, priorityBuffer)
			close(q.working)
			// log.Println("S3 Task dispatcher stopped")
			return
		}
	}
}

// flushTasks 刷新任务到工作协程
func (q *UploadQueue) flushTasks(taskBuffer, priorityBuffer []*UploadTask) {
	if len(priorityBuffer) == 0 && len(taskBuffer) == 0 {
		return
	}

	// 合并并排序任务
	allTasks := append(priorityBuffer, taskBuffer...)
	if q.config.PriorityEnabled {
		sort.Slice(allTasks, func(i, j int) bool {
			return allTasks[i].Priority < allTasks[j].Priority
		})
	}

	// 分发任务到工作协程
	for _, task := range allTasks {
		select {
		case q.working <- task:
			q.taskMap.Store(task.ID, task)
			q.updateWaitingTasks(-1)
		default:
			// 工作协程繁忙，将任务重新放回缓冲区
			if task.Priority <= 2 {
				select {
				case q.priorityChan <- task:
				default:
					log.Printf("S3 Priority channel full, task %s may be delayed", task.ID)
				}
			} else {
				select {
				case q.taskChan <- task:
				default:
					log.Printf("S3 Task channel full, task %s may be delayed", task.ID)
				}
			}
		}
	}
}

// worker 工作协程
func (q *UploadQueue) worker(id int, manager *S3OBSManager) {
	defer q.wg.Done()

	// log.Printf("S3 Upload worker %d started", id)

	for {
		select {
		case task := <-q.working:
			q.processNextTask(id, task, manager)
		case <-q.stopChan:
			// log.Printf("S3 Upload worker %d stopped", id)
			return
		}
	}
}

// 处理下一个任务
func (q *UploadQueue) processNextTask(workerID int, task *UploadTask, manager *S3OBSManager) {
	if task == nil {
		return
	}

	q.stats.mu.Lock()
	q.stats.ActiveWorkers++
	q.stats.mu.Unlock()

	result := q.processTask(task, manager)
	// 只要有上传结果了就处理
	if q.config.DeleteLocalEnable {
		// 删除本地文件
		os.Remove(task.LocalFilePath)
	}

	q.stats.mu.Lock()
	q.stats.ActiveWorkers--
	q.stats.mu.Unlock()

	select {
	case q.resultChan <- result:
	default:
		log.Printf("S3 Result channel full, result for task %s dropped", task.ID)
	}

	q.taskMap.Delete(task.ID)
}

// processTask 处理单个上传任务
func (q *UploadQueue) processTask(task *UploadTask, manager *S3OBSManager) *UploadResult {
	startTime := time.Now()
	result := &UploadResult{
		TaskID:    task.ID,
		Task:      task,
		StartTime: startTime,
	}

	defer func() {
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		if r := recover(); r != nil {
			result.Error = fmt.Errorf("upload panic: %v", r)
		}
	}()

	// 设置任务上下文
	if task.Context == nil {
		task.Context, task.CancelFunc = context.WithTimeout(context.Background(), task.Timeout)
	} else {
		// 确保上下文有超时设置
		if _, hasDeadline := task.Context.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			task.Context, cancel = context.WithTimeout(task.Context, task.Timeout)
			if task.CancelFunc != nil {
				task.CancelFunc()
			}
			task.CancelFunc = cancel
		}
	}

	// 执行上传（带重试机制）
	var record *VideoRecord
	var err error

	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("S3 Retry %d for task %s", attempt, task.ID)
			time.Sleep(q.config.RetryDelay * time.Duration(attempt))
		}

		record, err = manager.uploadWithContext(task.Context, task.LocalFilePath, task.S3Key)
		if err == nil {
			break
		}

		// 检查是否是超时错误
		if task.Context.Err() == context.DeadlineExceeded {
			err = fmt.Errorf("upload timeout after %v: %v", task.Timeout, err)
			break
		}

		// 检查是否应该重试
		if attempt == task.MaxRetries {
			err = fmt.Errorf("upload failed after %d attempts: %v", task.MaxRetries+1, err)
			break
		}

		// 检查错误是否可重试
		if !isRetryableError(err) {
			break
		}
	}

	result.Record = record
	result.Error = err

	return result
}

// isRetryableError 检查错误是否可重试
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 网络错误、超时错误等可以重试
	errorStr := err.Error()
	retryableErrors := []string{
		"timeout", "deadline exceeded", "network", "connection reset",
		"temporary", "throttling", "slow down", "service unavailable",
	}

	for _, retryableError := range retryableErrors {
		if strings.Contains(strings.ToLower(errorStr), retryableError) {
			return true
		}
	}

	return false
}

// AddTask 添加上传任务到队列
func (q *UploadQueue) AddTask(task *UploadTask) error {
	q.setTaskDefaults(task)

	if task.ID == "" {
		task.ID = generateTaskID()
	}

	if task.Priority <= 2 {
		select {
		case q.priorityChan <- task:
			q.updateStatsOnAdd()
			return nil
		default:
			return fmt.Errorf("S3 priority queue is full, max size: %d", cap(q.priorityChan))
		}
	} else {
		select {
		case q.taskChan <- task:
			q.updateStatsOnAdd()
			return nil
		default:
			return fmt.Errorf("S3 task queue is full, max size: %d", cap(q.taskChan))
		}
	}
}

// AddTaskSync 同步添加任务并等待结果
func (q *UploadQueue) AddTaskSync(task *UploadTask, timeout time.Duration) (*UploadResult, error) {
	if err := q.AddTask(task); err != nil {
		return nil, err
	}

	return q.WaitForTaskResult(task.ID, timeout)
}

// AddBatchTasks 批量添加任务
func (q *UploadQueue) AddBatchTasks(tasks []*UploadTask) ([]string, []error) {
	var taskIDs []string
	var errors []error

	for _, task := range tasks {
		if err := q.AddTask(task); err != nil {
			errors = append(errors, fmt.Errorf("task %s: %v", task.LocalFilePath, err))
		} else {
			taskIDs = append(taskIDs, task.ID)
		}
	}

	return taskIDs, errors
}

// AddBatchTasksAsync 异步批量添加任务
func (q *UploadQueue) AddBatchTasksAsync(tasks []*UploadTask, resultHandler func(*UploadResult)) {
	go func() {
		for _, task := range tasks {
			if err := q.AddTask(task); err != nil {
				continue
			}

			if resultHandler != nil {
				go func(t *UploadTask) {
					result, err := q.WaitForTaskResult(t.ID, t.Timeout+30*time.Second)
					if err != nil {
						resultHandler(&UploadResult{
							TaskID: t.ID,
							Task:   t,
							Error:  err,
						})
					} else {
						resultHandler(result)
					}
				}(task)
			}
		}
	}()
}

// WaitForTaskResult 等待特定任务的结果
func (q *UploadQueue) WaitForTaskResult(taskID string, timeout time.Duration) (*UploadResult, error) {
	timeoutChan := time.After(timeout)

	for {
		select {
		case result := <-q.resultChan:
			if result.TaskID == taskID {
				q.updateStatsOnResult(result)
				return result, nil
			}
			// 不是目标任务，放回结果通道
			go func() {
				select {
				case q.resultChan <- result:
				case <-time.After(100 * time.Millisecond):
					log.Printf("S3 Failed to return result to channel for task %s", result.TaskID)
				}
			}()

		case <-timeoutChan:
			return nil, fmt.Errorf("timeout waiting for S3 task %s result", taskID)

		case <-q.stopChan:
			return nil, fmt.Errorf("S3 queue stopped while waiting for task %s result", taskID)
		}
	}
}

// StreamResults 流式获取结果
func (q *UploadQueue) StreamResults(ctx context.Context) <-chan *UploadResult {
	resultStream := make(chan *UploadResult, 10)

	go func() {
		defer close(resultStream)

		for {
			select {
			case result := <-q.resultChan:
				select {
				case resultStream <- result:
					q.updateStatsOnResult(result)
				case <-ctx.Done():
					return
				}

			case <-ctx.Done():
				return

			case <-q.stopChan:
				return
			}
		}
	}()

	return resultStream
}

// GetQueueStats 获取队列统计
func (q *UploadQueue) GetQueueStats() QueueStats {
	q.stats.mu.RLock()
	defer q.stats.mu.RUnlock()

	stats := *q.stats
	stats.CurrentQueue = len(q.taskChan) + len(q.priorityChan)
	return stats
}

// setTaskDefaults 设置任务默认值
func (q *UploadQueue) setTaskDefaults(task *UploadTask) {
	if task.Timeout <= 0 {
		task.Timeout = q.config.DefaultTimeout
	}
	if task.MaxRetries <= 0 {
		task.MaxRetries = q.config.DefaultRetries
	}
	if task.Context == nil {
		task.Context, task.CancelFunc = context.WithTimeout(context.Background(), task.Timeout)
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("s3_task_%d", time.Now().UnixNano())
}

// updateStatsOnAdd 添加任务时更新统计
func (q *UploadQueue) updateStatsOnAdd() {
	q.stats.mu.Lock()
	defer q.stats.mu.Unlock()
	q.stats.TotalTasks++
	q.stats.WaitingTasks++
}

// updateStatsOnResult 处理结果时更新统计
func (q *UploadQueue) updateStatsOnResult(result *UploadResult) {
	q.stats.mu.Lock()
	defer q.stats.mu.Unlock()

	q.stats.WaitingTasks--
	if result.Error != nil {
		q.stats.FailedTasks++
	} else {
		q.stats.SuccessTasks++
	}
}

// updateWaitingTasks 更新等待任务数量
func (q *UploadQueue) updateWaitingTasks(delta int) {
	q.stats.mu.Lock()
	defer q.stats.mu.Unlock()
	q.stats.WaitingTasks += delta
}

// monitor 监控协程
func (q *UploadQueue) monitor() {
	defer q.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 创建协程搜集结果
	go func() {
		for {
			select {
			case result := <-q.resultChan:
				q.updateStatsOnResult(result)
			case <-q.stopChan:
				return
			}
		}
	}()

	// 监控对象存储文件上传情况
	for {
		select {
		case <-ticker.C:
			stats := q.GetQueueStats()
			log.Printf("S3 Queue Stats - Total: %d, Success: %d, Failed: %d, Queue: %d, Workings: %d",
				stats.TotalTasks, stats.SuccessTasks, stats.FailedTasks,
				stats.WaitingTasks, stats.ActiveWorkers)

		case <-q.stopChan:
			return
		}
	}
}

// Stop 停止上传队列
func (q *UploadQueue) Stop() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.isRunning {
		return
	}

	close(q.stopChan)
	q.wg.Wait()
	q.isRunning = false
	// log.Println("S3 Upload queue stopped")
}
