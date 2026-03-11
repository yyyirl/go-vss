// @Title        streamcheck
// @Description  rtsp_check
// @Create       Dingshuai 2025/8/27 11:45

package ms

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RTSPChecker 用于检测RTSP流状态
type RTSPChecker struct {
	Timeout time.Duration
}

// NewRTSPChecker 创建一个新的RTSP检测器
func NewRTSPChecker(timeout time.Duration) *RTSPChecker {
	return &RTSPChecker{
		Timeout: timeout,
	}
}

// SessionInfo 存储RTSP会话信息
type SessionInfo struct {
	SessionID string
	Timeout   int
}

// CheckResult 存储检测结果
type CheckResult struct {
	IsOnline  bool
	Response  string
	MediaInfo string
	Error     error
}

// Check 检测RTSP流是否在线
func (r *RTSPChecker) Check(rtspURL string) (*CheckResult, error) {
	// 解析URL
	u, err := url.Parse(rtspURL)
	if err != nil {
		return nil, fmt.Errorf("无效的RTSP URL: %v", err)
	}

	// 确保是RTSP协议
	if strings.ToLower(u.Scheme) != "rtsp" {
		return nil, fmt.Errorf("仅支持RTSP协议")
	}

	// 提取主机和端口
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "554" // RTSP默认端口
	}
	address := fmt.Sprintf("%s:%s", host, port)

	// 建立TCP连接
	conn, err := net.DialTimeout("tcp", address, r.Timeout)
	if err != nil {
		return nil, fmt.Errorf("无法连接到服务器: %v", err)
	}
	defer conn.Close()

	// 设置读写超时
	conn.SetDeadline(time.Now().Add(r.Timeout))

	// 生成CSeq
	var (
		cSeq           = 1
		urlWithoutAuth = r.getURLWithoutAuth(rtspURL)
		// sessionInfo    *SessionInfo
	)

	// 发送OPTIONS请求获取服务器支持的方法
	optionsReq := fmt.Sprintf("OPTIONS %s RTSP/1.0\r\nCSeq: %d\r\n\r\n", rtspURL, cSeq)
	_, err = conn.Write([]byte(optionsReq))
	if err != nil {
		return nil, fmt.Errorf("发送OPTIONS请求失败: %v", err)
	}

	// 读取OPTIONS响应
	reader := bufio.NewReader(conn)
	optionsResponse, err := r.readRTSPResponse(reader)
	if err != nil {
		return nil, fmt.Errorf("读取OPTIONS响应失败: %v", err)
	}

	// 检测OPTIONS出现401的情况
	if strings.Contains(optionsResponse, "401 Unauthorized") && u.User != nil {
		// 提取认证信息
		authInfo := r.extractAuthInfo(optionsResponse)
		if authInfo == nil {
			return nil, fmt.Errorf("需要认证但无法解析认证信息")
		}

		// 重试带有认证信息的OPTIONS请求
		cSeq++
		username := u.User.Username()
		password, _ := u.User.Password()

		authHeader := r.generateAuthHeader(authInfo, "OPTIONS", urlWithoutAuth, username, password)
		reqWithAuth := fmt.Sprintf("OPTIONS %s RTSP/1.0\r\nCSeq: %d\r\nAccept: application/sdp\r\n%s\r\n\r\n",
			urlWithoutAuth, cSeq, authHeader)

		_, err = conn.Write([]byte(reqWithAuth))
		if err != nil {
			return nil, fmt.Errorf("发送带认证的OPTIONS请求失败: %v", err)
		}

		// 读取认证后的响应
		optionsResponse, err = r.readRTSPResponse(reader)
		if err != nil {
			return nil, fmt.Errorf("读取认证后的OPTIONS响应失败: %v", err)
		}
	}

	// 检查OPTIONS响应状态
	if !strings.Contains(optionsResponse, "200 OK") {
		return nil, fmt.Errorf("OPTIONS请求失败: %s", r.getStatusLine(optionsResponse))
	}

	cSeq++

	// 发送DESCRIBE请求
	var describeReq string

	if u.User != nil {
		// 需要认证
		describeReq = fmt.Sprintf("DESCRIBE %s RTSP/1.0\r\nCSeq: %d\r\nAccept: application/sdp\r\n\r\n",
			urlWithoutAuth, cSeq)
	} else {
		// 无需认证
		describeReq = fmt.Sprintf("DESCRIBE %s RTSP/1.0\r\nCSeq: %d\r\nAccept: application/sdp\r\n\r\n",
			rtspURL, cSeq)
	}

	_, err = conn.Write([]byte(describeReq))
	if err != nil {
		return nil, fmt.Errorf("发送DESCRIBE请求失败: %v", err)
	}

	// 读取DESCRIBE响应
	describeResponse, err := r.readRTSPResponse(reader)
	if err != nil {
		return nil, fmt.Errorf("读取DESCRIBE响应失败: %v", err)
	}

	// 检查是否需要认证
	if strings.Contains(describeResponse, "401 Unauthorized") && u.User != nil {
		// 提取认证信息
		authInfo := r.extractAuthInfo(describeResponse)
		if authInfo == nil {
			return nil, fmt.Errorf("需要认证但无法解析认证信息")
		}

		// 重试带有认证信息的DESCRIBE请求
		cSeq++
		username := u.User.Username()
		password, _ := u.User.Password()

		authHeader := r.generateAuthHeader(authInfo, "DESCRIBE", urlWithoutAuth, username, password)
		describeReqWithAuth := fmt.Sprintf("DESCRIBE %s RTSP/1.0\r\nCSeq: %d\r\nAccept: application/sdp\r\n%s\r\n\r\n",
			urlWithoutAuth, cSeq, authHeader)

		_, err = conn.Write([]byte(describeReqWithAuth))
		if err != nil {
			return nil, fmt.Errorf("发送带认证的DESCRIBE请求失败: %v", err)
		}

		// 读取认证后的响应
		describeResponse, err = r.readRTSPResponse(reader)
		if err != nil {
			return nil, fmt.Errorf("读取认证后的DESCRIBE响应失败: %v", err)
		}
	}

	// 检查DESCRIBE响应状态
	if strings.Contains(describeResponse, "200 OK") {
		// 提取会话信息（如果存在）
		// sessionInfo = r.extractSessionInfo(describeResponse)

		// 提取媒体信息
		mediaInfo := r.extractMediaInfo(describeResponse)

		result := &CheckResult{
			IsOnline:  true,
			Response:  describeResponse,
			MediaInfo: mediaInfo,
		}

		// 发送TEARDOWN请求关闭会话 (拉流会话尚未建立，这里是否应该发送TEARDOWN)
		// err := r.sendTeardown(conn, urlWithoutAuth, cSeq+1, u)
		// if err != nil {
		// 	fmt.Printf("发送TEARDOWN请求失败: %v\n", err)
		// }

		return result, nil
	}

	// 即使DESCRIBE失败也尝试清理，发送TEARDOWN请求 (拉流会话尚未建立，这里是否应该发送TEARDOWN)
	// if sessionInfo != nil {
	// r.sendTeardown(conn, urlWithoutAuth, cSeq+1, u)
	// }

	return &CheckResult{
		IsOnline: false,
		Response: describeResponse,
		Error:    fmt.Errorf("DESCRIBE请求失败: %s", r.getStatusLine(describeResponse)),
	}, nil
}

