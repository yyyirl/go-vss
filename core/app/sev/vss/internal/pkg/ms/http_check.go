// @Title        streamcheck
// @Description  http_check
// @Create       Dingshuai 2025/8/28 14:02

package ms

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPChecker 流检测器
type HTTPChecker struct {
	Client    *http.Client
	Timeout   time.Duration
	UserAgent string
}

// StreamStatus 流状态信息
type StreamStatus struct {
	URL           string
	Protocol      string
	StreamType    string
	IsOnline      bool
	Error         string
	HTTPStatus    int
	ContentType   string
	ContentLength int
	HasVideo      bool
	HasAudio      bool
	Duration      time.Duration
	CheckedAt     time.Time
	Metadata      map[string]interface{}
}

// NewHTTPChecker 创建新的流检测器
func NewHTTPChecker(timeout time.Duration) *HTTPChecker {
	return &HTTPChecker{
		Client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     timeout,
				DisableCompression:  true,
				DisableKeepAlives:   false,
				MaxIdleConnsPerHost: 10,
			},
		},
		Timeout:   timeout,
		UserAgent: "SkeyeVSS/4.0 (compatible; Stream-Checker/1.0)",
	}
}

// CheckStream 检查流状态
func (c *HTTPChecker) CheckStream(streamURL string) (*StreamStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	status := &StreamStatus{
		IsOnline:  false,
		URL:       streamURL,
		CheckedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// 验证URL格式
	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		status.Error = err.Error()
		return status, err
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "GET", streamURL, nil)
	if err != nil {
		status.IsOnline = false
		status.Error = err.Error()
		return status, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	startTime := time.Now()
	resp, err := c.Client.Do(req)
	if err != nil {
		status.IsOnline = false
		status.Error = err.Error()
		return status, err
	}
	defer resp.Body.Close()

	// 记录响应时间
	status.Duration = time.Since(startTime)
	status.HTTPStatus = resp.StatusCode
	status.ContentType = resp.Header.Get("Content-Type")
	status.ContentLength = int(resp.ContentLength)

	// 检查HTTP状态
	if resp.StatusCode != http.StatusOK {
		status.IsOnline = false
		status.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
		return status, nil
	}

	// 根据内容类型检测流类型
	streamType := c.detectStreamType(status.ContentType, parsedURL.Path)
	status.StreamType = streamType

	// 根据流类型进行特定检测
	switch streamType {
	case "hls":
		return c.checkHLSStream(resp.Body, status)
	case "flv":
		return c.checkFLVStream(resp.Body, status)
	default:
		return c.checkGenericStream(resp.Body, status)
	}
}

// detectStreamType 检测流类型
func (c *HTTPChecker) detectStreamType(contentType, path string) string {
	// 根据Content-Type检测
	switch {
	case strings.Contains(contentType, "vnd.apple.mpegurl"),
		strings.Contains(contentType, "application/x-mpegurl"),
		strings.Contains(contentType, "application/vnd.apple.mpegurl"):
		return "hls"
	case strings.Contains(contentType, "video/x-flv"),
		strings.Contains(contentType, "application/x-flv"):
		return "flv"
	}

	// 根据文件扩展名检测
	switch {
	case strings.HasSuffix(path, ".m3u8"):
		return "hls"
	case strings.HasSuffix(path, ".m3u"):
		return "hls"
	case strings.HasSuffix(path, ".flv"):
		return "flv"
	}

	return "unknown"
}

// checkHLSStream 检查HLS流
func (c *HTTPChecker) checkHLSStream(body io.Reader, status *StreamStatus) (*StreamStatus, error) {
	status.Protocol = "HLS"

	// 读取部分内容进行分析
	reader := bufio.NewReader(body)
	firstBytes, err := reader.Peek(1024)
	if err != nil && err != io.EOF {
		status.Error = err.Error()
		status.IsOnline = false
		return status, err
	}

	content := string(firstBytes)
	lines := strings.Split(content, "\n")

	// 分析HLS特征
	isHLS := false
	hasExtM3U := false
	hasExtInf := false
	segmentCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "#EXTM3U" {
			hasExtM3U = true
		}
		if strings.HasPrefix(line, "#EXTINF:") {
			hasExtInf = true
			segmentCount++
		}
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			status.Metadata["is_master_playlist"] = true
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			segmentCount++
		}
	}

	// 判断是否为有效的HLS流
	if hasExtM3U && (hasExtInf || segmentCount > 0) {
		isHLS = true
		status.Metadata["segment_count"] = segmentCount
		status.Metadata["has_extm3u"] = hasExtM3U
		status.Metadata["has_extinf"] = hasExtInf
	}

	if isHLS {
		status.IsOnline = true
		status.Metadata["stream_type"] = "hls"
	} else {
		status.IsOnline = false
		status.Error = "不是有效的HLS流"
	}

	return status, nil
}

