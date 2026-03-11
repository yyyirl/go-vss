package systemOperationLogs

import (
	"errors"

	"github.com/mitchellh/mapstructure"

	"skeyevss/core/pkg/functions"
)

type Item struct {
	*SystemOperationLogs

	Username   string `json:"username"`
	UseDBCache bool   `json:"-"`
}

func NewItem() *Item {
	return new(Item)
}

func (i *Item) ConvToModel(call func(*Item) *Item) (*SystemOperationLogs, error) {
	if i.SystemOperationLogs == nil {
		return nil, nil
	}

	if call != nil {
		i = call(i)
	}

	return i.SystemOperationLogs, nil
}

// map转struct
func (i *Item) MapToModel(input map[string]interface{}) (*Item, error) {
	if input == nil {
		return nil, errors.New("input object is nil")
	}

	var model SystemOperationLogs
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			DecodeHook: mapstructure.DecodeHookFunc(functions.MapStructureHook),
			Result:     &model,
			// TagName:    "mapstructure",
		},
	)
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(input); err != nil {
		return nil, err
	}

	return &Item{SystemOperationLogs: &model}, nil
}

func (*Item) CheckMap(input map[string]interface{}) (map[string]interface{}, error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}

	for column := range input {
		if !functions.Contains(column, Columns) {
			return nil, errors.New("column: " + column + " does not exist")
		}
	}

	return input, nil
}

type Type = uint

const (
	_                    = iota
	TypeLogin            // 后台登录
	TypePermissionCreate // 权限创建
	TypePermissionUpdate // 权限更新
	TypePermissionDelete // 权限删除
	TypeRoleCreate       // 角色创建
	TypeRoleDelete       // 角色删除
	TypeRoleUpldate      // 角色修改
	TypeDepartmentCreate // 组织部门创建
	TypeDepartmentDelete // 组织部门删除
	TypeDepartmentUpdate // 组织部门修改
	TypeAdminCreate      // 管理员创建
	TypeAdminDelete      // 管理员删除
	TypeAdminUpdate      // 管理员修改

	TypeDictCreate               // 字典创建
	TypeDictDelete               // 字典删除
	TypeDictUpdate               // 字典修改
	TypeAdminPasswordUpdate      // 管理员密码修改
	TypeInitializePasswordUpdate // 初始化密码
	TypeCrontabDictCreate        // 任务创建
	TypeCrontabDictDelete        // 任务删除
	TypeCrontabDictUpdate        // 任务修改
	TypeSettingUpdate            // 设置修改

	TypeDeviceItemCreate // 设备创建
	TypeDeviceItemDelete // 设备删除
	TypeDeviceItemUpdate // 设备修改

	TypeDeviceChannelCreate // 通道创建
	TypeDeviceChannelDelete // 通道删除
	TypeDeviceChannelUpdate // 通道修改

	TypeMSCreate // media server创建
	TypeMSDelete // media server删除
	TypeMSUpdate // media server修改

	TypeVideoSKDelete // 平台视频删除
	TypeVideoSKUpdate // 平台视频修改

	TypeAlarmDelete // 报警删除

	TypeVPCreate // video project创建
	TypeVPDelete // video project删除
	TypeVPUpdate // video project修改

	TypeDeviceCascadeCreate // 平台级联创建
	TypeDeviceCascadeDelete // 平台级联删除
	TypeDeviceCascadeUpdate // 平台级联修改

	baseMax = 5000
	Known   = 10000
)

const (
	_ = iota + baseMax
)

