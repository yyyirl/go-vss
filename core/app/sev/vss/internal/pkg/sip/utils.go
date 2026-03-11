package sip

import (
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ghettovoice/gosip/sip"

	"skeyevss/core/app/sev/vss/internal/types"
	"skeyevss/core/pkg/functions"
)

var (
	sipLogChan          = make(chan *sipLogChanType, 100)
	sipLogFile          *os.File
	sipLogFileCreatedAt int64
)

type sipLogChanType struct {
	content,
	dir string
}

func init() {
	go logToFile()
}

func logToFile() {
	defer func() {
		_ = sipLogFile.Close()
	}()

	for {
		select {
		case data := <-sipLogChan:
			var now = functions.NewTimer().Now()
			if now-sipLogFileCreatedAt >= 24*3600 {
				_ = sipLogFile.Close()
				sipLogFile = nil
			}

			if sipLogFile == nil {
				var err error
				sipLogFile, err = os.OpenFile(path.Join(data.dir, fmt.Sprintf("%s.log", functions.NewTimer().Format("ymd"))), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println("日志文件的打开错误 :", err)
					continue
				}
				sipLogFileCreatedAt = functions.NewTimer().DayInitTimestamp(0)
			}

			if _, err := sipLogFile.WriteString(data.content); err != nil {
				fmt.Println("写入日志文件错误 :", err)
			}
		}
	}
}

func ParseToRequest(req sip.Request) (*types.Request, error) {
	from, ok := req.From()
	if !ok || from.Address == nil {
		return nil, errors.New("OnRegister, no from")
	}

	return &types.Request{
		ID:       from.Address.User().String(),
		Source:   req.Source(),
		Body:     req.Body(),
		Original: req,
		DeviceAddr: sip.Address{
			DisplayName: from.DisplayName,
			Uri:         from.Address,
		},
		Authorization:     req.GetHeaders("Authorization"),
		TransportProtocol: getTransportProtocol(req),
	}, nil
}

func getTransportProtocol(req sip.Request) string {
	viaHop, ok := req.ViaHop()
	if !ok {
		return "UDP"
	}

	var transport = strings.ToUpper(strings.TrimSpace(viaHop.Transport))
	switch transport {
	case "UDP", "TCP", "TLS", "WS", "WSS":
		return transport
	default:
		return "UDP"
	}
}

func XmlEncode(v interface{}) (string, error) {
	data, err := xml.MarshalIndent(v, "", " ")
	if err != nil {
		return "", err
	}

	return `<?xml version="1.0" encoding="GB2312"?>` + "\n" + string(data) + "\n", nil
}

func SipLog(useSipLogToFile bool, logPath, rType, prefix string, req sip.Request, content string) {
	// functions.XLogInfo("\n\n", prefix, ": START ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	// functions.LogInfo(content)
	// // b, _ := functions.JSONMarshal(strings.Split(content, "\n"))
	// // functions.LogInfo(string(b))
	// functions.XLogInfo("\n", prefix, ": END ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++\n\n")

	var link string
	switch rType {
	case types.BroadcastTypeSipRequest:
		link = fmt.Sprintf("[%s]>>>>>>[%s]>>>>>>", req.Source(), req.Destination())
	case types.BroadcastTypeSipReceive:
		link = fmt.Sprintf("[%s]<<<<<<[%s]<<<<<<", req.Destination(), req.Source())
	case types.BroadcastTypeSipResponse:
		link = fmt.Sprintf("[%s]>>>>>>[%s]>>>>>>", req.Destination(), req.Source())
	case types.BroadcastTypeSipReceiveResponse:
		link = fmt.Sprintf("[%s]<<<<<<[%s]<<<<<<", req.Source(), req.Destination())
	}

	if !useSipLogToFile {
		functions.PrintStyle("red", fmt.Sprintf("[%s] %s [%s]%s", prefix, time.Now().Format("2006-01-02 15:04:05"), req.Transport(), link))
		functions.PrintStyle("blue", content)
		return
	}

	var sipLog = "\n" + fmt.Sprintf("[%s] %s [%s]%s", prefix, time.Now().Format("2006-01-02 15:04:05"), req.Transport(), link) + "\n" + content
	sipLogChan <- &sipLogChanType{
		content: sipLog,
		dir:     logPath,
	}
}

