package skSolutions

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*SkSolutions)(nil)

type SkSolutions struct {
	ID          uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Title       string `gorm:"column:title;type:varchar(100);NOT NULL;comment:'标题" json:"title"`
	State       uint   `gorm:"column:state;type:tinyint(4);default:0;NOT NULL;comment:'启用状态 0 未上架 1 上架'" json:"state"`
	Description string `gorm:"column:description;type:varchar(255);NOT NULL;comment:'描述简介'" json:"description"`
	Banner      string `gorm:"column:banner;type:varchar(255);NOT NULL;comment:'首图'" json:"banner"`
	Content     string `gorm:"column:content;type:longtext;NOT NULL;comment:'内容'" json:"content"`
	PublishAt   uint64 `gorm:"column:publishAt;default:0;NOT NULL;comment:'发布时间'" json:"publishAt"`
	CreatedAt   uint64 `gorm:"column:createdAt;default:0;NOT NULL;comment:'创建时间'" json:"createdAt"`
	UpdatedAt   uint64 `gorm:"column:updatedAt;default:0;NOT NULL;comment:'更新时间';comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (s SkSolutions) ToMap() map[string]interface{} {
	return functions.StructToMap(s, "json", nil)
}

func (s SkSolutions) Columns() []string {
	return Columns
}

func (s SkSolutions) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (s SkSolutions) PrimaryKey() string {
	return PrimaryId
}

func (s SkSolutions) TableName() string {
	return "skSolutions"
}

func (s SkSolutions) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s SkSolutions) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s SkSolutions) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (s SkSolutions) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		s.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	s.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return s
}

// CorrectionMap map数据修正
func (s SkSolutions) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())
	return data
}

// UseCache 数据库缓存
func (s SkSolutions) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (s SkSolutions) ConvToItem() (*Item, error) {
	var useDBCache = false
	if s.DefaultModel != nil {
		useDBCache = s.DefaultModel.UseDBCache
	}

	return &Item{
		SkSolutions: &s,
		UseDBCache:  useDBCache,
	}, nil
}

func (s SkSolutions) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(s)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
