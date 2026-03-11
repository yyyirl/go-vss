// Copyright 2020, Chef.  All rights reserved.
// https://gitee.com/openskeye-lab/skeyesms
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package base

import (
	"encoding/hex"
	"fmt"
	"time"
)

// av codec
const (
	// AudioCodecAac StatGroup.AudioCodec
	AudioCodecAac   = "AAC"
	AudioCodecG711U = "PCMU"
	AudioCodecG711A = "PCMA"
	AudioCodecOpus  = "OPUS"

	// VideoCodecAvc StatGroup.VideoCodec
	VideoCodecAvc  = "H264"
	VideoCodecHevc = "H265"
	VideoCodecAV1  = "AV1"
)

const (
	// RtmpTypeIdAudio spec-rtmp_specification_1.0.pdf
	// 7.1. Types of Messages
	RtmpTypeIdAudio        uint8 = 8
	RtmpTypeIdVideo        uint8 = 9
	RtmpTypeIdMetadata     uint8 = 18 // RtmpTypeIdDataMessageAmf0
	RtmpTypeIdSetChunkSize uint8 = 1
	// RtmpTypeIdAck 和 RtmpTypeIdWinAckSize 的含义：
	//
	// 一端向另一端发送 RtmpTypeIdWinAckSize ，要求对端每收够一定数据（一定数据的阈值包含在 RtmpTypeIdWinAckSize 信令中）后，向本端回复 RtmpTypeIdAck 。
	//
	// 常见的应用场景：数据发送端要求数据接收端定时发送心跳信令给本端。
	RtmpTypeIdAck         uint8 = 3
	RtmpTypeIdUserControl uint8 = 4
	// RtmpTypeIdWinAckSize 见 RtmpTypeIdAck
	RtmpTypeIdWinAckSize         uint8 = 5
	RtmpTypeIdBandwidth          uint8 = 6
	RtmpTypeIdCommandMessageAmf3 uint8 = 17
	RtmpTypeIdCommandMessageAmf0 uint8 = 20
	RtmpTypeIdAggregateMessage   uint8 = 22

	// RtmpUserControlStreamBegin RtmpUserControlXxx...
	//
	// user control message type
	//
	RtmpUserControlStreamBegin  uint8 = 0
	RtmpUserControlRecorded     uint8 = 4
	RtmpUserControlPingRequest  uint8 = 6
	RtmpUserControlPingResponse uint8 = 7

	// RtmpFrameTypeKey spec-video_file_format_spec_v10.pdf
	// Video tags
	//   VIDEODATA
	//     FrameType UB[4]
	//     CodecId   UB[4]
	//   AVCVIDEOPACKET
	//     AVCPacketType   UI8
	//     CompositionTime SI24
	//     Data            UI8[n]
	RtmpFrameTypeKey   uint8 = 1
	RtmpFrameTypeInter uint8 = 2

	// RtmpCodecIdAvc
	//
	// Video tags -> VIDEODATA -> CodecID
	//
	// 1: JPEG (currently unused)
	// 2: Sorenson H.263
	// 3: Screen video
	// 4: On2 VP6
	// 5: On2 VP6 with alpha channel
	// 6: Screen video version 2
	// 7: AVC
	// 12: HEVC
	// 13: AV1 https://github.com/tencentyun/flv/blob/main/FLV_Codec.md
	RtmpCodecIdAvc  uint8 = 7
	RtmpCodecIdHevc uint8 = 12
	RtmpCodecIdAV1  uint8 = 13

	// RtmpAvcPacketTypeSeqHeader RtmpAvcPacketTypeNalu RtmpHevcPacketTypeSeqHeader RtmpHevcPacketTypeNalu
	// 注意，按照标准文档上描述，PacketType还有可能为2：
	// 2: AVC end of sequence (lower level NALU sequence ender is not required or supported)
	//
	// 我自己遇到过在流结尾时，对端发送 27 02 00 00 00的情况（比如我们的使用wontcry.flv的单元测试，最后一个包）
	//
	RtmpAvcPacketTypeSeqHeader  uint8 = 0
	RtmpAvcPacketTypeNalu       uint8 = 1
	RtmpHevcPacketTypeSeqHeader       = RtmpAvcPacketTypeSeqHeader
	RtmpHevcPacketTypeNalu            = RtmpAvcPacketTypeNalu
	RtmpAV1PacketTypeSeqHeader        = RtmpAvcPacketTypeSeqHeader
	RtmpAV1PacketTypeNalu             = RtmpAvcPacketTypeNalu

	// enhanced-rtmp packetType https://github.com/veovera/enhanced-rtmp
	RtmpExVideoPacketTypeSequenceStart uint8 = 0
	RtmpExVideoPacketTypeCodedFrames   uint8 = 1 // CompositionTime不为0时有这个类型
	RtmpExVideoPacketTypeSequenceEnd   uint8 = 2
	RtmpExVideoPacketTypeCodedFramesX  uint8 = 3

	// RtmpExFrameTypeKeyFrame RtmpExFrameTypeXXX...
	//
	// The following FrameType values are defined:
	// 0 = reserved
	// 1 = key frame (a seekable frame)
	// 2 = inter frame (a non-seekable frame)
	// ...
	RtmpExFrameTypeKeyFrame uint8 = 1

	RtmpAvcKeyFrame    = RtmpFrameTypeKey<<4 | RtmpCodecIdAvc
	RtmpHevcKeyFrame   = RtmpFrameTypeKey<<4 | RtmpCodecIdHevc
	RtmpAV1KeyFrame    = RtmpFrameTypeKey<<4 | RtmpCodecIdAV1
	RtmpAvcInterFrame  = RtmpFrameTypeInter<<4 | RtmpCodecIdAvc
	RtmpHevcInterFrame = RtmpFrameTypeInter<<4 | RtmpCodecIdHevc
	RtmpAV1InterFrame  = RtmpFrameTypeInter<<4 | RtmpCodecIdAV1

	// RtmpSoundFormatAac spec-video_file_format_spec_v10.pdf
	// Audio tags
	//   AUDIODATA
	//     SoundFormat UB[4]
	//     SoundRate   UB[2]
	//     SoundSize   UB[1]
	//     SoundType   UB[1]
	//   AACAUDIODATA
	//     AACPacketType UI8
	//     Data          UI8[n]
	// 注意，视频的CodecId是后4位，音频是前4位
	RtmpSoundFormatG711A uint8 = 7
	RtmpSoundFormatG711U uint8 = 8
	RtmpSoundFormatAac   uint8 = 10
	RtmpSoundFormatOpus  uint8 = 13

	// ext的codec自定义一个codecid
	RtmpSoundFormatEAC3 uint8 = 50

	RtmpAacPacketTypeSeqHeader = 0
	RtmpAacPacketTypeRaw       = 1

	RtmpExAudioPacketTypeSequenceStart      uint8 = 0
	RtmpExAudioPacketTypeCodedFrames        uint8 = 1 // CompositionTime不为0时有这个类型
	RtmpExAudioPacketTypeSequenceEnd        uint8 = 2
	RtmpExAudioPacketTypeMultichannelConfig uint8 = 4
	RtmpExAudioPacketTypeMultitrack         uint8 = 5
	RtmpExAudioPacketTypeModEx              uint8 = 7
)