var (
	Types = functions.MergeMaps(
		map[Type]Type{
			TypeLogin:               TypeLogin,
			TypeDepartmentCreate:    TypeDepartmentCreate,
			TypeDepartmentDelete:    TypeDepartmentDelete,
			TypeDepartmentUpdate:    TypeDepartmentUpdate,
			TypePermissionCreate:    TypePermissionCreate,
			TypePermissionUpdate:    TypePermissionUpdate,
			TypePermissionDelete:    TypePermissionDelete,
			TypeRoleCreate:          TypeRoleCreate,
			TypeRoleDelete:          TypeRoleDelete,
			TypeRoleUpldate:         TypeRoleUpldate,
			TypeAdminCreate:         TypeAdminCreate,
			TypeAdminDelete:         TypeAdminDelete,
			TypeAdminUpdate:         TypeAdminUpdate,
			TypeDictCreate:          TypeDictCreate,
			TypeDictDelete:          TypeDictDelete,
			TypeDictUpdate:          TypeDictUpdate,
			TypeAdminPasswordUpdate: TypeAdminPasswordUpdate,
			TypeSettingUpdate:       TypeSettingUpdate,

			TypeCrontabDictCreate: TypeCrontabDictCreate,
			TypeCrontabDictDelete: TypeCrontabDictDelete,
			TypeCrontabDictUpdate: TypeCrontabDictUpdate,

			TypeDeviceItemCreate: TypeDeviceItemCreate,
			TypeDeviceItemDelete: TypeDeviceItemDelete,
			TypeDeviceItemUpdate: TypeDeviceItemUpdate,

			TypeDeviceChannelCreate: TypeDeviceChannelCreate,
			TypeDeviceChannelDelete: TypeDeviceChannelDelete,
			TypeDeviceChannelUpdate: TypeDeviceChannelUpdate,

			TypeMSCreate: TypeMSCreate,
			TypeMSDelete: TypeMSDelete,
			TypeMSUpdate: TypeMSUpdate,

			TypeVPCreate: TypeVPCreate,
			TypeVPDelete: TypeVPDelete,
			TypeVPUpdate: TypeVPUpdate,

			TypeVideoSKDelete: TypeVideoSKDelete,
			TypeVideoSKUpdate: TypeVideoSKUpdate,
			TypeAlarmDelete:   TypeAlarmDelete,

			TypeDeviceCascadeCreate: TypeDeviceCascadeCreate,
			TypeDeviceCascadeDelete: TypeDeviceCascadeDelete,
			TypeDeviceCascadeUpdate: TypeDeviceCascadeUpdate,

			Known: Known,
		},
		// extension
		map[Type]Type{},
	)

	TypeViews = functions.MergeMaps(
		map[Type]string{
			TypeLogin:               "后台登录",
			TypePermissionCreate:    "权限创建",
			TypePermissionUpdate:    "权限更新",
			TypePermissionDelete:    "权限删除",
			TypeRoleCreate:          "角色创建",
			TypeRoleDelete:          "角色删除",
			TypeRoleUpldate:         "角色修改",
			TypeAdminCreate:         "管理员创建",
			TypeAdminDelete:         "管理员删除",
			TypeAdminUpdate:         "管理员修改",
			TypeDictCreate:          "字典创建",
			TypeDictDelete:          "字典删除",
			TypeDictUpdate:          "字典修改",
			TypeCrontabDictCreate:   "任务创建",
			TypeCrontabDictDelete:   "任务删除",
			TypeCrontabDictUpdate:   "任务修改",
			TypeDepartmentCreate:    "组织部门创建",
			TypeDepartmentDelete:    "组织部门删除",
			TypeDepartmentUpdate:    "组织部门修改",
			TypeAdminPasswordUpdate: "密码更新",
			TypeSettingUpdate:       "设置修改",
			TypeDeviceItemCreate:    "设备创建",
			TypeDeviceItemDelete:    "设备更新",
			TypeDeviceItemUpdate:    "设备删除",
			TypeDeviceChannelCreate: "通道创建",
			TypeDeviceChannelDelete: "通道更新",
			TypeDeviceChannelUpdate: "通道删除",
			TypeMSCreate:            "media server创建",
			TypeMSDelete:            "media server删除",
			TypeMSUpdate:            "media server修改",
			TypeVPCreate:            "video project创建",
			TypeVPDelete:            "video project删除",
			TypeVPUpdate:            "video project修改",
			TypeVideoSKDelete:       "平台视频删除",
			TypeVideoSKUpdate:       "平台视频修改",
			TypeAlarmDelete:         "报警删除",
			TypeDeviceCascadeCreate: "平台级联创建",
			TypeDeviceCascadeDelete: "平台级联删除",
			TypeDeviceCascadeUpdate: "平台级联修改",

			Known: "未知类型",
		},
		// extension
		map[Type]string{},
	)
)
