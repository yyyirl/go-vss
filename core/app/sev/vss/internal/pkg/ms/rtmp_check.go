// @Title        streamcheck
// @Description  rtmp_check
// @Create       Dingshuai 2025/8/27 16:42

package ms

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	rtmp0 "skeyevss/core/app/sev/vss/internal/pkg/ms/rtmp"
)

// RTMPChecker 用于检测RTMP流状态
type RTMPChecker struct {
	Timeout time.Duration
	rtmp0.HandshakeClientComplex
	packer *rtmp0.MessagePacker
	hc     rtmp0.IHandshakeClient
}

// NewRTMPChecker 创建一个新的RTMP检测器
func NewRTMPChecker(timeout time.Duration, HandshakeComplexFlag bool) *RTMPChecker {
	var (
		hc rtmp0.IHandshakeClient
	)

	if HandshakeComplexFlag {
		hc = &rtmp0.HandshakeClientComplex{}
	} else {
		hc = &rtmp0.HandshakeClientSimple{}
	}
	return &RTMPChecker{
		Timeout: timeout,
		packer:  rtmp0.NewMessagePacker(),
		hc:      hc,
	}
}

// RTMPCheckResult 存储检测结果
type RTMPCheckResult struct {
	IsOnline bool
	Error    error
	Metadata string
}

// Check 检测RTMP流是否在线
func (r *RTMPChecker) Check(rtmpURL string) (*RTMPCheckResult, error) {
	// 解析URL
	u, err := url.Parse(rtmpURL)
	if err != nil {
		return nil, fmt.Errorf("无效的RTMP URL: %v", err)
	}

	// 确保是RTMP协议
	if strings.ToLower(u.Scheme) != "rtmp" {
		return nil, fmt.Errorf("仅支持RTMP协议")
	}

	// 提取主机和端口
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "1935" // RTMP默认端口
	}
	address := fmt.Sprintf("%s:%s", host, port)

	// 建立TCP连接
	conn, err := net.DialTimeout("tcp", address, r.Timeout)
	if err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("无法连接到服务器: %v", err),
		}, nil
	}
	defer conn.Close()

	// 设置读写超时
	conn.SetDeadline(time.Now().Add(r.Timeout))

	// 执行RTMP握手
	err = r.handshake(conn)
	if err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("RTMP握手失败: %v", err),
		}, nil
	}

	// 连接RTMP应用
	streamKey := strings.TrimPrefix(u.Path, "/")
	if streamKey == "" {
		streamKey = "live"
	}

	// s.log.Infof("[%s] > W SetChunkSize %d.", s.UniqueKey(), LocalChunkSize)
	if err := r.packer.WriteChunkSize(conn, rtmp0.LocalChunkSize); err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("writeChunkSize失败: %v", err),
		}, nil
	}

	toURL := fmt.Sprintf("%s://%s/%s", u.Scheme, address, streamKey)
	// s.log.Infof("[%s] > W connect('%s'). tcUrl=%s", s.UniqueKey(), s.appName(), s.tcUrl())
	if err := r.packer.WriteConnect(conn, streamKey, toURL, false); err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("发送连接命令失败: %v", err),
		}, nil
	}

	// // 发送连接命令
	// err = r.sendConnectCommand(conn, u.Host, streamKey, u.User)
	// if err != nil {
	// 	return &RTMPCheckResult{
	// 		IsOnline: false,
	// 		Error:    fmt.Errorf("发送连接命令失败: %v", err),
	// 	}, nil
	// }

	// 读取服务器响应
	result, err := r.readServerResponse(conn, streamKey)
	if err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("读取服务器响应失败: %v", err),
		}, nil
	}

	return result, nil
}

// 由于Go标准库没有提供随机数生成，这里添加一个简单的实现
func randRead(b []byte) (n int, err error) {
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
	}
	return len(b), nil
}

func (r *RTMPChecker) handshake(conn net.Conn) error {
	// s.log.Infof("[%s] > W Handshake C0+C1.", s.UniqueKey())
	if err := r.hc.WriteC0C1(conn); err != nil {
		return err
	}

	if err := r.hc.ReadS0S1(conn); err != nil {
		return err
	}
	// s.log.Infof("[%s] < R Handshake S0+S1.", s.UniqueKey())

	// s.log.Infof("[%s] > W Handshake C2.", s.UniqueKey())
	if err := r.hc.WriteC2(conn); err != nil {
		return err
	}

	if err := r.hc.ReadS2(conn); err != nil {
		return err
	}
	// s.log.Infof("[%s] < R Handshake S2.", s.UniqueKey())
	return nil
}

