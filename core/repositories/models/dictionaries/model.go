package dictionaries

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Dictionaries)(nil)

type Dictionaries struct {
	ID         uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Name       string `gorm:"column:name;NOT NULL;DEFAULT:'';comment:'名称 默认'" json:"name"`
	UniqueId   string `gorm:"column:uniqueId;uniqueIndex:dictionaries_uniqueId;type:CHAR(40);NOT NULL;comment:'唯一id 标注值'" json:"uniqueId"`
	MultiValue string `gorm:"column:multiValue;type:VARCHAR(255);NOT NULL;DEFAULT:'';comment:'多值匹配 多个值用\n分隔'" json:"multiValue"`
	ParentId   uint   `gorm:"column:parentId;DEFAULT:0;NOT NULL;comment:'父级id'" json:"parentId"`
	State      uint   `gorm:"column:state;DEFAULT:0;NOT NULL;comment:'启用状态 0 未启用 1 启用'" json:"state"`
	Readonly   uint   `gorm:"column:readonly;DEFAULT:0;NOT NULL;comment:'启用状态 1 只读'" json:"readonly"`
	CreatedAt  uint64 `gorm:"column:createdAt;DEFAULT:0;NOT NULL;comment:'创建时间'" json:"createdAt"`
	UpdatedAt  uint64 `gorm:"column:updatedAt;DEFAULT:0;NOT NULL;comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (d Dictionaries) ToMap() map[string]interface{} {
	return functions.StructToMap(d, "json", nil)
}

func (d Dictionaries) Columns() []string {
	return Columns
}

func (d Dictionaries) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (d Dictionaries) PrimaryKey() string {
	return PrimaryId
}

func (d Dictionaries) TableName() string {
	return "sk-dictionaries"
}

func (d Dictionaries) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (d Dictionaries) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (d Dictionaries) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (d Dictionaries) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		d.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	d.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return d
}

// CorrectionMap map数据修正
func (d Dictionaries) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())

	return data
}

// UseCache 数据库缓存
func (d Dictionaries) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (d Dictionaries) ConvToItem() (*Item, error) {
	var useDBCache = false
	if d.DefaultModel != nil {
		useDBCache = d.DefaultModel.UseDBCache
	}

	return &Item{
		Dictionaries: &d,
		UseDBCache:   useDBCache,
	}, nil
}

func (d Dictionaries) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(d)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