// sendTeardown 发送TEARDOWN请求关闭RTSP会话
func (r *RTSPChecker) sendTeardown(conn net.Conn, url string, cSeq int, u *url.URL) error {
	// 构建TEARDOWN请求
	teardownReq := fmt.Sprintf("TEARDOWN %s RTSP/1.0\r\nCSeq: %d\r\n\r\n",
		url, cSeq)

	// 如果需要认证，添加认证头
	if u.User != nil {
		// 这里简化处理，实际应用中可能需要重新获取nonce等认证信息
		username := u.User.Username()
		password, _ := u.User.Password()
		authInfo := &AuthInfo{Realm: "Streaming Server", Nonce: "1234567890"} // 简化处理
		authHeader := r.generateAuthHeader(authInfo, "TEARDOWN", url, username, password)
		teardownReq = fmt.Sprintf("TEARDOWN %s RTSP/1.0\r\nCSeq: %d\r\n%s\r\n\r\n",
			url, cSeq, authHeader)
	}

	// 发送TEARDOWN请求
	_, err := conn.Write([]byte(teardownReq))
	if err != nil {
		return fmt.Errorf("发送TEARDOWN请求失败: %v", err)
	}

	// 读取响应（可选，根据是否需要确认）
	reader := bufio.NewReader(conn)
	response, err := r.readRTSPResponse(reader)
	if err != nil {
		return fmt.Errorf("读取TEARDOWN响应失败: %v", err)
	}

	if !strings.Contains(response, "200") {
		return fmt.Errorf("TEARDOWN请求失败: %s", r.getStatusLine(response))
	}

	fmt.Println("RTSP会话已正确关闭")
	return nil
}