type FourCC uint32

var (
	// video
	FourCCAV1  FourCC = 'a'<<24 | 'v'<<16 | '0'<<8 | '1'
	FourCCVP9  FourCC = 'v'<<24 | 'p'<<16 | '0'<<8 | '9'
	FourCCHEVC FourCC = 'h'<<24 | 'v'<<16 | 'c'<<8 | '1'
	FourCCAVC  FourCC = 'a'<<24 | 'v'<<16 | 'c'<<8 | '1'

	// audio
	FourCCOpus FourCC = 'O'<<24 | 'p'<<16 | 'u'<<8 | 's'
	FourCCAC3  FourCC = 'a'<<24 | 'c'<<16 | '-'<<8 | '3'
	FourCCEac3 FourCC = 'e'<<24 | 'c'<<16 | '-'<<8 | '3'
	FourCCAac  FourCC = 'm'<<24 | 'p'<<16 | '4'<<8 | 'a'
	FourCCMP3  FourCC = '.'<<24 | 'm'<<16 | 'p'<<8 | '3'
	FourCCFlac FourCC = 'f'<<24 | 'L'<<16 | 'a'<<8 | 'C'
)

type RtmpHeader struct {
	Csid         int
	MsgLen       uint32 // 不包含header的大小
	MsgTypeId    uint8  // 8 audio 9 video 18 metadata
	MsgStreamId  int
	TimestampAbs uint32 // dts, 经过计算得到的流上的绝对时间戳，单位毫秒
}

type RtmpMsg struct {
	Header       RtmpHeader
	Payload      []byte    // Payload不包含Header内容。如果需要将RtmpMsg序列化成RTMP chunk，可调用 rtmp.ChunkDivider 相关的函数
	CreateTime   time.Time // 帧创建时间
	Id           uint64    // 帧序号
	CodecChanged bool      // 编解码信息发生变化
}

