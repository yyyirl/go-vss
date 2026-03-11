// Copyright 2021, Chef.  All rights reserved.
// https://gitee.com/openskeye-lab/skeyesms
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package base

import (
	"errors"
	"fmt"
)

// ----- 通用的 ---------------------------------------------------------------------------------------------------------

var (
	ErrShortBuffer  = errors.New("svr: buffer too short")
	ErrFileNotExist = errors.New("svr: file not exist")
)

// ----- pkg/aac -------------------------------------------------------------------------------------------------------

var ErrSamplingFrequencyIndex = errors.New("svr.aac: invalid sampling frequency index")

// ----- pkg/aac -------------------------------------------------------------------------------------------------------

var ErrAvc = errors.New("svr.avc: error")

// ----- pkg/base ------------------------------------------------------------------------------------------------------

var (
	ErrAddrEmpty               = errors.New("svr.base: http server addr empty")
	ErrMultiRegisterForPattern = errors.New("svr.base: http server multiple registrations for pattern")

	ErrSessionNotStarted = errors.New("svr.base: session has not been started yet")

	ErrInvalidUrl = errors.New("svr.base: invalid url")
)

// ----- pkg/hevc ------------------------------------------------------------------------------------------------------

var ErrHevc = errors.New("svr.hevc: error")

// ----- pkg/hls -------------------------------------------------------------------------------------------------------

var (
	ErrHls                = errors.New("svr.hls: error")
	ErrHlsSessionNotFound = errors.New("svr.hls: hls session not found")
)

// ----- pkg/rtmp ------------------------------------------------------------------------------------------------------

var (
	ErrAmfInvalidType = errors.New("svr.rtmp: invalid amf0 type")
	ErrAmfTooShort    = errors.New("svr.rtmp: too short to unmarshal amf0 data")
	ErrAmfNotExist    = errors.New("svr.rtmp: not exist")

	ErrRtmpShortBuffer   = errors.New("svr.rtmp: buffer too short")
	ErrRtmpUnexpectedMsg = errors.New("svr.rtmp: unexpected msg")
)

func NewErrAmfInvalidType(b byte) error {
	return fmt.Errorf("%w. b=%d", ErrAmfInvalidType, b)
}

func NewErrRtmpShortBuffer(need, actual int, msg string) error {
	return fmt.Errorf("%w. need=%d, actual=%d, msg=%s", ErrRtmpShortBuffer, need, actual, msg)
}

// ----- pkg/rtprtcp ---------------------------------------------------------------------------------------------------

var ErrRtpRtcpShortBuffer = errors.New("svr.rtprtcp: buffer too short")

// ----- pkg/rtsp ------------------------------------------------------------------------------------------------------

var (
	ErrRtsp                     = errors.New("svr.rtsp: error")
	ErrRtspClosedByObserver     = errors.New("svr.rtsp: close by observer")
	ErrRtspUnsupportedTransport = errors.New("svr.rtsp: unsupported Transport")
)

// ----- pkg/sdp -------------------------------------------------------------------------------------------------------

var ErrSdp = errors.New("svr.sdp: error")

// ----- pkg/logic -----------------------------------------------------------------------------------------------------

var (
	ErrDupInStream      = errors.New("svr.logic: in stream already exist at group")
	ErrDisposedInStream = errors.New("svr.logic: in stream already disposed")

	ErrSimpleAuthParamNotFound = errors.New("svr.logic: simple auth failed since url param svr_secret not found")
	ErrSimpleAuthFailed        = errors.New("svr.logic: simple auth failed since url param svr_secret invalid")
)

// ---------------------------------------------------------------------------------------------------------------------
