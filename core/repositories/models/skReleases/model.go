package skReleases

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*SkReleases)(nil)

type SkReleases struct {
	ID          uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Title       string `gorm:"column:title;type:varchar(100);NOT NULL;comment:'标题" json:"title"`
	Version     string `gorm:"column:version;type:char(20);NOT NULL;comment:'版本号'" json:"version"`
	State       uint   `gorm:"column:state;type:tinyint(4);default:0;NOT NULL;comment:'启用状态 0 未上架 1 上架'" json:"state"`
	Description string `gorm:"column:description;type:longtext;NOT NULL;comment:'描述'" json:"description"`
	CreatedAt   uint64 `gorm:"column:createdAt;default:0;NOT NULL;comment:'创建时间'" json:"createdAt"`
	UpdatedAt   uint64 `gorm:"column:updatedAt;default:0;NOT NULL;comment:'更新时间';comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (s SkReleases) ToMap() map[string]interface{} {
	return functions.StructToMap(s, "json", nil)
}

func (s SkReleases) Columns() []string {
	return Columns
}

func (s SkReleases) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (s SkReleases) PrimaryKey() string {
	return PrimaryId
}

func (s SkReleases) TableName() string {
	return "skReleases"
}

func (s SkReleases) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s SkReleases) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s SkReleases) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (s SkReleases) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		s.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	s.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return s
}

// CorrectionMap map数据修正
func (s SkReleases) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())
	return data
}

// UseCache 数据库缓存
func (s SkReleases) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (s SkReleases) ConvToItem() (*Item, error) {
	var useDBCache = false
	if s.DefaultModel != nil {
		useDBCache = s.DefaultModel.UseDBCache
	}

	return &Item{
		SkReleases: &s,
		UseDBCache: useDBCache,
	}, nil
}

func (s SkReleases) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(s)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
