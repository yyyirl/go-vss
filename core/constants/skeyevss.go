// @Title        constants
// @Description  skeyevss
// @Create       yirl 2025/3/13 15:33

package constants

import "errors"

// ------------------------------------------------------ 启动参数

// web server 启动参数
const (
	SevWebParamWebStaticDir = "web-static-dir"
	SevWebParamCertPem      = "cert-pem"
	SevWebParamCertKey      = "cert-key"
)

// 服务启动参数
const (
	SevParamSevType          = "sev-type"
	SevParamEnv              = "env"
	SevParamDev              = "dev"
	SevParamConfig           = "f"
	SevParamActivateCodePath = "activate-code-path"
)

// ------------------------------------------------------ 文件信息

// 环境变量文件名
const (
	EnvFileNameDev   = ".env.local"
	EnvFileNameProd  = ".env.prod"
	EnvFileNameBuild = ".env.build"
	EnvFileNameOld   = ".env.prod.old"

	ActivateCodeFileName = ".activate.code"
)

const (
	VideoPlayAddressTypeRtsp     = "RTSP"
	VideoPlayAddressTypeRtmp     = "RTMP"
	VideoPlayAddressTypeHttpFlv  = "HTTP-FLV"
	VideoPlayAddressTypeWsFlv    = "WS-FLV"
	VideoPlayAddressTypeHls      = "HLS"
	VideoPlayAddressTypeHttpFmp4 = "HTTP-FMP4"
	VideoPlayAddressTypeWebrtc   = "WEBRTC"
	VideoPlayAddressTypeWebrtcDc = "webrtcDC"
)

var (
	VideoPlayAddressTypes = []string{
		VideoPlayAddressTypeRtsp,
		VideoPlayAddressTypeRtmp,
		VideoPlayAddressTypeHttpFlv,
		VideoPlayAddressTypeWsFlv,
		VideoPlayAddressTypeHls,
		VideoPlayAddressTypeHttpFmp4,
		VideoPlayAddressTypeWebrtc,
		VideoPlayAddressTypeWebrtcDc,
	}
)

var DeviceUnregistered = errors.New("设备未注册")
