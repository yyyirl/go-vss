package localization

type Item struct {
	ZH string `json:"zh"`
	EN string `json:"en"`
}

var (
	M0001 = &Item{
		ZH: "参数错误",
		EN: "参数错误",
	}
	M0002 = &Item{
		ZH: "生成失败",
		EN: "生成失败",
	}
	M0003 = &Item{
		ZH: "验证失败",
		EN: "验证失败",
	}
	M0004 = &Item{
		ZH: "系统参数`aes key`错误",
		EN: "系统参数`aes key`错误",
	}
	M0005 = &Item{
		ZH: "系统参数`expire`错误",
		EN: "系统参数`expire`错误",
	}
	M0006 = &Item{
		ZH: "登录超时",
		EN: "登录超时",
	}
	M0007 = &Item{
		ZH: "服务器错误",
		EN: "服务器错误",
	}
	M0008 = &Item{
		ZH: "用户id获取失败",
		EN: "用户id获取失败",
	}
	M0009 = &Item{
		ZH: "系统数据转换失败",
		EN: "系统数据转换失败",
	}
	M0010 = &Item{
		ZH: "获取失败",
		EN: "获取失败",
	}
	M0011 = &Item{
		ZH: "未分配部门",
		EN: "未分配部门",
	}
	M0012 = &Item{
		ZH: "未授权",
		EN: "未授权",
	}

	M0026 = &Item{
		ZH: "数据解析失败",
		EN: "数据解析失败",
	}

	M00271 = &Item{
		ZH: "操作失败",
		EN: "操作失败",
	}

	M00272 = &Item{
		ZH: "临时文件创建失败",
		EN: "临时文件创建失败",
	}

	M00273 = &Item{
		ZH: "邮件发送失败",
		EN: "邮件发送失败",
	}

	M00274 = &Item{
		ZH: "其他客户端正在使用",
		EN: "其他客户端正在使用",
	}

	M00275 = &Item{
		ZH: "脚本创建失败",
		EN: "脚本创建失败",
	}

	M00276 = &Item{
		ZH: "执行失败",
		EN: "执行失败",
	}

	M00277 = &Item{
		ZH: "文件不存在",
		EN: "文件不存在",
	}

	M00278 = &Item{
		ZH: "启动脚本下载失败",
		EN: "启动脚本下载失败",
	}

	M00279 = &Item{
		ZH: "启动脚本解压失败",
		EN: "启动脚本解压失败",
	}

	M00280 = &Item{
		ZH: "未获取到相关信息",
		EN: "未获取到相关信息",
	}
	M00281 = &Item{
		ZH: "激活码信息错误",
		EN: "激活码信息错误",
	}
	M00282 = &Item{
		ZH: "设备码获取失败",
		EN: "设备码获取失败",
	}
	M00283 = &Item{
		ZH: "设备码不匹配",
		EN: "设备码不匹配",
	}
	M00284 = &Item{
		ZH: "激活码已过期",
		EN: "激活码已过期",
	}
	M00285 = &Item{
		ZH: "激活码更新失败",
		EN: "激活码更新失败",
	}

	M00300 = &Item{
		ZH: "等待设备注册",
		EN: "等待设备注册",
	}

	M00400 = &Item{
		ZH: "文件备份失败",
		EN: "文件备份失败",
	}
	M00401 = &Item{
		ZH: "更新失败, 请检查配置合法性",
		EN: "更新失败, 请检查配置合法性",
	}
	M00402 = &Item{
		ZH: "文件更新失败",
		EN: "文件更新失败",
	}
	M00403 = &Item{
		ZH: "地图解析失败",
		EN: "地图解析失败",
	}
)

var (
	M1001 = &Item{
		ZH: "数据库获取失败",
		EN: "数据库获取失败",
	}

	M1002 = &Item{
		ZH: "数据库设置失败",
		EN: "数据库设置失败",
	}
	M1003 = &Item{
		ZH: "数据库创建失败",
		EN: "数据库创建失败",
	}
	M1004 = &Item{
		ZH: "数据库删除失败",
		EN: "数据库删除失败",
	}
	// M1005 = &Item{
	// 	ZH: "缓存设置失败",
	// 	EN: "缓存设置失败",
	// }

	M1006 = &Item{
		ZH: "无权限",
		EN: "无权限",
	}
	M1007 = &Item{
		ZH: "未设置权限",
		EN: "未设置权限",
	}
)

var (
	MR1000 = &Item{
		ZH: "类型错误",
		EN: "类型错误",
	}
	MR1001 = &Item{
		ZH: "字段Data未设置",
		EN: "字段Data未设置",
	}
	MR1002 = &Item{
		ZH: "解析错误",
		EN: "解析错误",
	}
	MR1003 = &Item{
		ZH: "请求失败",
		EN: "请求失败",
	}
	MR1004 = &Item{
		ZH: "参数错误",
		EN: "参数错误",
	}
	MR1005 = &Item{
		ZH: "请求超时",
		EN: "请求超时",
	}
	MR1006 = &Item{
		ZH: "无权限",
		EN: "无权限",
	}
	MR1007 = &Item{
		ZH: "设置失败",
		EN: "设置失败",
	}
	MR1008 = &Item{
		ZH: "获取失败",
		EN: "获取失败",
	}
)

func Make(msg ...string) *Item {
	if len(msg) == 0 {
		return MR1003
	}

	if len(msg) == 1 {
		return &Item{msg[0], ""}
	}

	return &Item{msg[0], msg[1]}
}
