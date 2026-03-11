package mediaServers

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*MediaServers)(nil)

// media server管理
type MediaServers struct {
	ID                       uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;COMMENT:'主键'" json:"id"`
	Name                     string `gorm:"column:name;NOT NULL;type:varchar(50);COMMENT:'设备名称'" json:"name"`
	IP                       string `gorm:"column:ip;NOT NULL;type:char(30);COMMENT:'服务ip(内网)'" json:"ip"`
	ExtIP                    string `gorm:"column:extIP;NOT NULL;type:char(30);DEFAULT:'';COMMENT:'服务ip(外网)'" json:"extIP"`
	Port                     uint   `gorm:"column:port;NOT NULL;type:int(4);COMMENT:'服务端口'" json:"port"`
	MediaServerStreamPortMin uint   `gorm:"column:mediaServerStreamPortMin;default:15000;NOT NULL;type:int(4);COMMENT:'推流端口范围最小值'" json:"mediaServerStreamPortMin"`
	MediaServerStreamPortMax uint   `gorm:"column:mediaServerStreamPortMax;default:19000;NOT NULL;type:int(4);COMMENT:'推流端口范围最大值'" json:"mediaServerStreamPortMax"`
	IsDef                    uint   `gorm:"column:isDef;default:0;NOT NULL;type:tinyint(4);COMMENT:'是否是默认服务 0 否 1 是'" json:"isDef"`
	State                    uint   `gorm:"column:state;default:1;NOT NULL;type:tinyint(4);COMMENT:'启用状态 0 未启用 1 启用'" json:"state"`
	CreatedAt                uint64 `gorm:"column:createdAt;DEFAULT:0;NOT NULL;COMMENT:'创建时间'" json:"createdAt"`
	UpdatedAt                uint64 `gorm:"column:updatedAt;DEFAULT:0;NOT NULL;COMMENT:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (m MediaServers) ToMap() map[string]interface{} {
	return functions.StructToMap(m, "json", nil)
}

func (m MediaServers) Columns() []string {
	return Columns
}

func (m MediaServers) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (m MediaServers) PrimaryKey() string {
	return PrimaryId
}

func (m MediaServers) TableName() string {
	return "sk-media-servers"
}

func (m MediaServers) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (m MediaServers) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (m MediaServers) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (m MediaServers) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		m.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	m.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return m
}

// CorrectionMap map数据修正
func (m MediaServers) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())
	return data
}

// UseCache 数据库缓存
func (m MediaServers) UseCache() *orm.UseCacheAdvanced {
	//	return &orm.UseCacheAdvanced{
	//		Query:   true,
	//		Update: true,
	//	 CacheKeyPrefix: m.TableName(),
	//		Driver: new(orm.CacheRedisDriver),
	//		Expire: 60,
	//	}
	return nil
}

// ConvToItem 数据转换
func (m MediaServers) ConvToItem() (*Item, error) {
	var useDBCache = false
	if m.DefaultModel != nil {
		useDBCache = m.DefaultModel.UseDBCache
	}

	return &Item{
		MediaServers: &m,
		UseDBCache:   useDBCache,
	}, nil
}

func (m MediaServers) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(m)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
