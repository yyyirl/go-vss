package permissions

import "skeyevss/core/pkg/functions"

type PType uint

type IdType string

type ActionType string

const (
	ActionCreateType ActionType = "create"
	ActionDeleteType ActionType = "delete"
	ActionUpdateType ActionType = "update"
)

type Item struct {
	UniqueId   IdType     `json:"uniqueId"`
	Name       string     `json:"name"`
	Super      bool       `json:"super"`
	Universal  bool       `json:"universal"` // 不验证
	ActionType ActionType `json:"type"`
	Level      int        `json:"level"`

	Children []*Item `json:"children"`
}

func init() {
	initVerifyData(append([]*Item{}, backend, frontend), -1)
}

// 设备操作员
func EquipmentOperator() string {
	var data []IdType
	for _, item := range backend.Children {
		if item.UniqueId == P_0_3 {
			// 设备
			data = append(data, item.UniqueId)
			for _, v := range item.Children {
				data = append(data, v.UniqueId)
				for _, v1 := range v.Children {
					data = append(data, v1.UniqueId)
				}
			}
		} else if item.UniqueId == P_0_4 {
			// 录像
			data = append(data, item.UniqueId)
			for _, v := range item.Children {
				data = append(data, v.UniqueId)
				for _, v1 := range v.Children {
					data = append(data, v1.UniqueId)
				}
			}
		} else if item.UniqueId == P_0_6 {
			// 日志
			data = append(data, item.UniqueId)
			for _, v := range item.Children {
				if v.UniqueId == P_0_6_4 {
					// 报警日志
					data = append(data, v.UniqueId)
					for _, v1 := range v.Children {
						data = append(data, v1.UniqueId)
					}
				}
			}
		} else if item.UniqueId == P_0_7 {
			// 内部调用
			data = append(data, item.UniqueId)
			for _, v := range item.Children {
				data = append(data, v.UniqueId)
				for _, v1 := range v.Children {
					data = append(data, v1.UniqueId)
				}
			}
		}
	}

	for _, item := range frontend.Children {
		if item.UniqueId == P_1_0 {
			// 首页概览
			data = append(data, item.UniqueId)
		}
		if item.UniqueId == P_1_4 {
			// 视频调阅
			data = append(data, item.UniqueId)
			for _, v := range item.Children {
				data = append(data, v.UniqueId)
				for _, v1 := range v.Children {
					data = append(data, v1.UniqueId)
				}
			}
		} else if item.UniqueId == P_1_3 || item.UniqueId == P_1_5 {
			// 设备 录像
			data = append(data, item.UniqueId)
			for _, v := range item.Children {
				data = append(data, v.UniqueId)
				for _, v1 := range v.Children {
					data = append(data, v1.UniqueId)
				}
			}
		} else if item.UniqueId == P_1_6 {
			data = append(data, item.UniqueId)
			for _, v := range item.Children {
				if v.UniqueId == P_1_6_4 {
					// 报警管理
					data = append(data, v.UniqueId)
					for _, v1 := range v.Children {
						data = append(data, v1.UniqueId)
					}
				}
			}
		}
	}

	v, _ := functions.ToString(data)
	return v
}

func Source() map[string][]*Item {
	return map[string][]*Item{
		"frontend": frontend.Children,
		"backend":  backend.Children,
	}
}