// checkFLVStream 检查FLV流
func (c *HTTPChecker) checkFLVStream(body io.Reader, status *StreamStatus) (*StreamStatus, error) {
	status.Protocol = "FLV"

	reader := bufio.NewReader(body)

	// 读取FLV头（9字节）
	header := make([]byte, 9)
	n, err := io.ReadFull(reader, header)
	if err != nil {
		status.IsOnline = false
		status.Error = err.Error()
		return status, err
	}

	// 验证FLV头
	if n < 9 || string(header[0:3]) != "FLV" {
		status.IsOnline = false
		status.Error = "不是有效的FLV文件"
		return status, nil
	}

	// 解析FLV头信息
	version := header[3]
	flags := header[4]
	hasVideo := (flags & 0x01) != 0
	hasAudio := (flags & 0x04) != 0

	status.HasVideo = hasVideo
	status.HasAudio = hasAudio
	status.Metadata["flv_version"] = version
	status.Metadata["has_video"] = hasVideo
	status.Metadata["has_audio"] = hasAudio

	// 尝试读取前一个Tag大小（4字节）
	prevTagSize := make([]byte, 4)
	_, err = io.ReadFull(reader, prevTagSize)
	if err != nil {
		status.IsOnline = false
		status.Error = "FLV流不完整"
		return status, nil
	}

	status.IsOnline = true
	return status, nil
}

// checkGenericStream 检查通用流
func (c *HTTPChecker) checkGenericStream(body io.Reader, status *StreamStatus) (*StreamStatus, error) {
	status.Protocol = "HTTP"
	status.StreamType = "generic"

	// 读取部分内容进行简单验证
	reader := bufio.NewReader(body)
	firstBytes, err := reader.Peek(512)
	if err != nil && err != io.EOF {
		status.IsOnline = false
		status.Error = err.Error()
		return status, err
	}

	if len(firstBytes) > 0 {
		status.IsOnline = true
		status.Metadata["content_sample"] = string(firstBytes[:min(100, len(firstBytes))])
	} else {
		status.IsOnline = false
		status.Error = "响应内容为空"
	}

	return status, nil
}

// CheckMultipleStreams 批量检查多个流
func (c *HTTPChecker) CheckMultipleStreams(urls []string) map[string]*StreamStatus {
	results := make(map[string]*StreamStatus)

	for _, url := range urls {
		status, err := c.CheckStream(url)
		if err != nil {
			status.Error = err.Error()
		}
		results[url] = status
	}

	return results
}

// min 辅助函数，返回最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func HTTPCheckDemo(httpURL string) error {

	fmt.Printf("正在检测 HTTP(s) 流: %s.....\n", httpURL)

	// 设置5秒超时
	checker := NewHTTPChecker(5 * time.Second)
	result, err := checker.CheckStream(httpURL)
	if err != nil {
		fmt.Printf("http(s)流检测失败: %v\n", err)
		return nil
	}

	if result.IsOnline {
		fmt.Printf("检测http(s)流完毕，结果: %s...%s流在线\n", httpURL, result.StreamType)
	} else {
		if result.Error != "" {
			fmt.Printf("错误信息: %v\n", result.Error)
		}
		fmt.Printf("检测http(s)流完毕，结果: %s...%s流离线\n", httpURL, result.StreamType)
	}

	return nil
}

func MultHTTPCheckDemo() {
	checker := NewHTTPChecker(5 * time.Second)

	// 测试URL列表
	testURLs := []string{
		"https://example.com/live/stream.m3u8",           // HLS流
		"https://example.com/live/stream.flv",            // FLV流
		"https://httpbin.org/status/200",                 // 普通HTTP
		"https://invalid-domain-that-does-not-exist.com", // 无效URL
	}

	results := checker.CheckMultipleStreams(testURLs)

	for url, status := range results {
		fmt.Printf("URL: %s\n", url)
		fmt.Printf("协议: %s\n", status.Protocol)
		fmt.Printf("流类型: %s\n", status.StreamType)
		fmt.Printf("状态: %v\n", status.IsOnline)
		fmt.Printf("HTTP状态: %d\n", status.HTTPStatus)
		fmt.Printf("响应时间: %v\n", status.Duration)
		fmt.Printf("内容类型: %s\n", status.ContentType)
		fmt.Printf("内容长度: %d bytes\n", status.ContentLength)

		if status.Error != "" {
			fmt.Printf("错误: %s\n", status.Error)
		}

		// 显示特定于流类型的元数据
		if status.StreamType == "hls" {
			if segments, ok := status.Metadata["segment_count"].(int); ok {
				fmt.Printf("分段数量: %d\n", segments)
			}
		} else if status.StreamType == "flv" {
			fmt.Printf("包含视频: %t\n", status.HasVideo)
			fmt.Printf("包含音频: %t\n", status.HasAudio)
		}
	}
}
