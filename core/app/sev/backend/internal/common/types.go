// @Title        common
// @Description  types
// @Create       yirl 2025/3/26 13:11

package common

import (
	"skeyevss/core/common/source/permissions"
	"skeyevss/core/repositories/models/settings"
	systemOperationLogs "skeyevss/core/repositories/models/system-operation-logs"
	"skeyevss/core/tps"
)

type Setting struct {
	// 超级管理员
	Showcase bool `json:"showcase"`
	Super    uint `json:"super"`
	// 操作日志类型
	SystemOperationLogTypes map[systemOperationLogs.Type]string `json:"system-operation-log-types"`
	// 前端权限列表
	Permissions map[string][]*permissions.Item `json:"permissions"`
	// 当前登录用户的权限信息
	PermissionIds []string `json:"permissionIds"`
	// 内网ip
	InternalIP string `json:"internal-ip"`
	// 系统设置
	Setting *settings.Content `json:"setting"`
	// 流媒体传输模式
	MediaTransModes map[uint]string `json:"media-trans-modes"`
	// 接入协议
	AccessProtocols map[uint]string `json:"access-protocols"`
	// 接入协议 颜色区分
	AccessProtocolColors map[uint]string `json:"access-protocol-colors"`
	// 通道过滤
	ChannelFilters map[string]string `json:"channel-filters"`
	// 码流索引
	BitstreamIndexes map[uint]string `json:"bitstream-indexes"`
	// 摄像机云台类型
	PTXTypes map[uint]string `json:"ptz-types"`
	// vss http url
	VssHttpUrl string `json:"vssHttpUrl"`
	// vss sse url
	VssSseUrl string `json:"vssSseUrl"`
	// media server url
	MSUrl string `json:"msUrl"`
	// websocket
	WSUrl string `json:"wsUrl"`
	// rtmp port
	RtmpPort int `json:"rtmpPort"`
	// 文件代理url
	ProxyFileUrl string `json:"proxy-file-url"`
	// 播放地址类型
	MSVideoPlayAddressTypes []string `json:"ms-video-play-address-types"`
	// pprof列表
	PProf        []tps.OptionItem `json:"pprof"`
	PProfFileDir string           `json:"pprof-file-dir"`
	// 报警类型
	AlarmTypes map[uint]string `json:"alarm-types"`
	// 报警类型扩展参数
	EventTypes map[uint]string `json:"event-types"`
	// 报警方式
	AlarmMethods map[uint]string `json:"alarm-methods"`
	// 报警级别
	AlarmPriorities map[uint]string `json:"alarm-priorities"`
	// 信令传输协议
	CascadeSipProtocols map[uint]string `json:"cascade-sip-protocols"`
	// sip默认服务器端口(上级)
	SipPort int `json:"sip-port"`
	// 下级默认sip服务器端口(上级)
	CascadeSipPort int `json:"cascade-sip-port"`
	// 接口文档目录
	ApiDocDir string `json:"api-doc-dir"`
	// referer
	Referer string `json:"referer"`
	// 构建时间
	BuildTime string `json:"bt"`
	// 天地图key
	TMapKey string `json:"tmap-key"`
}
