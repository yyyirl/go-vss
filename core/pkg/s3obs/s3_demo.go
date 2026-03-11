// @Title        s3obs
// @Description  s3_demo
// @Create       dingshuai 2025/9/25 13:39

package s3obs

import (
	"context"
	"fmt"
	"log"
	"time"
)

func Demo() {

	// 配置上传队列
	queueConfig := &UploadQueueConfig{
		MaxConcurrentUploads: 6,
		MaxQueueSize:         1000,
		DefaultTimeout:       1 * time.Minute,
		DefaultRetries:       3,
		RetryDelay:           3 * time.Second,
		PriorityEnabled:      true,
		ResultBufferSize:     100,
	}

	// 配置S3连接（以MinIO为例）
	// 华为云对象存储测试
	cfg := &S3Config{
		Endpoint:        "",
		Region:          "",
		AccessKeyID:     "",
		SecretAccessKey: "",
		BucketName:      "",
		ForcePathStyle:  true, // MinIO需要这个设置
	}

	// 创建S3管理器
	manager, err := NewS3OBSManager(cfg, queueConfig)
	if err != nil {
		log.Fatal("Failed to create S3 manager:", err)
	}
	defer manager.Close()

	ctx := context.Background()

	// 1. 单文件同步/异步上传
	exampleSingleUpload(manager)

	// 2. 批量同步上传
	exampleBatchUploadSync(manager)

	// 3. 批量异步上传
	exampleBatchUploadAsync(manager)

	// 4. 流式上传处理
	exampleStreamingUpload(manager)

	// 5. 查询功能演示
	exampleQueryOperations(ctx, manager)

	// 6. 删除功能演示
	exampleDeleteOperations(ctx, manager)

	// 7. 监控和统计
	exampleMonitoring(manager)

	// 6. 获取存储桶使用情况
	usage, err := manager.GetBucketUsage(ctx)
	if err != nil {
		log.Fatal("Get bucket usage failed:", err)
	}
	fmt.Printf("Bucket usage: %.2f MB\n", float64(usage)/(1024*1024))
}

// 示例1: 单文件同步上传
func exampleSingleUpload(manager *S3OBSManager) {
	fmt.Println("\n--- 1. 单文件同步上传 ---")

	// 基本上传
	result, err := manager.UploadMP4(
		"video1.mp4",
		"videos/2024/01/15/camera1.mp4",
	)
	if err != nil {
		log.Printf("基本上传失败: %v", err)
	} else if result.Error != nil {
		log.Printf("上传错误: %v", result.Error)
	} else {
		fmt.Printf("✓ 上传成功: %s (大小: %d bytes, 耗时: %v)\n",
			result.Record.Key, result.Record.Size, result.Duration)
	}

	// 带选项的上传
	result, err = manager.UploadMP4(
		"important-video.mp4",
		"videos/2024/01/15/important-camera1.mp4",
		WithTimeout(5*time.Minute),
		WithRetries(5),
		WithPriority(1), // 高优先级
	)
	if err != nil {
		log.Printf("带选项上传失败: %v", err)
	} else if result.Error != nil {
		log.Printf("带选项上传错误: %v", result.Error)
	} else {
		fmt.Printf("✓ 高优先级上传成功: %s\n", result.Record.Key)
	}

	// 带上下文的上传
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// 同步上传
	result, err = manager.UploadMP4(
		"urgent-video.mp4",
		"videos/2024/01/15/urgent-camera1.mp4",
		WithContext(ctx),
		WithPriority(0), // 最高优先级
	)
	if err != nil {
		log.Printf("上下文上传失败: %v", err)
	} else if result.Error != nil {
		log.Printf("上下文上传错误: %v", result.Error)
	} else {
		fmt.Printf("✓ 上下文上传成功: %s\n", result.Record.Key)
	}

	// 异步上传
	// result, err = manager.UploadMP4Async(
	// 	"urgent-video.mp4",
	// 	"videos/2024/01/15/urgent-camera1.mp4",
	// 	WithContext(ctx),
	// 	WithPriority(0), // 最高优先级
	// )
}

