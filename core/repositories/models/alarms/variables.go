package alarms

var (
	ColumnID               = "id"
	ColumnDeviceUniqueId   = "deviceUniqueId"
	ColumnAlarmMethod      = "alarmMethod"
	ColumnAlarmPriority    = "alarmPriority"
	ColumnAlarmDescription = "alarmDescription"
	ColumnLongitude        = "longitude"
	ColumnLatitude         = "latitude"
	ColumnAlarmType        = "alarmType"
	ColumnEventType        = "eventType"
	ColumnSnapshot         = "snapshot"
	ColumnVideo            = "video"
	ColumnCreatedAt        = "createdAt"
)

var Columns = []string{
	ColumnID,
	ColumnDeviceUniqueId,
	ColumnAlarmMethod,
	ColumnAlarmPriority,
	ColumnAlarmDescription,
	ColumnLongitude,
	ColumnLatitude,
	ColumnAlarmType,
	ColumnEventType,
	ColumnSnapshot,
	ColumnVideo,
	ColumnCreatedAt,
}

const (
	PrimaryId = "id"
)

// 报警类型
const (
	_            uint = iota
	AlarmType_1       // 视频丢失报警 => 1
	AlarmType_2       // 设备防拆报警 => 2
	AlarmType_3       // 存储欖形闲备磁盘满报警 => 3
	AlarmType_4       // 设备高温报警 => 4
	AlarmType_5       // 设备低温报警 => 5
	AlarmType_6       // 人工视频报警 => 1
	AlarmType_7       // 运动目标检测报警 => 2
	AlarmType_8       // 遗留物检测报警 => 3
	AlarmType_9       // 物体移除检测报警 => 4
	AlarmType_10      // 绊线检测报警 => 5
	AlarmType_11      // 人侵检测报警 => 6
	AlarmType_12      // 逆行检测报警 => 7
	AlarmType_13      // 徘徊检测报警 => 8
	AlarmType_14      // 流量统计报警 => 9
	AlarmType_15      // 密度检测报警 => 10
	AlarmType_16      // 视频异常检测报警 => 11
	AlarmType_17      // 快速移动报警 => 12
	AlarmType_18      // 存储设备磁盘故障报警 => 1
	AlarmType_19      // 存储设备风扇故障报警。 => 2
)

var AlarmTypes = map[uint]string{
	AlarmType_1:  "视频丢失报警",
	AlarmType_2:  "设备防拆报警",
	AlarmType_3:  "存储欖形闲备磁盘满报警",
	AlarmType_4:  "设备高温报警",
	AlarmType_5:  "设备低温报警",
	AlarmType_6:  "人工视频报警",
	AlarmType_7:  "运动目标检测报警",
	AlarmType_8:  "遗留物检测报警",
	AlarmType_9:  "物体移除检测报警",
	AlarmType_10: "绊线检测报警",
	AlarmType_11: "人侵检测报警",
	AlarmType_12: "逆行检测报警",
	AlarmType_13: "徘徊检测报警",
	AlarmType_14: "流量统计报警",
	AlarmType_15: "密度检测报警",
	AlarmType_16: "视频异常检测报警",
	AlarmType_17: "快速移动报警",
	AlarmType_18: "存储设备磁盘故障报警",
	AlarmType_19: "存储设备风扇故障报警",
}

// 报警类型扩展参数
const (
	_ uint = iota
	EventType_1
	EventType_2
)

var EventTypes = map[uint]string{
	EventType_1: "进入区域",
	EventType_2: "离开区域",
}

// 报警方式
const (
	_             uint = iota
	AlarmMethod_1      // 电话报警
	AlarmMethod_2      // 设备报警
	AlarmMethod_3      // 短信报警
	AlarmMethod_4      // GPS报警
	AlarmMethod_5      // 视频报警,
	AlarmMethod_6      // 设备故障报警
	AlarmMethod_7      // 其他报警
)

var AlarmMethods = map[uint]string{
	AlarmMethod_1: "电话报警",
	AlarmMethod_2: "设备报警",
	AlarmMethod_3: "短信报警",
	AlarmMethod_4: "GPS报警",
	AlarmMethod_5: "视频报警",
	AlarmMethod_6: "设备故障报警",
	AlarmMethod_7: "其他报警",
}

// 报警级别
const (
	_               uint = iota
	AlarmPriority_1      // 为一级警情
	AlarmPriority_2      // 为二级警情
	AlarmPriority_3      // 为三级警情
	AlarmPriority_4      // 为四级警情
)

var AlarmPriorities = map[uint]string{
	AlarmPriority_1: "一级警情",
	AlarmPriority_2: "二级警情",
	AlarmPriority_3: "三级警情",
	AlarmPriority_4: "四级警情",
}
