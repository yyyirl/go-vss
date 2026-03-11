/**
 * @Author:         yi
 * @Description:    main
 * @Version:        1.0.0
 * @Date:           2023/6/28 23:48
 */
package pubsub

import "skeyevss/core/tps"

// 消息列表
type redisMessages = []string

// 消息列表最大容量
const maxMessageCount = 5000

// 心跳检测清空数据周期
const heartbeatInterval = 500 // 毫秒

// 没有消息进入是最后一次发送时间间隔
const sendInterval = 500 // 毫秒

// configuration
type Conf struct {
	Email tps.YamlEmail
	// 消息列表最大容量
	MaxMessageCount,
	// 心跳检测清空数据周期
	HeartbeatInterval,
	// 没有消息进入是最后一次发送时间间隔
	SendInterval int
	// 当前节点域名
	Host string
}