// readRTSPResponse 读取完整的RTSP响应
func (r *RTSPChecker) readRTSPResponse(reader *bufio.Reader) (string, error) {
	var (
		// response bytes.Buffer
		sipBytes = make([]byte, 1024)
	)
	// strings.Split(s,":")
	n, err := reader.Read(sipBytes)
	if err != nil {
		return "", err
	}
	return string(sipBytes[:n]), nil

	// for {
	// 	line, err := reader.ReadString('\n')
	// 	if err != nil {
	// 		return response.String(), err
	// 	}
	//
	// 	response.WriteString(line)
	//
	// 	// 检查是否到达响应末尾（空行）
	// 	if line == "\r\n" {
	// 		break
	// 	}
	// }
	// return response.String(), nil
}

// getStatusLine 从响应中提取状态行
func (r *RTSPChecker) getStatusLine(response string) string {
	lines := strings.Split(response, "\r\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return response
}

// getURLWithoutAuth 从URL中移除认证信息
func (r *RTSPChecker) getURLWithoutAuth(rtspURL string) string {
	u, err := url.Parse(rtspURL)
	if err != nil {
		return rtspURL
	}

	// 移除用户信息
	u.User = nil
	return u.String()
}

// AuthInfo 存储认证信息
type AuthInfo struct {
	Realm     string
	Nonce     string
	Algorithm string
}

// extractAuthInfo 从401响应中提取认证信息
func (r *RTSPChecker) extractAuthInfo(response string) *AuthInfo {
	lines := strings.Split(response, "\r\n")
	authInfo := &AuthInfo{}

	for _, line := range lines {
		if strings.HasPrefix(line, "WWW-Authenticate:") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) < 2 {
				continue
			}

			authLine := parts[1]
			if strings.HasPrefix(authLine, "Digest") {
				// 解析Digest认证参数
				re := regexp.MustCompile(`(\w+)="([^"]+)"`)
				matches := re.FindAllStringSubmatch(authLine, -1)

				for _, match := range matches {
					switch match[1] {
					case "realm":
						authInfo.Realm = match[2]
					case "nonce":
						authInfo.Nonce = match[2]
					case "algorithm":
						authInfo.Algorithm = match[2]
					}
				}
				return authInfo
			}
		}
	}

	return nil
}

// extractSessionInfo 从响应中提取会话信息
func (r *RTSPChecker) extractSessionInfo(response string) *SessionInfo {
	lines := strings.Split(response, "\r\n")
	sessionInfo := &SessionInfo{}

	for _, line := range lines {
		if strings.HasPrefix(line, "Session:") {
			parts := strings.SplitN(line, ";", 2)
			if len(parts) > 0 {
				sessionID := strings.TrimSpace(parts[0][8:]) // 移除"Session:"
				sessionInfo.SessionID = sessionID

				// 提取超时时间（如果存在）
				if len(parts) > 1 {
					timeoutParts := strings.Split(parts[1], "=")
					if len(timeoutParts) == 2 && strings.TrimSpace(timeoutParts[0]) == "timeout" {
						if timeout, err := strconv.Atoi(timeoutParts[1]); err == nil {
							sessionInfo.Timeout = timeout
						}
					}
				}
			}
			return sessionInfo
		}
	}

	return nil
}

