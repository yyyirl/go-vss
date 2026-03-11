package skAdmins

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*SkAdmins)(nil)

type SkAdmins struct {
	ID        uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Username  string `gorm:"column:username;NOT NULL;comment:'用户名'" json:"username"`
	Password  string `gorm:"column:password;NOT NULL;comment:'密码'" json:"password"`
	CreatedAt uint64 `gorm:"column:createdAt;default:0;NOT NULL;comment:'创建时间'" json:"createdAt"`
	UpdatedAt uint64 `gorm:"column:updatedAt;default:0;NOT NULL;comment:'更新时间';comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (s SkAdmins) ToMap() map[string]interface{} {
	return functions.StructToMap(s, "json", nil)
}

func (s SkAdmins) Columns() []string {
	return Columns
}

func (s SkAdmins) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (s SkAdmins) PrimaryKey() string {
	return PrimaryId
}

func (s SkAdmins) TableName() string {
	return "skAdmins"
}

func (s SkAdmins) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s SkAdmins) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s SkAdmins) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (s SkAdmins) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		s.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	s.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return s
}

// CorrectionMap map数据修正
func (s SkAdmins) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())
	return data
}

// UseCache 数据库缓存
func (s SkAdmins) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (s SkAdmins) ConvToItem() (*Item, error) {
	var useDBCache = false
	if s.DefaultModel != nil {
		useDBCache = s.DefaultModel.UseDBCache
	}

	return &Item{
		SkAdmins:   &s,
		UseDBCache: useDBCache,
	}, nil
}

func (s SkAdmins) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(s)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
