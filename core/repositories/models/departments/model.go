package departments

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Departments)(nil)

// 组织机构

type Departments struct {
	ID                 uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Name               string `gorm:"column:name;NOT NULL;comment:'组织部门名称'" json:"name"`       // 组织部门名称
	Remark             string `gorm:"column:remark;NOT NULL;comment:'备注'" json:"remark"`       // 备注
	ParentId           uint64 `gorm:"column:parentId;NOT NULL;comment:'父级部门'" json:"parentId"` // 父级部门
	CascadeDepUniqueId string `gorm:"column:cascadeDepUniqueId;type:char(70);uniqueIndex:departments_cascadeDepUniqueId;DEFAULT:NULL;COMMENT:'级联编号'" json:"cascadeDepUniqueId"`
	RoleIds            string `gorm:"column:roleIds;comment:'角色id集合'" json:"roleIds"`                         // 角色id集合
	State              uint   `gorm:"column:state;default:1;NOT NULL;comment:'启用状态 0 未启用 1 启用'" json:"state"` // 启用状态 0 未启用 1 启用
	CreatedAt          uint64 `gorm:"column:createdAt;default:0;NOT NULL;comment:'创建时间'" json:"createdAt"`
	UpdatedAt          uint64 `gorm:"column:updatedAt;default:0;NOT NULL;comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (d Departments) ToMap() map[string]interface{} {
	return functions.StructToMap(d, "json", nil)
}

func (d Departments) Columns() []string {
	return Columns
}

func (d Departments) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (d Departments) PrimaryKey() string {
	return PrimaryId
}

func (d Departments) TableName() string {
	return "sk-departments"
}

func (d Departments) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (d Departments) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (d Departments) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (d Departments) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		d.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	d.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return d
}

// CorrectionMap map数据修正
func (d Departments) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())
	return data
}

// UseCache 数据库缓存
func (d Departments) UseCache() *orm.UseCacheAdvanced {
	return &orm.UseCacheAdvanced{
		Create:         true,
		Query:          true,
		Delete:         true,
		Update:         true,
		Row:            true,
		Raw:            true,
		CacheKeyPrefix: d.TableName(),
		Driver:         new(orm.CacheMemoryDriver),
		Expire:         60,
	}
}

// ConvToItem 数据转换
func (d Departments) ConvToItem() (*Item, error) {
	var roleIds []uint64
	if d.RoleIds == "" {
		roleIds = []uint64{}
	} else {
		if err := functions.ConvStringToType(d.RoleIds, &roleIds); err != nil {
			return nil, err
		}
	}

	var useDBCache = false
	if d.DefaultModel != nil {
		useDBCache = d.DefaultModel.UseDBCache
	}

	return &Item{
		Departments: &d,
		RoleIds:     roleIds,
		UseDBCache:  useDBCache,
	}, nil
}

func (d Departments) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(d)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