// generateAuthHeader 生成认证头
func (r *RTSPChecker) generateAuthHeader(authInfo *AuthInfo, method, uri, username, password string) string {
	// 简化实现，只支持MD5算法
	ha1 := r.md5Hash(fmt.Sprintf("%s:%s:%s", username, authInfo.Realm, password))
	ha2 := r.md5Hash(fmt.Sprintf("%s:%s", method, uri))
	response := r.md5Hash(fmt.Sprintf("%s:%s:%s", ha1, authInfo.Nonce, ha2))

	return fmt.Sprintf("Authorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\", algorithem=\"MD5\"",
		username, authInfo.Realm, authInfo.Nonce, uri, response)
}

// md5Hash 计算MD5哈希值
func (r *RTSPChecker) md5Hash(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// extractMediaInfo 从DESCRIBE响应中提取媒体信息
func (r *RTSPChecker) extractMediaInfo(response string) string {
	lines := strings.Split(response, "\r\n")
	var mediaInfo strings.Builder
	inSDP := false

	for _, line := range lines {
		if line == "" {
			inSDP = true
			continue
		}

		if inSDP {
			if strings.HasPrefix(line, "m=") {
				mediaInfo.WriteString("媒体流: ")
				parts := strings.Split(line, " ")
				if len(parts) >= 4 {
					mediaType := parts[0]
					port := parts[1]
					protocol := parts[2]
					codec := parts[3]

					mediaInfo.WriteString(fmt.Sprintf("类型=%s, 端口=%s, 协议=%s, 编码=%s\n",
						mediaType[2:], port, protocol, codec))
				}
			} else if strings.HasPrefix(line, "a=control:") {
				mediaInfo.WriteString(fmt.Sprintf("控制URL: %s\n", line[10:]))
			} else if strings.HasPrefix(line, "a=rtpmap:") {
				parts := strings.Split(line, " ")
				if len(parts) >= 2 {
					payloadType := strings.Split(parts[0], ":")[1]
					codecInfo := parts[1]
					mediaInfo.WriteString(fmt.Sprintf("负载类型%s: %s\n", payloadType, codecInfo))
				}
			}
		}
	}

	return mediaInfo.String()
}

// TODO: RTSP检测在线调用示例
// streamcheck.RTSPCheckDemo("rtsp://admin:Ds123456@127.0.0.1/cam/realmonitor?channel=1&subtype=0")
// streamcheck.RTSPCheckDemo("rtsp://admin:Ds123456@192.168.0.109/cam/realmonitor?channel=1&subtype=0")
// streamcheck.RTSPCheckDemo("rtsp://admin:pass123456@192.168.0.101:554/h264/ch1/main/av_stream")
// streamcheck.RTSPCheckDemo("rtsp://admin:Ds123456@192.168.0.104:554/h264/ch1/main/av_stream")
func RTSPCheckDemo(rtspURL string) error {

	fmt.Printf("正在检测 RTSP 流: %s.....\n", rtspURL)

	// 设置5秒超时
	checker := NewRTSPChecker(5 * time.Second)
	result, err := checker.Check(rtspURL)
	if err != nil {
		fmt.Printf("RTSP流检测失败: %v\n", err)
		return err
	}

	if result.MediaInfo != "" {
		fmt.Printf("媒体信息: %s", result.MediaInfo)
	}

	if result.IsOnline {
		fmt.Printf("检测RTSP流完毕，结果: %s...流在线\n", rtspURL)
	} else {
		if result.Error != nil {
			fmt.Printf("错误信息: %v\n", result.Error)
		}
		fmt.Printf("检测RTSP流完毕，结果: %s...流离线\n", rtspURL)
	}

	return nil
}
