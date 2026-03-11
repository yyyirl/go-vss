// @Title        analysis
// @Description  main
// @Create       yiyiyi 2025/9/9 16:04

package pprof

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PProfParams struct {
	// cpu heap goroutine block mutex
	Services []ServiceConfig `json:"services" validate:"required"`
	Type     string          `json:"type" validate:"oneof=cpu heap goroutine block mutex"`
	Duration int             `json:"duration" validate:"required"`

	Dir string          `json:"-"`
	Ctx context.Context `json:"-"`
}

type ServiceConfig struct {
	Name string `json:"name" validate:"required"`
	Host string `json:"host" validate:"required"`
	Port int    `json:"port" validate:"required"`
}

// {"services":[{"name":"backendapi","host":"localhost","port":11020},{"name":"dbrpc","host":"localhost","port":11021}],"type":"cpu","duration":30}

func NewAnalyzePProf(req *PProfParams) (map[string]interface{}, []string) {
	// 验证类型
	if req.Type == "cpu" && req.Duration <= 0 {
		req.Duration = 30 // 默认30秒
	}

	var wg sync.WaitGroup
	results := make(map[string]interface{})
	filePaths := make([]string, 0)
	var mu sync.Mutex

	for _, service := range req.Services {
		wg.Add(1)
		go func(svc ServiceConfig) {
			defer wg.Done()

			profileData, filePath, err := fetchPProf(svc, req.Type, req.Duration, req.Dir, req.Ctx)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				results[svc.Name] = map[string]interface{}{
					"success": false,
					"error":   err.Error(),
				}
			} else {
				results[svc.Name] = map[string]interface{}{
					"success": true,
					"data":    profileData,
					"file":    filePath,
				}
				if filePath != "" {
					filePaths = append(filePaths, filePath)
				}
			}
		}(service)
	}

	wg.Wait()

	return results, filePaths
}

func fetchPProf(service ServiceConfig, profileType string, duration int, dir string, ctx context.Context) (interface{}, string, error) {
	baseURL := fmt.Sprintf("http://%s:%d/debug/pprof", service.Host, service.Port)
	var profileURL string

	switch profileType {
	case "cpu":
		profileURL = fmt.Sprintf("%s/profile?seconds=%d", baseURL, duration)
	case "heap":
		profileURL = baseURL + "/heap"
	case "goroutine":
		profileURL = baseURL + "/goroutine"
	case "block":
		profileURL = baseURL + "/block"
	case "mutex":
		profileURL = baseURL + "/mutex"
	default:
		return nil, "", fmt.Errorf("不支持的profile类型: %s", profileType)
	}

	// 下载profile文件
	req, err := http.NewRequestWithContext(ctx, "GET", profileURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %v", err)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("请求失败: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("服务返回错误: %s", resp.Status)
	}

	// 保存文件
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s-%s.pprof", service.Name, profileType, timestamp)
	filePath := filepath.Join(dir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return nil, "", fmt.Errorf("写入文件失败: %v", err)
	}

	defer func() {
		// 生成SVG
		if _, err := NewPProfSVGGenerator(filePath).WithOutputDir(dir).GenerateSVG(); err != nil {
			log.Printf("svg生成失败: %v", err)
		}
	}()

	// 对于非CPU类型，可以尝试解析简单信息
	if profileType != "cpu" {
		summary, err := parseProfileSummary(filePath, profileType)
		if err != nil {
			log.Printf("解析profile摘要失败: %v", err)
			// 仍然返回文件路径
			return map[string]string{"file": filename}, filename, nil
		}

		return summary, filename, nil
	}

	return map[string]string{
		"file":     filename,
		"duration": fmt.Sprintf("%ds", duration),
	}, filename, nil
}

func parseProfileSummary(filePath, profileType string) (map[string]interface{}, error) {
	// 这里可以使用go tool pprof的命令行输出解析
	// 简化版：返回文件信息
	return map[string]interface{}{
		"file":       filepath.Base(filePath),
		"type":       profileType,
		"size_bytes": getFileSize(filePath),
	}, nil
}

func getFileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return info.Size()
}