// readServerResponse 读取服务器响应
func (r *RTMPChecker) readServerResponse(conn net.Conn, streamKey string) (*RTMPCheckResult, error) {
	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	// 尝试读取服务器响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("读取响应失败: %v", err),
		}, nil
	}

	// 分析响应数据
	if n > 0 {
		// 检查是否包含成功的响应标识
		// 实际实现需要解析RTMP协议消息
		if bytes.Contains(buffer[:n], []byte("_result")) ||
			bytes.Contains(buffer[:n], []byte("onStatus")) {
			return &RTMPCheckResult{
				IsOnline: true,
				Metadata: fmt.Sprintf("响应数据: %s", hex.EncodeToString(buffer[:n])),
			}, nil
		}

		// 检查是否包含错误响应
		if bytes.Contains(buffer[:n], []byte("_error")) ||
			bytes.Contains(buffer[:n], []byte("error")) {
			return &RTMPCheckResult{
				IsOnline: false,
				Error:    fmt.Errorf("服务器返回错误: %s", hex.EncodeToString(buffer[:n])),
			}, nil
		}
	}

	// 如果收到任何响应，认为服务器在线但可能需要进一步验证
	return &RTMPCheckResult{
		IsOnline: true,
		Metadata: "收到服务器响应，但无法解析具体内容",
	}, nil
}

// 简化版RTMP检测（通过尝试连接和简单握手）
func (r *RTMPChecker) simpleCheck(rtmpURL string) (*RTMPCheckResult, error) {
	// 解析URL
	u, err := url.Parse(rtmpURL)
	if err != nil {
		return nil, fmt.Errorf("无效的RTMP URL: %v", err)
	}
	// 确保是RTMP协议
	if strings.ToLower(u.Scheme) != "rtmp" {
		return nil, fmt.Errorf("仅支持RTMP协议")
	}
	// 提取主机和端口
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "1935" // RTMP默认端口
	}
	address := fmt.Sprintf("%s:%s", host, port)

	// 建立TCP连接
	conn, err := net.DialTimeout("tcp", address, r.Timeout)
	if err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("无法连接到服务器: %v", err),
		}, nil
	}
	defer conn.Close()

	// 设置读写超时
	conn.SetDeadline(time.Now().Add(r.Timeout))

	// 发送RTMP握手初始字节
	// 执行RTMP握手
	err = r.handshake(conn)
	if err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("RTMP握手失败: %v", err),
		}, nil
	}

	// 尝试读取响应
	response := make([]byte, 1)
	_, err = conn.Read(response)
	if err != nil {
		return &RTMPCheckResult{
			IsOnline: false,
			Error:    fmt.Errorf("读取握手响应失败: %v", err),
		}, nil
	}

	// 检查响应是否为有效的RTMP版本
	if response[0] == 0x03 {
		return &RTMPCheckResult{
			IsOnline: true,
			Metadata: "RTMP服务器响应成功",
		}, nil
	}

	return &RTMPCheckResult{
		IsOnline: false,
		Error:    fmt.Errorf("无效的RTMP响应: %x", response[0]),
	}, nil
}

func RTMPCheckDemo(rtmpURL string) error {

	fmt.Printf("正在检测 RTMP 流......: %s\n", rtmpURL)

	checker := NewRTMPChecker(5*time.Second, false)

	result, err := checker.Check(rtmpURL)
	// 使用简化版检测（完整实现需要更复杂的RTMP协议处理）
	// result, err := checker.simpleCheck(rtmpURL)
	if err != nil {
		fmt.Printf("检测失败: %v\n", err)
		return err
	}

	if result.IsOnline {
		fmt.Println("RTMP流在线")
		if result.Metadata != "" {
			fmt.Println("详细信息:", result.Metadata)
		}
	} else {
		fmt.Println("RTMP流离线")
		if result.Error != nil {
			fmt.Printf("错误信息: %v\n", result.Error)
			return result.Error
		}
	}
	return nil
}
