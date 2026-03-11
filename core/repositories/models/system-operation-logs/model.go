package systemOperationLogs

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*SystemOperationLogs)(nil)

// 系统操作日志

type SystemOperationLogs struct {
	ID        uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Userid    uint64 `gorm:"column:userid;NOT NULL;comment:'管理员id'" json:"userid"` // 管理员id
	Type      uint   `gorm:"column:type;NOT NULL;comment:'操作类型'" json:"type"`      // 操作类型
	Data      string `gorm:"column:data;NOT NULL;comment:'操作数据内容'" json:"data"`    // 操作数据内容
	IP        string `gorm:"column:ip;NOT NULL;comment:'ip'" json:"ip"`            // ip
	Mac       string `gorm:"column:mac;NOT NULL;comment:'mac地址'" json:"mac"`       // mac地址
	CreatedAt uint64 `gorm:"column:createdAt;default:0;NOT NULL;comment:'创建时间'" json:"createdAt"`
	UpdatedAt uint64 `gorm:"column:updatedAt;default:0;NOT NULL;comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (s SystemOperationLogs) ToMap() map[string]interface{} {
	return functions.StructToMap(s, "json", nil)
}

func (s SystemOperationLogs) Columns() []string {
	return Columns
}

func (s SystemOperationLogs) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (s SystemOperationLogs) PrimaryKey() string {
	return PrimaryId
}

func (s SystemOperationLogs) TableName() string {
	return "sk-system-operation-logs"
}

func (s SystemOperationLogs) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s SystemOperationLogs) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s SystemOperationLogs) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (s SystemOperationLogs) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		s.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	s.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return s
}

// CorrectionMap map数据修正
func (s SystemOperationLogs) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())

	return data
}

// UseCache 数据库缓存
func (s SystemOperationLogs) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (s SystemOperationLogs) ConvToItem() (*Item, error) {
	var useDBCache = false
	if s.DefaultModel != nil {
		useDBCache = s.DefaultModel.UseDBCache
	}

	return &Item{
		SystemOperationLogs: &s,
		UseDBCache:          useDBCache,
	}, nil
}

func (s SystemOperationLogs) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(s)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
