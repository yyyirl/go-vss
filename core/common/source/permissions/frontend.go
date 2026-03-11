package permissions

const (
	P_1       IdType = "P_1"
	P_1_0     IdType = "P_1_0"
	P_1_1     IdType = "P_1_1"
	P_1_1_1   IdType = "P_1_1_1"
	P_1_1_1_2 IdType = "P_1_1_1_2"
	P_1_1_1_3 IdType = "P_1_1_1_3"
	P_1_1_1_4 IdType = "P_1_1_1_4"
	P_1_1_1_5 IdType = "P_1_1_1_5"
	P_1_1_1_6 IdType = "P_1_1_1_6"
	P_1_1_2   IdType = "P_1_1_2"
	P_1_1_2_1 IdType = "P_1_1_2_1"
	P_1_1_2_2 IdType = "P_1_1_2_2"
	P_1_1_3   IdType = "P_1_1_3"
	P_1_1_3_1 IdType = "P_1_1_3_1"
	P_1_1_3_2 IdType = "P_1_1_3_2"
	P_1_1_3_3 IdType = "P_1_1_3_3"
	P_1_1_4   IdType = "P_1_1_4"
	P_1_1_4_1 IdType = "P_1_1_4_1"
	P_1_1_4_2 IdType = "P_1_1_4_2"
	P_1_1_4_3 IdType = "P_1_1_4_3"
	P_1_6     IdType = "P_1_6"
	P_1_6_1   IdType = "P_1_6_1"
	P_1_6_2   IdType = "P_1_6_2"
	P_1_6_3   IdType = "P_1_6_3"
	P_1_6_4   IdType = "P_1_6_4"
	P_1_6_4_1 IdType = "P_1_6_4_1"
	P_1_6_4_2 IdType = "P_1_6_4_2"
	P_1_6_5   IdType = "P_1_6_5"
	P_1_6_6   IdType = "P_1_6_6"
	P_1_6_7   IdType = "P_1_6_7"
	P_1_6_8   IdType = "P_1_6_8"
	P_1_1_6   IdType = "P_1_1_6"
	P_1_2     IdType = "P_1_2"
	P_1_2_1   IdType = "P_1_2_1"
	P_1_2_1_1 IdType = "P_1_2_1_1"
	P_1_2_1_2 IdType = "P_1_2_1_2"
	P_1_2_1_3 IdType = "P_1_2_1_3"
	P_1_2_2   IdType = "P_1_2_2"
	P_1_2_2_1 IdType = "P_1_2_2_1"
	P_1_2_2_2 IdType = "P_1_2_2_2"
	P_1_2_2_3 IdType = "P_1_2_2_3"
	P_1_2_3   IdType = "P_1_2_3"
	P_1_2_3_1 IdType = "P_1_2_3_1"
	P_1_4_3   IdType = "P_1_4_3"
	P_1_4_3_1 IdType = "P_1_4_3_1"
	P_1_4_3_2 IdType = "P_1_4_3_2"
	P_1_4_3_3 IdType = "P_1_4_3_3"
	P_1_4_4   IdType = "P_1_4_4"
	P_1_4_4_1 IdType = "P_1_4_4_1"
	P_1_4_4_2 IdType = "P_1_4_4_2"
	P_1_4_4_3 IdType = "P_1_4_4_3"
	P_1_4_5   IdType = "P_1_4_5"
	P_1_4_5_1 IdType = "P_1_4_5_1"
	P_1_4_6   IdType = "P_1_4_6"
	P_1_7     IdType = "P_1_7"

	P_1_3     IdType = "P_1_3"
	P_1_3_1   IdType = "P_1_3_1"
	P_1_3_1_1 IdType = "P_1_3_1_1"
	P_1_3_1_2 IdType = "P_1_3_1_2"
	P_1_3_1_3 IdType = "P_1_3_1_3"
	P_1_3_2   IdType = "P_1_3_2"
	P_1_3_2_1 IdType = "P_1_3_2_1"
	P_1_3_2_2 IdType = "P_1_3_2_2"
	P_1_3_2_3 IdType = "P_1_3_2_3"
	P_1_3_3   IdType = "P_1_3_3"
	P_1_3_3_1 IdType = "P_1_3_3_1"
	P_1_3_3_2 IdType = "P_1_3_3_2"
	P_1_3_3_3 IdType = "P_1_3_3_3"

	P_1_4     IdType = "P_1_4"
	P_1_4_1   IdType = "P_1_4_1"
	P_1_4_2   IdType = "P_1_4_2"
	P_1_4_2_1 IdType = "P_1_4_2_1"

	P_1_5     IdType = "P_1_5"
	P_1_5_1   IdType = "P_1_5_1"
	P_1_5_1_1 IdType = "P_1_5_1_1"
	P_1_5_2   IdType = "P_1_5_2"
	P_1_5_2_3 IdType = "P_1_5_2_3"
	P_1_5_2_2 IdType = "P_1_5_2_2"
)