func (msg RtmpMsg) Clone() (ret RtmpMsg) {
	ret.Header = msg.Header
	ret.Payload = make([]byte, len(msg.Payload))
	copy(ret.Payload, msg.Payload)
	return
}

func (msg RtmpMsg) IsVideo() bool {
	return msg.Header.MsgTypeId == RtmpTypeIdVideo
}

func (msg RtmpMsg) IsAudio() bool {
	return msg.Header.MsgTypeId == RtmpTypeIdAudio
}

func (msg RtmpMsg) Dts() uint32 {
	return msg.Header.TimestampAbs
}

// Pts
func (msg RtmpMsg) Pts() uint32 {
	if msg.Header.MsgTypeId == RtmpTypeIdAudio {
		return msg.Header.TimestampAbs
	}
	return msg.Header.TimestampAbs + BeUint24(msg.Payload[2:])
}

func (msg RtmpMsg) Cts() uint32 {
	if msg.Header.MsgTypeId == RtmpTypeIdAudio {
		return 0
	}

	isExtHeader := msg.Payload[0] & 0x80
	if isExtHeader != 0 {
		packetType := msg.Payload[0] & 0x0F
		switch packetType {
		case RtmpExVideoPacketTypeCodedFrames:
			return BeUint24(msg.Payload[5:])
		case RtmpExVideoPacketTypeCodedFramesX:
			return 0
		default:
			// Log.Warnf("RtmpMsg.Cts: packetType invalid, packetType=%d", packetType)
			return 0
		}
	}

	return BeUint24(msg.Payload[2:])
}

func (msg RtmpMsg) DebugString() string {
	isExtHeader := msg.Payload[0] & 0x80
	if msg.Header.MsgTypeId == RtmpTypeIdVideo && isExtHeader != 0 {
		frameType := msg.Payload[0] >> 4 & 0x07
		packetType := msg.Payload[0] & 0x0F // e.g. RtmpExVideoPacketTypeSequenceStart
		if isExtHeader != 0 {
			return fmt.Sprintf("type=%d,len=%d,dts=%d, ext(%d, %d, %d), payload=%s",
				msg.Header.MsgTypeId, msg.Header.MsgLen, msg.Header.TimestampAbs,
				isExtHeader, frameType, packetType,
				hex.Dump(Prefix(msg.Payload, 64)))
		}
	}

	return fmt.Sprintf("type=%d,len=%d,dts=%d, payload=%s",
		msg.Header.MsgTypeId, msg.Header.MsgLen, msg.Header.TimestampAbs, hex.Dump(Prefix(msg.Payload, 64)))
}

func (msg RtmpMsg) Type() string {
	switch msg.Header.MsgTypeId {
	case RtmpTypeIdVideo:
		return "video"
	case RtmpTypeIdAudio:
		return "audio"
	case RtmpTypeIdMetadata:
		return "metadata"
	}

	return "unknown"
}

//
// func (msg RtmpMsg) CodecIdString() string {
// 	name := "UNKNOWN"
// 	switch msg.Header.MsgTypeId {
// 	case RtmpTypeIdVideo:
// 		codecId := msg.VideoCodecId()
// 		if codecId == RtmpCodecIdAvc {
// 			name = VideoCodecAvc
// 		} else if codecId == RtmpCodecIdHevc {
// 			name = VideoCodecHevc
// 		} else if codecId == RtmpCodecIdAV1 {
// 			name = VideoCodecAV1
// 		}
// 	case RtmpTypeIdAudio:
// 		codecId := msg.AudioCodecId()
// 		if codecId == RtmpSoundFormatG711A {
// 			name = AudioCodecG711A
// 		} else if codecId == RtmpSoundFormatG711U {
// 			name = AudioCodecG711U
// 		} else if codecId == RtmpSoundFormatAac {
// 			name = AudioCodecAac
// 		} else if codecId == RtmpSoundFormatOpus {
// 			name = AudioCodecOpus
// 		}
// 	}
//
// 	return name
// }

func (msg RtmpMsg) SeqID() uint64 {
	return msg.Id
}

func (msg RtmpMsg) Size() int {
	return len(msg.Payload)
}

func (msg RtmpMsg) Dump16Bytes() string {
	return hex.Dump(Prefix(msg.Payload, 16))
}

func (msg RtmpMsg) Time() string {
	return msg.CreateTime.Format("2006-01-02 15:04:05.000")
}
