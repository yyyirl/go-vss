package alarms

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Alarms)(nil)

// 报警记录
type Alarms struct {
	ID               uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;COMMENT:'主键'" json:"id"`
	DeviceUniqueId   string `gorm:"column:deviceUniqueId;index:deviceUniqueId;type:char(70);comment:'设备id';NOT NULL" json:"deviceUniqueId"`
	AlarmMethod      uint   `gorm:"column:alarmMethod;type:tinyint(5);comment:'报警方式';default:0;NOT NULL" json:"alarmMethod"`
	AlarmPriority    uint   `gorm:"column:alarmPriority;type:tinyint(5);comment:'报警级别';default:0;NOT NULL" json:"alarmPriority"`
	AlarmDescription string `gorm:"column:alarmDescription;type:varchar(255);NOT NULL;comment:'报警描述'" json:"alarmDescription"`
	Longitude        string `gorm:"column:longitude;varchar(100);NOT NULL;DEFAULT:'';COMMENT:'经度'" json:"longitude"`
	Latitude         string `gorm:"column:latitude;varchar(100);NOT NULL;DEFAULT:'';COMMENT:'纬度'" json:"latitude"`
	AlarmType        uint   `gorm:"column:alarmType;type:tinyint(5);comment:'报警类型';default:0;NOT NULL" json:"alarmType"`
	EventType        uint   `gorm:"column:eventType;type:tinyint(5);comment:'报警类型扩展参数';default:0;NOT NULL" json:"eventType"`
	Snapshot         string `gorm:"column:snapshot;type:varchar(255);comment:'快照';default:'';NOT NULL" json:"snapshot"`
	Video            string `gorm:"column:video;type:varchar(255);comment:'录像';default:'';NOT NULL" json:"video"`
	CreatedAt        uint64 `gorm:"column:createdAt;type:bigint;default:0;NOT NULL;COMMENT:'报警时间'" json:"createdAt"`

	*orm.DefaultModel
}

func (a Alarms) ToMap() map[string]interface{} {
	return functions.StructToMap(a, "json", nil)
}

func (a Alarms) Columns() []string {
	return Columns
}

func (a Alarms) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (a Alarms) PrimaryKey() string {
	return PrimaryId
}

func (a Alarms) TableName() string {
	return "sk-alarms"
}

func (a Alarms) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (a Alarms) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (a Alarms) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (a Alarms) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		a.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}

	return a
}

// CorrectionMap map数据修正
func (a Alarms) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	return data
}

// UseCache 数据库缓存
func (a Alarms) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (a Alarms) ConvToItem() (*Item, error) {
	var useDBCache = false
	if a.DefaultModel != nil {
		useDBCache = a.DefaultModel.UseDBCache
	}

	return &Item{
		Alarms:     &a,
		UseDBCache: useDBCache,
	}, nil
}

func (a Alarms) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(a)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
