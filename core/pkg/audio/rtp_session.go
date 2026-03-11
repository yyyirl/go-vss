package audio

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/ghettovoice/gosip/sip"
	"github.com/pion/rtp"
)

const (
	AudioCodecPcma = "PCMA"
	AudioCodecPcmu = "PCMU"
	AudioCodecPs   = "PS"
	AudioCodecAAC  = "MPEG4-GENERIC"

	AudioPayloadTypeDefault = 8
	AudioSampleRateDefault  = 8000
)

type (
	RTPSession struct {
		localRTPPort,
		localRTCPPort int
		remoteRTPAddr,
		remoteRTCPAddr *net.UDPAddr
		rtpConn,
		rtcpConn *net.UDPConn
		sequenceNum uint16
		timestamp,
		ssrc,
		sampleRate uint32
		payloadType uint8
		audioCodec  string
		startTime   time.Time
		sendCount   int // 发送包计数器（用于调试）
	}

	TalkSessionItem struct {
		Status     bool
		ActivateAt int64
		CallID     string
		SSRC       uint32

		// sip invite发送过来的数据
		RTPRemoteIP string
		RTPRtpPort,
		RTPRtcpPort,
		RTPUsablePort int
		RTPSession *RTPSession

		AudioCodec string
		AudioPayloadType,
		AudioSampleRate int

		ACKReq sip.Request
	}
)

func NewRTPSession(item *TalkSessionItem) (*RTPSession, error) {
	if item.SSRC == 0 {
		item.SSRC = rand.Uint32()
	}

	var (
		payloadType = AudioPayloadTypeDefault
		sampleRate  = AudioSampleRateDefault
		audioCodec  = AudioCodecPcma
	)
	if item.AudioPayloadType <= 0 {
		item.AudioPayloadType = payloadType
	}

	if item.AudioSampleRate <= 0 {
		item.AudioSampleRate = sampleRate
	}

	if item.AudioCodec == "" {
		item.AudioCodec = audioCodec
	}

	var (
		err     error
		session = &RTPSession{
			sequenceNum: uint16(rand.Intn(0xFFFF)),
			timestamp:   0,
			ssrc:        item.SSRC,
			payloadType: uint8(item.AudioPayloadType),
			sampleRate:  uint32(item.AudioSampleRate),
			audioCodec:  item.AudioCodec,
			startTime:   time.Now(),
			sendCount:   0,
		}
	)
	session.remoteRTPAddr, err = net.ResolveUDPAddr(
		"udp",
		fmt.Sprintf("%s:%d", item.RTPRemoteIP, item.RTPRtpPort),
	)
	if err != nil {
		return nil, fmt.Errorf("resolve remote RTP addr failed: %v", err)
	}

	session.remoteRTCPAddr, err = net.ResolveUDPAddr(
		"udp",
		fmt.Sprintf("%s:%d", item.RTPRemoteIP, item.RTPRtcpPort),
	)
	if err != nil {
		return nil, fmt.Errorf("resolve remote RTCP addr failed: %v", err)
	}

	// 绑定本地端口
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", item.RTPUsablePort))
	if err != nil {
		return nil, fmt.Errorf("resolve local addr failed: %v", err)
	}

	// 创建RTP连接
	session.rtpConn, err = net.DialUDP("udp", localAddr, session.remoteRTPAddr)
	if err != nil {
		return nil, fmt.Errorf("create RTP connection failed: %v", err)
	}

	// 获取实际绑定的本地端口
	var localRTPAddr = session.rtpConn.LocalAddr().(*net.UDPAddr)
	session.localRTPPort = localRTPAddr.Port
	session.localRTCPPort = item.RTPUsablePort + 1
	return session, nil
}

// 发送单个RTP音频包
func (s *RTPSession) SendAudioPacket(audioData []byte, samplesPerPacket int) error {
	if s.rtpConn == nil {
		return fmt.Errorf("RTP connection not established")
	}

	if len(audioData) == 0 {
		return fmt.Errorf("audio data is empty")
	}

	var (
		// 创建RTP包
		packet = &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				Padding:        false,
				Extension:      false,
				Marker:         false, // 语音包通常为false
				PayloadType:    s.payloadType,
				SequenceNumber: s.sequenceNum,
				Timestamp:      s.timestamp,
				SSRC:           s.ssrc,
			},
			Payload: audioData,
		}

		// 编码RTP包
		packetBuf = make([]byte, 1500)
	)
	n, err := packet.MarshalTo(packetBuf)
	if err != nil {
		return fmt.Errorf("RTP marshal failed: %v", err)
	}

	// 发送RTP包
	if _, err := s.rtpConn.Write(packetBuf[:n]); err != nil {
		return fmt.Errorf("RTP write failed: %v", err)
	}

	s.sendCount++
	// 更新序列号和时间戳
	s.sequenceNum++
	// 正确的时间戳增量：每个样本增加1（对于8000Hz的G.711）
	// samplesPerPacket = len(audioData) 因为G.711是8位，每个字节一个样本
	s.timestamp += uint32(samplesPerPacket)
	return nil
}

func (s *RTPSession) SendAudioStream(pcmData []byte) error {
	if len(pcmData) == 0 {
		return fmt.Errorf("pcm data is empty")
	}

	// 转g711
	var (
		encodedData []byte
		err         error
	)
	if strings.ToUpper(s.audioCodec) == AudioCodecAAC {
		// TODO: 这里应该改成 aac编码
		encodedData, err = G711AEncode(pcmData)
	} else if strings.ToUpper(s.audioCodec) == AudioCodecPcmu {
		// pcmu
		encodedData, err = G711UEncode(pcmData)
	} else {
		// pcma
		encodedData, err = G711AEncode(pcmData)
	}

	if err != nil {
		return fmt.Errorf("G.711 %s encoding failed: %v", s.audioCodec, err)
	}

	return s.SendAudioPacket(encodedData, len(encodedData))
}

func (s *RTPSession) Stop() {
	if s.rtpConn != nil {
		_ = s.rtpConn.Close()
	}

	if s.rtcpConn != nil {
		_ = s.rtcpConn.Close()
	}
}
