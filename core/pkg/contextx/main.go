package contextx

import (
	"context"
	"strings"

	"skeyevss/core/constants"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/tps"
)

func GetLanguage(ctx context.Context) string {
	return getCtxString(ctx, constants.HEADER_LANG)
}

func GetNewToken(ctx context.Context) string {
	return getCtxString(ctx, constants.HEADER_NEW_TOKEN)
}

func GetCtxIP(ctx context.Context) string {
	return getCtxString(ctx, constants.HEADER_IP)
}

func IsHttps(ctx context.Context) bool {
	return strings.Index(getCtxString(ctx, constants.HEADER_TLS), "https") == 0
}

func GetCtxMAC(ctx context.Context) string {
	return getCtxString(ctx, constants.HEADER_MAC)
}

func getCtxString(ctx context.Context, key string) string {
	var data = ctx.Value(key)
	if data == nil {
		return ""
	}

	if v, ok := data.(interface{}); ok {
		return v.(string)
	}

	return ""
}

func getCtxBool(ctx context.Context, key string) bool {
	var data = ctx.Value(key)
	if data == nil {
		return false
	}

	if v, ok := data.(interface{}); ok {
		return v.(bool)
	}

	return false
}

func GetSuperState(ctx context.Context) uint {
	v, err := functions.InterfaceToNumber[uint](ctx.Value(constants.CTX_SUPER_STATE))
	if err != nil {
		return 0
	}

	return v
}

func GetDepartmentIds(ctx context.Context) []uint64 {
	var data = ctx.Value(constants.CTX_DEP_IDS)
	if data == nil {
		return nil
	}

	ids, err := functions.SliceInterfaceToNumber[uint64](data)
	if err != nil {
		return nil
	}

	return ids
}

func GetPermissionIds(ctx context.Context) []string {
	var data = ctx.Value(constants.CTX_PERMISSION_IDS)
	if data == nil {
		return nil
	}

	var ids []string
	if v, ok := data.([]string); ok {
		for _, item := range v {
			ids = append(ids, item)
		}
	}

	if len(ids) <= 0 {
		return []string{}
	}

	return ids
}

func GetLogoutState(ctx context.Context) bool {
	return getCtxBool(ctx, constants.HEADER_IS_LOGOUT)
}

func GetShowcaseState(ctx context.Context) bool {
	return getCtxBool(ctx, constants.CTX_SHOWCASE)
}

func GetIsInternalReq(ctx context.Context) bool {
	return getCtxBool(ctx, constants.CTX_VSS_IS_INTERNAL_REQ)
}

func GetResetPwdState(ctx context.Context) bool {
	return getCtxBool(ctx, constants.CTX_NEEDED_RESET_PWD)
}

func GetCtxReqStartTime(ctx context.Context) int64 {
	v, err := functions.InterfaceToNumber[int64](ctx.Value(constants.CTX_REQ_START_TIME))
	if err != nil {
		return 0
	}

	return v
}

// 获取当前登录用户id
func GetCtxUserid(ctx context.Context) uint {
	v, err := functions.InterfaceToNumber[uint](ctx.Value(constants.CTX_USERID))
	if err != nil {
		return 0
	}

	return v
}

func GetCtxPlatform(ctx context.Context) uint {
	v, err := functions.InterfaceToNumber[uint](ctx.Value(constants.HEADER_PLATFORM))
	if err != nil {
		return 0
	}

	return v
}

// 获取当前登录用户信息
func GetCtxUserinfo(ctx context.Context) *tps.TokenItem {
	var data = ctx.Value(constants.CTX_TOKEN_INFO)
	if data == nil {
		return nil
	}

	if v, ok := data.(*tps.TokenItem); ok {
		return v
	}
	return nil
}

// 获取请求信息
func GetCtxRequestInfo(ctx context.Context) map[string]interface{} {
	var data = ctx.Value(constants.CTX_REQUESTS)
	if data == nil {
		return nil
	}

	if v, ok := data.(map[string]interface{}); ok {
		return v
	}
	return nil
}