var frontend = &Item{
	UniqueId: P_1,
	Name:     "前台权限",
	Children: []*Item{
		{
			UniqueId:  P_1_0,
			Name:      "首页概览",
			Universal: true,
		},
		{
			UniqueId: P_1_1,
			Name:     "system",
			Children: []*Item{
				{
					UniqueId: P_1_1_1,
					Name:     "系统设置",
					Children: []*Item{
						{
							UniqueId: P_1_1_1_2,
							Name:     "重启服务",
						},
						{
							UniqueId: P_1_1_1_3,
							Name:     "服务升级",
						},
						{
							UniqueId: P_1_1_1_4,
							Name:     "检查更新",
						},
						{
							UniqueId: P_1_1_1_5,
							Name:     "激活查询",
						},
						{
							UniqueId: P_1_1_1_6,
							Name:     "提交激活文件",
						},
					},
				},
				{
					UniqueId: P_1_1_2,
					Name:     "角色管理",
					Children: []*Item{
						{
							UniqueId:   P_1_1_2_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId:   P_1_1_2_2,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_1_3,
					Name:     "部门管理",
					Children: []*Item{
						{
							UniqueId:   P_1_1_3_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId:   P_1_1_3_2,
							Name:       "更新",
							ActionType: ActionUpdateType,
						},
						{
							UniqueId:   P_1_1_3_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_1_4,
					Name:     "管理员管理",
					Children: []*Item{
						{
							UniqueId:   P_1_1_4_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId:   P_1_1_4_2,
							Name:       "更新",
							ActionType: ActionUpdateType,
						},
						{
							UniqueId:   P_1_1_4_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_1_6,
					Name:     "系统信息",
				},
			},
		},
		{
			UniqueId: P_1_2,
			Name:     "配置中心",
			Children: []*Item{
				{
					UniqueId: P_1_2_1,
					Name:     "字典",
					Children: []*Item{
						{
							UniqueId:   P_1_2_1_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId: P_1_2_1_2,
							Name:     "更新",
						},
						{
							UniqueId:   P_1_2_1_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_2_2,
					Name:     "Crontab",
					Children: []*Item{
						{
							UniqueId:   P_1_2_2_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId:   P_1_2_2_2,
							Name:       "更新",
							ActionType: ActionUpdateType,
						},
						{
							UniqueId:   P_1_2_2_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_2_3,
					Name:     "系统设置",
					Children: []*Item{
						{
							UniqueId:   P_1_2_3_1,
							Name:       "更新",
							ActionType: ActionUpdateType,
						},
					},
				},
				{
					UniqueId: P_1_4_3,
					Name:     "媒体服务",
					Children: []*Item{
						{
							UniqueId:   P_1_4_3_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId:   P_1_4_3_2,
							Name:       "更新",
							ActionType: ActionUpdateType,
						},
						{
							UniqueId:   P_1_4_3_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_4_4,
					Name:     "录像计划",
					Children: []*Item{
						{
							UniqueId:   P_1_4_4_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId:   P_1_4_4_2,
							Name:       "更新",
							ActionType: ActionUpdateType,
						},
						{
							UniqueId:   P_1_4_4_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_4_5,
					Name:     "服务器配置",
					Children: []*Item{
						{
							UniqueId:   P_1_4_5_1,
							Name:       "更新",
							ActionType: ActionUpdateType,
						},
					},
				},
				{
					UniqueId: P_1_4_6,
					Name:     "接口文档",
				},
			},
		},
		{
			UniqueId: P_1_3,
			Name:     "设备管理",
			Children: []*Item{
				{
					UniqueId: P_1_3_1,
					Name:     "设备",
					Children: []*Item{
						{
							UniqueId:   P_1_3_1_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId: P_1_3_1_2,
							Name:     "更新",
						},
						{
							UniqueId:   P_1_3_1_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_3_2,
					Name:     "设备",
					Children: []*Item{
						{
							UniqueId:   P_1_3_2_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId: P_1_3_2_2,
							Name:     "更新",
						},
						{
							UniqueId:   P_1_3_2_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_3_3,
					Name:     "平台级联",
					Children: []*Item{
						{
							UniqueId:   P_1_3_3_1,
							Name:       "创建",
							ActionType: ActionCreateType,
						},
						{
							UniqueId: P_1_3_3_2,
							Name:     "更新",
						},
						{
							UniqueId:   P_1_3_3_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
					},
				},
			},
		},
		{
			UniqueId: P_1_4,
			Name:     "视频调阅",
			Children: []*Item{
				{
					UniqueId: P_1_4_1,
					Name:     "预览",
				},
				{
					UniqueId: P_1_4_2,
					Name:     "组织架构",
				},
				{
					UniqueId: P_1_4_2_1,
					Name:     "通道归属设置",
				},
			},
		},
		{
			UniqueId: P_1_5,
			Name:     "录像管理",
			Children: []*Item{
				{
					UniqueId: P_1_5_1,
					Name:     "设备录像",
					Children: []*Item{
						{
							UniqueId:   P_1_5_1_1,
							Name:       "视频下载",
							ActionType: ActionDeleteType,
						},
					},
				},
				{
					UniqueId: P_1_5_2,
					Name:     "平台录像",
					Children: []*Item{
						{
							UniqueId:   P_1_5_2_3,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
						{
							UniqueId: P_1_5_2_2,
							Name:     "更新",
						},
					},
				},
			},
		},
		{
			UniqueId: P_1_6,
			Name:     "日志",
			Children: []*Item{
				{
					UniqueId: P_1_6_3,
					Name:     "操作日志",
				},
				{
					UniqueId: P_1_6_2,
					Name:     "运行日志",
				},
				{
					UniqueId: P_1_6_1,
					Name:     "查看日志",
				},
				{
					UniqueId: P_1_6_5,
					Name:     "SIP日志",
				},
				{
					UniqueId: P_1_6_6,
					Name:     "性能分析",
				},
				{
					UniqueId: P_1_6_7,
					Name:     "响应查询",
				},
				{
					UniqueId: P_1_6_4,
					Name:     "报警管理",
					Children: []*Item{
						{
							UniqueId:   P_1_6_4_1,
							Name:       "删除",
							ActionType: ActionDeleteType,
						},
						{
							UniqueId: P_1_6_4_2,
							Name:     "列表",
						},
					},
				},
				{
					UniqueId: P_1_6_8,
					Name:     "VSS Sev State",
				},
			},
		},
		{
			UniqueId: P_1_7,
			Name:     "电子地图",
		},
	},
}
