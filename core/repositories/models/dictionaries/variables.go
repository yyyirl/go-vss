package dictionaries

var (
	ColumnId         = "id"
	ColumnName       = "name"
	ColumnUniqueId   = "uniqueId"
	ColumnMultiValue = "multiValue"
	ColumnParentId   = "parentId"
	ColumnState      = "state"
	ColumnReadonly   = "readonly"
	ColumnCreatedAt  = "createdAt"
	ColumnUpdatedAt  = "updatedAt"
)

var Columns = []string{
	ColumnId,
	ColumnName,
	ColumnUniqueId,
	ColumnMultiValue,
	ColumnParentId,
	ColumnState,
	ColumnReadonly,
	ColumnCreatedAt,
	ColumnUpdatedAt,
}

const (
	PrimaryId = "id"
)

const (
	UniqueIdDeviceManufacturer    = "device-manufacturer"
	UniqueIdDeviceManufacturer_2  = "device-manufacturer-2"  // Hikvision
	UniqueIdDeviceManufacturer_3  = "device-manufacturer-3"  // Dahua
	UniqueIdDeviceManufacturer_4  = "device-manufacturer-4"  // UNIVIEW
	UniqueIdDeviceManufacturer_5  = "device-manufacturer-5"  // OPENSKEYE
	UniqueIdDeviceManufacturer_6  = "device-manufacturer-6"  // 海康
	UniqueIdDeviceManufacturer_7  = "device-manufacturer-7"  // 大华
	UniqueIdDeviceManufacturer_8  = "device-manufacturer-8"  // 宇视
	UniqueIdDeviceManufacturer_9  = "device-manufacturer-9"  // 视开
	UniqueIdDeviceManufacturer_20 = "device-manufacturer-20" // 未知
)

const (
	UniqueIdAlarmLevel   = "device-alarm-level"   // 报警级别
	UniqueIdAlarmLevel_1 = "device-alarm-level-1" // 1级警情
	UniqueIdAlarmLevel_2 = "device-alarm-level-2" // 2级警情
	UniqueIdAlarmLevel_3 = "device-alarm-level-3" // 3级警情
	UniqueIdAlarmLevel_4 = "device-alarm-level-4" // 4级警情
)

const (
	UniqueIdAlarmType   = "alarm-type"   // 报警方式
	UniqueIdAlarmType_1 = "alarm-type-1" // 电话报警
	UniqueIdAlarmType_2 = "alarm-type-2" // 设备报警
	UniqueIdAlarmType_3 = "alarm-type-3" // 短信报警
	UniqueIdAlarmType_4 = "alarm-type-4" // GPS报警
	UniqueIdAlarmType_5 = "alarm-type-5" // 视频报警
	UniqueIdAlarmType_6 = "alarm-type-6" // 设备故障报警
	UniqueIdAlarmType_7 = "alarm-type-7" // 其他报警
)