// 示例2: 批量同步上传
func exampleBatchUploadSync(manager *S3OBSManager) {
	fmt.Println("\n--- 2. 批量同步上传 ---")

	tasks := []*UploadTask{
		{
			LocalFilePath: "video2.mp4",
			S3Key:         "videos/2024/01/15/camera2.mp4",
			Priority:      2,
			Timeout:       10 * time.Minute,
		},
		{
			LocalFilePath: "video3.mp4",
			S3Key:         "videos/2024/01/15/camera3.mp4",
			Priority:      1, // 更高优先级
			MaxRetries:    5,
		},
		{
			LocalFilePath: "video4.mp4",
			S3Key:         "videos/2024/01/15/camera4.mp4",
			Priority:      3,
		},
	}

	startTime := time.Now()
	results, err := manager.BatchUploadMP4(tasks)
	if err != nil {
		log.Printf("批量上传失败: %v", err)
		return
	}

	successCount := 0
	for i, result := range results {
		if result.Error != nil {
			fmt.Printf("✗ 任务 %d 失败: %v\n", i+1, result.Error)
		} else {
			fmt.Printf("✓ 任务 %d 成功: %s\n", i+1, result.Record.Key)
			successCount++
		}
	}

	fmt.Printf("批量上传完成: %d/%d 成功, 总耗时: %v\n",
		successCount, len(tasks), time.Since(startTime))
}

// 示例3: 批量异步上传
func exampleBatchUploadAsync(manager *S3OBSManager) {
	fmt.Println("\n--- 3. 批量异步上传 ---")

	tasks := []*UploadTask{
		{
			LocalFilePath: "video5.mp4",
			S3Key:         "videos/2024/01/15/camera5.mp4",
			Priority:      2,
		},
		{
			LocalFilePath: "video6.mp4",
			S3Key:         "videos/2024/01/15/camera6.mp4",
			Priority:      1,
		},
	}

	// 创建结果通道
	resultChan := make(chan *UploadResult, len(tasks))
	completed := make(chan bool)

	// 启动异步上传
	err := manager.BatchUploadAsync(tasks, func(result *UploadResult) {
		resultChan <- result

		if result.Error != nil {
			fmt.Printf("异步上传失败: %s - %v\n", result.TaskID, result.Error)
		} else {
			fmt.Printf("异步上传成功: %s -> %s\n", result.TaskID, result.Record.Key)
		}

		// 检查是否所有任务完成
		if len(resultChan) == len(tasks) {
			completed <- true
		}
	})

	if err != nil {
		log.Printf("异步上传设置失败: %v", err)
		return
	}

	fmt.Println("异步上传已启动，等待完成...")

	// 等待完成或超时
	select {
	case <-completed:
		fmt.Println("✓ 所有异步上传任务完成")
	case <-time.After(30 * time.Minute):
		fmt.Println("⚠ 异步上传超时")
	case <-time.After(5 * time.Second):
		fmt.Println("异步上传进行中...（演示超时）")
	}

	close(resultChan)
	close(completed)
}

// 示例4: 流式上传处理
func exampleStreamingUpload(manager *S3OBSManager) {
	fmt.Println("\n--- 4. 流式上传处理 ---")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 启动流式结果监听
	resultStream := manager.uploadQueue.StreamResults(ctx)

	// 模拟实时添加任务
	go func() {
		for i := 0; i < 5; i++ {
			task := &UploadTask{
				LocalFilePath: fmt.Sprintf("live-video-%d.mp4", i+1),
				S3Key:         fmt.Sprintf("videos/live/2024/01/15/camera-live-%d.mp4", i+1),
				Priority:      0, // 实时任务最高优先级
				Timeout:       2 * time.Minute,
			}

			if err := manager.uploadQueue.AddTask(task); err != nil {
				log.Printf("添加实时任务失败: %v", err)
			} else {
				fmt.Printf("添加实时任务: %s\n", task.S3Key)
			}

			time.Sleep(2 * time.Second)
		}
	}()

	// 处理流式结果
	resultCount := 0
	for {
		select {
		case result, ok := <-resultStream:
			if !ok {
				fmt.Println("流式处理结束")
				return
			}

			resultCount++
			if result.Error != nil {
				fmt.Printf("流式任务失败: %s - %v\n", result.TaskID, result.Error)
			} else {
				fmt.Printf("流式任务成功: %s (耗时: %v)\n", result.TaskID, result.Duration)
			}

			if resultCount >= 5 {
				fmt.Println("✓ 流式处理完成")
				return
			}

		case <-ctx.Done():
			fmt.Println("流式处理超时")
			return
		}
	}
}