const (
	PresetSet  = 0x81
	PresetCall = 0x82
	PresetDel  = 0x83
)

// 设备控制 ---------------------------------------------

const PTZFirstByte = 0xA5

func getAssembleCode() uint8 {
	return (PTZFirstByte>>4 + PTZFirstByte&0xF + 0) % 16
}

func getVerificationCode(ptz []byte) {
	sum := uint8(0)
	for i := 0; i < len(ptz)-1; i++ {
		sum += ptz[i]
	}
	ptz[len(ptz)-1] = sum
}

// 注1 : 字节4 中的 Bit5、Bit4 分别控制镜头变倍的缩小和放大, 字节4 中的 Bit3、Bit2、Bit1、Bit0 位分别控制云台
//
//	上、 下、 左、 右方向的转动, 相应 Bit 位置1 时, 启动云台向相应方向转动, 相应 Bit 位清0 时, 停止云台相应
//	方向的转动。 云台的转动方向以监视器显示图像的移动方向为准。
//
// 注2: 字节5 控制水平方向速度, 速度范围由慢到快为00H~FFH; 字节6 控制垂直方向速度, 速度范围由慢到快
//
//	为00H-FFH。
//
// 注3: 字节7 的高4 位为变焦速度, 速度范围由慢到快为0H~FH; 低4 位为地址的高4 位。
type SipDeviceControlPtz struct {
	ZoomOut bool
	ZoomIn  bool
	Up      bool
	Down    bool
	Left    bool
	Right   bool
	Speed   byte // 0-8
}

func (p *SipDeviceControlPtz) Pack() string {
	var buf = make([]byte, 8)
	buf[0] = PTZFirstByte
	buf[1] = getAssembleCode()
	buf[2] = 1
	buf[4] = 0
	buf[5] = 0
	buf[6] = 0
	if p.ZoomOut {
		buf[3] |= 1 << 5
		buf[6] = p.Speed << 4
	}

	if p.ZoomIn {
		buf[3] |= 1 << 4
		buf[6] = p.Speed << 4
	}
	if p.Up {
		buf[3] |= 1 << 3
		buf[5] = p.Speed
	}
	if p.Down {
		buf[3] |= 1 << 2
		buf[5] = p.Speed
	}
	if p.Left {
		buf[3] |= 1 << 1
		buf[4] = p.Speed
	}
	if p.Right {
		buf[3] |= 1
		buf[4] = p.Speed
	}
	getVerificationCode(buf)
	return hex.EncodeToString(buf)
}

func (p *SipDeviceControlPtz) Stop() string {
	var buf = make([]byte, 8)
	buf[0] = PTZFirstByte
	buf[1] = getAssembleCode()
	buf[2] = 1
	buf[3] = 0
	buf[4] = 0
	buf[5] = 0
	buf[6] = 0
	getVerificationCode(buf)
	return hex.EncodeToString(buf)
}

/*
注1 : 字节4 中的 Bit3 为1 时, 光圈缩小;Bit2 为1 时, 光圈放大。 Bit1 为1 时, 聚焦近;Bit0 为1 时, 聚焦远。 Bit3~

	Bit0 的相应位清0, 则相应控制操作停止动作。

注2: 字节5 表示聚焦速度, 速度范围由慢到快为00H~FFH。
注3: 字节6 表示光圈速度, 速度范围由慢到快为00H~FFH
*/
type SipDeviceControlFi struct {
	IrisIn    bool
	IrisOut   bool
	FocusNear bool
	FocusFar  bool
	Speed     byte // 0-8
}

func (f *SipDeviceControlFi) Pack() string {
	var buf = make([]byte, 8)
	buf[0] = PTZFirstByte
	buf[1] = getAssembleCode()
	buf[2] = 1
	buf[3] |= 1 << 6

	buf[4] = 0
	buf[5] = 0
	buf[6] = 0
	if f.IrisIn {
		buf[3] |= 1 << 3
		buf[5] = f.Speed
	}
	if f.IrisOut {
		buf[3] |= 1 << 2
		buf[5] = f.Speed
	}
	if f.FocusNear {
		buf[3] |= 1 << 1
		buf[4] = f.Speed
	}
	if f.FocusFar {
		buf[3] |= 1
		buf[4] = f.Speed
	}
	getVerificationCode(buf)
	return hex.EncodeToString(buf)
}

