// @Title        文件下载
// @Description  带下载进度
// @Create       yiyiyi 2025/11/8 17:05

package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/xmap"
)

type status string

const (
	StatusDownloading status = "downloading"
	StatusCompleted   status = "completed"
	StatusError       status = "error"
	StatusCancelled   status = "cancelled"
)

type DownloadTask struct {
	TaskID     string
	URL        string
	Filepath   string
	Progress   float64
	Downloaded int64
	Total      int64
	Speed      float64
	Status     status
	Message    string
	StartTime  time.Time
}

type DownloadResponse struct {
	TaskID string `json:"task_id"` // 下载任务ID
	Status string `json:"status"`  // 任务状态
}

type ProgressResponse struct {
	TaskID     string  `json:"task_id"`
	Progress   float64 `json:"progress"`   // 进度百分比
	Downloaded int64   `json:"downloaded"` // 已下载字节数
	Total      int64   `json:"total"`      // 总字节数
	Speed      float64 `json:"speed"`      // 下载速度 KB/s
	Status     status  `json:"status"`     // 状态
	Message    string  `json:"message"`    // 状态信息
}

type DownloadManager struct {
	tasks   *xmap.XMap[string, *DownloadTask]
	clients *xmap.XMap[string, chan ProgressUpdate]
}

type ProgressUpdate struct {
	Downloaded int64   `json:"downloaded"`
	Total      int64   `json:"total"`
	Progress   float64 `json:"progress"`
	Speed      float64 `json:"speed"`
	TaskID     string  `json:"taskID"`
	Status     status  `json:"status"`
	Message    string  `json:"message"`
	Filepath   string  `json:"filepath"`
}

var (
	manager *DownloadManager
	once    sync.Once
)

func GetManager() *DownloadManager {
	once.Do(func() {
		manager = &DownloadManager{
			tasks:   xmap.New[string, *DownloadTask](100),
			clients: xmap.New[string, chan ProgressUpdate](100),
		}
	})
	return manager
}

func (dm *DownloadManager) CreateTask(url, fileName, saveDir string) *DownloadTask {
	_ = functions.MakeDir(saveDir)

	var taskID = url
	if fileName == "" {
		fileName = filepath.Base(url)
	}

	var (
		filePath = filepath.Join(saveDir, fileName)
		task     = &DownloadTask{
			TaskID:    taskID,
			URL:       url,
			Filepath:  filePath,
			Status:    StatusDownloading,
			StartTime: time.Now(),
		}
	)
	dm.tasks.Set(taskID, task)

	return task
}

func (dm *DownloadManager) CancelDownload(taskID string) {
	task, exists := dm.tasks.Get(taskID)
	if !exists {
		return
	}

	task.updateStatus(StatusCancelled, "cancelled")
	dm.notifyClients(task)
}

func (dm *DownloadManager) Finished(taskID string) {
	if ch, exists := dm.clients.Get(taskID); exists {
		close(ch)
		dm.clients.Remove(taskID)
	}

	dm.tasks.Remove(taskID)
}

func (dm *DownloadManager) StartDownload(_ context.Context, task *DownloadTask) {
	defer func() {
		if r := recover(); r != nil {
			task.Status = StatusError
			task.Message = fmt.Sprintf("abnormal: %v", r)
			dm.notifyClients(task)
		}
	}()

	_ = os.Remove(task.Filepath)

	cancelableReq, err := http.NewRequest("GET", task.URL, nil)
	if err != nil {
		task.updateStatus(StatusError, "http client create failed: "+err.Error())
		dm.notifyClients(task)
		return
	}

	var client = &http.Client{}
	resp, err := client.Do(cancelableReq)
	if err != nil {
		task.updateStatus(StatusError, "request failed: "+err.Error())
		dm.notifyClients(task)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		task.updateStatus(StatusError, fmt.Sprintf("HTTP error: %d", resp.StatusCode))
		dm.notifyClients(task)
		return
	}

	task.Total = resp.ContentLength
	file, err := os.Create(task.Filepath)
	if err != nil {
		task.updateStatus(StatusError, "create file failed err: "+err.Error())
		dm.notifyClients(task)
		return
	}
	defer file.Close()

	var (
		buffer    = make([]byte, 32*1024) // 32KB buffer
		totalRead = int64(0)
	)
	for {
		n, err := resp.Body.Read(buffer)
		if task.Status == StatusCancelled {
			task.updateStatus(StatusCancelled, "cancelled")
			dm.notifyClients(task)
			break
		}

		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				task.updateStatus(StatusError, "write file failed, err: "+writeErr.Error())
				dm.notifyClients(task)
				break
			}

			totalRead += int64(n)
			task.updateProgress(totalRead)
			dm.notifyClients(task)
		}

		if err != nil {
			if err == io.EOF {
				task.updateStatus(StatusCompleted, "下载完成")
				dm.notifyClients(task)
				break
			}
			task.updateStatus(StatusError, "read file failed, err: "+err.Error())
			dm.notifyClients(task)
			break
		}
	}

	if task.Status == StatusCancelled {
		functions.LogInfo("下载已取消")
		_ = os.Remove(task.Filepath)
	} else {
		functions.LogInfo("下载已结束")
	}
}

func (dm *DownloadManager) GetTask(taskID string) *DownloadTask {
	v, _ := dm.tasks.Get(taskID)
	return v
}

func (dm *DownloadManager) TaskNum() int {
	return dm.tasks.Len()
}

func (dm *DownloadManager) ClientNum() int {
	return dm.clients.Len()
}

func (dm *DownloadManager) Subscribe(taskID string) chan ProgressUpdate {
	var ch = make(chan ProgressUpdate, 10)
	dm.clients.Set(taskID, ch)
	return ch
}

func (dm *DownloadManager) Unsubscribe(taskID string) {
	if ch, exists := dm.clients.Get(taskID); exists {
		close(ch)
		dm.clients.Remove(taskID)
	}
}

func (dm *DownloadManager) CheckExists(taskID string) bool {
	_, exists := dm.tasks.Get(taskID)
	return exists
}

func (dm *DownloadManager) notifyClients(task *DownloadTask) {
	ch, exists := dm.clients.Get(task.TaskID)
	if exists {
		ch <- ProgressUpdate{
			TaskID:     task.TaskID,
			Progress:   task.Progress,
			Downloaded: task.Downloaded,
			Total:      task.Total,
			Speed:      task.Speed,
			Status:     task.Status,
			Message:    task.Message,
			Filepath:   task.Filepath,
		}
	}
}

type ProgressReader struct {
	Reader     io.Reader
	OnProgress func(readBytes int64)
	read       int64
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 {
		pr.read += int64(n)
		if pr.OnProgress != nil {
			pr.OnProgress(pr.read)
		}
	}
	return n, err
}

// DownloadTask 的方法
func (dt *DownloadTask) updateProgress(downloaded int64) {
	dt.Downloaded = downloaded
	if dt.Total > 0 {
		dt.Progress = float64(downloaded) / float64(dt.Total) * 100
	}

	// 计算下载速度
	elapsed := time.Since(dt.StartTime).Seconds()
	if elapsed > 0 {
		dt.Speed = float64(downloaded) / 1024 / elapsed // KB/s
	}
}

func (dt *DownloadTask) updateStatus(v status, message string) {
	dt.Status = v
	dt.Message = message
}