// 示例5: 查询功能演示
func exampleQueryOperations(ctx context.Context, manager *S3OBSManager) {
	fmt.Println("\n--- 5. 查询功能演示 ---")

	// 按天查询
	today := time.Now()
	dayRecords, err := manager.QueryRecordsByDay(ctx, today, "videos")
	if err != nil {
		log.Printf("按天查询失败: %v", err)
	} else {
		fmt.Printf("今天(%s)的录像数量: %d\n", today.Format("2006-01-02"), len(dayRecords))
		for i, record := range dayRecords {
			if i < 3 { // 只显示前3条
				fmt.Printf("  %s - %s (%d bytes)\n",
					record.Key, record.LastModified.Format("15:04:05"), record.Size)
			}
		}
		if len(dayRecords) > 3 {
			fmt.Printf("  ... 还有 %d 条记录\n", len(dayRecords)-3)
		}
	}

	// 按月查询
	monthRecords, err := manager.QueryRecordsByMonth(ctx, 2024, time.January, "videos")
	if err != nil {
		log.Printf("按月查询失败: %v", err)
	} else {
		fmt.Printf("2024年1月的录像数量: %d\n", len(monthRecords))
	}

	// 按时间范围查询
	startTime := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC)
	timeRangeRecords, err := manager.QueryRecordsByTimeRange(ctx, startTime, endTime, "videos")
	if err != nil {
		log.Printf("时间范围查询失败: %v", err)
	} else {
		fmt.Printf("2024-01-15 全天的录像数量: %d\n", len(timeRangeRecords))
	}
}

// 示例6: 删除功能演示
func exampleDeleteOperations(ctx context.Context, manager *S3OBSManager) {
	fmt.Println("\n--- 6. 删除功能演示 ---")

	// 注意：在实际使用中要小心删除操作
	// 这里只是演示用法，不实际执行删除

	today := time.Now()
	testDate := today.AddDate(0, 0, -1) // 昨天的日期

	fmt.Printf("演示删除 %s 的录像...\n", testDate.Format("2006-01-02"))

	// 先查询要删除的记录
	records, err := manager.QueryRecordsByDay(ctx, testDate, "videos")
	if err != nil {
		log.Printf("查询删除记录失败: %v", err)
		return
	}

	if len(records) == 0 {
		fmt.Println("没有找到要删除的记录")
		return
	}

	fmt.Printf("找到 %d 条可删除记录\n", len(records))

	// 演示批量删除（注释掉实际删除代码）
	/*
		deletedCount, err := manager.DeleteRecordsByDay(ctx, testDate, "videos")
		if err != nil {
			log.Printf("按天删除失败: %v", err)
		} else {
			fmt.Printf("成功删除 %d 条记录\n", deletedCount)
		}
	*/

	fmt.Println("删除功能演示完成（实际删除代码已注释）")
}

// 示例7: 监控和统计
func exampleMonitoring(manager *S3OBSManager) {
	fmt.Println("\n--- 7. 监控和统计 ---")

	// 获取实时统计
	stats := manager.GetQueueStats()
	fmt.Printf("实时队列统计:\n")
	fmt.Printf("  总任务数: %d\n", stats.TotalTasks)
	fmt.Printf("  成功任务: %d\n", stats.SuccessTasks)
	fmt.Printf("  失败任务: %d\n", stats.FailedTasks)
	fmt.Printf("  等待任务: %d\n", stats.WaitingTasks)
	fmt.Printf("  活跃工作线程: %d\n", stats.ActiveWorkers)

	// 模拟监控循环
	fmt.Println("监控数据采样（5秒）...")
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second)
		stats := manager.GetQueueStats()
		fmt.Printf("  采样 %d: 活跃=%d, 等待=%d, 成功=%d, 失败=%d\n",
			i+1, stats.ActiveWorkers, stats.WaitingTasks,
			stats.SuccessTasks, stats.FailedTasks)
	}
}