func (f *SipDeviceControlFi) Stop() string {
	var buf = make([]byte, 8)
	buf[0] = PTZFirstByte
	buf[1] = getAssembleCode()
	buf[2] = 1
	buf[3] = 0
	buf[4] = 0
	buf[5] = 0
	buf[6] = 0
	getVerificationCode(buf)
	return hex.EncodeToString(buf)
}

type Preset struct {
	CMD   byte
	Point byte
}

func (p *Preset) Pack() string {
	var buf = make([]byte, 8)
	buf[0] = PTZFirstByte
	buf[1] = getAssembleCode()
	buf[2] = 1

	buf[3] = p.CMD

	buf[4] = 0
	buf[5] = p.Point
	buf[6] = 0
	getVerificationCode(buf)
	return hex.EncodeToString(buf)
}

func VideoRecordMapKey(deviceUniqueId, channelUniqueId string, sn int64) string {
	return functions.Md5String(fmt.Sprintf("%s%s%d", deviceUniqueId, channelUniqueId, sn))
}

func MakeCascadeUserAgent(productName, internalIp string) string {
	return fmt.Sprintf("%s %s", productName, internalIp)
}

func ParsePtzCmd(ptzcmd string) (*types.DeviceControlReq, error) {
	ptz, _ := hex.DecodeString(ptzcmd)
	if len(ptz) != 8 {
		return nil, fmt.Errorf("ptz command must be 8 bytes")
	}

	b1, b2, b3, b4, b5, b6, b7 := 0xA5, 0x0F, 0x01, 0x00, 0x00, 0x00, 0x00
	b1 = int(ptz[0])
	b2 = int(ptz[1])
	b3 = int(ptz[2])
	b4 = int(ptz[3])
	b5 = int(ptz[4])
	b6 = int(ptz[5])
	b7 = int(ptz[6])

	var (
		b8    = int(ptz[7])
		b8Tmp = (b1 + b2 + b3 + b4 + b5 + b6 + b7) % 256
	)
	if b8 != b8Tmp {
		return nil, fmt.Errorf("ptz command must be 256 bytes")
	}

	var data = new(types.DeviceControlReq)
	switch b4 {
	case 0x01: // right
		data.Horizontal = -1
		data.Speed = b5

	case 0x02: // left
		data.Horizontal = 1
		data.Speed = b5

	case 0x04: // down
		data.Vertical = -1
		data.Speed = b6
	case 0x08: // up
		data.Vertical = 1
		data.Speed = b6
	case 0x04 | 0x01: // downright
		data.Vertical = -1
		data.Horizontal = -1
		if b5 > b6 {
			data.Speed = b5
		} else {
			data.Speed = b6
		}

	case 0x04 | 0x02: // downleft
		data.Vertical = -1
		data.Horizontal = 1
		if b5 > b6 {
			data.Speed = b5
		} else {
			data.Speed = b6
		}

	case 0x08 | 0x01: // upright
		data.Vertical = 1
		data.Horizontal = -1
		if b5 > b6 {
			data.Speed = b5
		} else {
			data.Speed = b6
		}

	case 0x08 | 0x02: // upleft
		data.Vertical = 1
		data.Horizontal = 1
		if b5 > b6 {
			data.Speed = b5
		} else {
			data.Speed = b6
		}

	case 0x10: // zoomin
		data.Minifier = 1
		data.Speed = (b7 >> 4) * 17

	case 0x20: // zoomout
		data.Minifier = -1
		data.Speed = (b7 >> 4) * 17

	case 0x00:
		data.Stop = true
		data.Speed = 100
	}
	data.Speed = data.Speed / 25

	return data, nil
}

func WSTalkKey(deviceUniqueId, channelUniqueId string) string {
	return fmt.Sprintf("%s-%s", deviceUniqueId, channelUniqueId)
}
