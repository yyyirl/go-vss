/**
 * @Author:         yi
 * @Description:    user
 * @Version:        1.0.0
 * @Date:           2024/10/31 16:25
 */
package constants

const (
	USERS_STATUS_DEF uint = iota
	USERS_STATUS_FREEZE
)

const (
	USER_STATE_DEF          uint = iota // 默认未订阅
	USER_STATE_PERMANENT                // 永久
	USER_STATE_SUBSCRIPTION             // 按时订阅
)
