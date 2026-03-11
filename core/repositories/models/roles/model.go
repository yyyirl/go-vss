package roles

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Roles)(nil)

type Roles struct {
	ID                  uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Name                string `gorm:"column:name;NOT NULL;comment:'角色名称'" json:"name"`                        // 角色名称
	PermissionUniqueIds string `gorm:"column:permissionUniqueIds;comment:'权限id集合'" json:"permissionUniqueIds"` // 权限id集合
	State               uint   `gorm:"column:state;default:0;comment:'启用状态 0 默认不启用 1 启用'" json:"state"`        // 启用状态 0 默认不启用 1 启用
	Remark              string `gorm:"column:remark;comment:'备注'" json:"remark"`                               // 备注
	IsDel               uint   `gorm:"column:isDel;default:0;comment:'删除状态1已删除'" json:"isDel"`
	CreatedAt           uint64 `gorm:"column:createdAt;default:0;comment:'创建时间'" json:"createdAt"`
	UpdatedAt           uint64 `gorm:"column:updatedAt;default:0;comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (r Roles) ToMap() map[string]interface{} {
	return functions.StructToMap(r, "json", nil)
}

func (r Roles) Columns() []string {
	return Columns
}

func (r Roles) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (r Roles) PrimaryKey() string {
	return PrimaryId
}

func (r Roles) TableName() string {
	return "sk-roles"
}

func (r Roles) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (r Roles) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (r Roles) OnConflictColumns(_ []string) []string {
	return nil
}

func (r Roles) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		r.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}

	if r.PermissionUniqueIds == "" {
		r.PermissionUniqueIds = "[]"
	}

	r.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return r
}

func (r Roles) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())
	if v, ok := data[ColumnPermissionUniqueIds]; ok {
		if val, ok := v.([]interface{}); ok {
			var ids []string
			for _, item := range val {
				if v, ok := item.(string); ok {
					ids = append(ids, v)
				}
			}

			b, err := functions.JSONMarshal(ids)
			if err == nil {
				data[ColumnPermissionUniqueIds] = string(b)
			}
		}
	}

	return data
}

func (r Roles) UseCache() *orm.UseCacheAdvanced {
	return &orm.UseCacheAdvanced{
		Create:         true,
		Query:          true,
		Delete:         true,
		Update:         true,
		Row:            true,
		Raw:            true,
		CacheKeyPrefix: r.TableName(),
		Driver:         new(orm.CacheMemoryDriver),
		Expire:         60,
	}
}

func (r Roles) ConvToItem() (*Item, error) {
	var useDBCache = false
	if r.DefaultModel != nil {
		useDBCache = r.DefaultModel.UseDBCache
	}

	var permissionUniqueIds []string
	if r.PermissionUniqueIds != "" {
		if err := functions.JSONUnmarshal([]byte(r.PermissionUniqueIds), &permissionUniqueIds); err != nil {
			return nil, err
		}
	}

	if permissionUniqueIds == nil {
		permissionUniqueIds = []string{}
	}

	return &Item{
		Roles:               &r,
		PermissionUniqueIds: permissionUniqueIds,
		UseDBCache:          useDBCache,
	}, nil
}
