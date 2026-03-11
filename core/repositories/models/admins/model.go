package admins

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Admins)(nil)

type Admins struct {
	ID        uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Username  string `gorm:"column:username;type:char(70);uniqueIndex:admins_username;NOT NULL;comment:'用户名'" json:"username"`
	Password  string `gorm:"column:password;NOT NULL;comment:'密码'" json:"password"`
	Nickname  string `gorm:"column:nickname;NOT NULL;comment:'昵称'" json:"nickname"`
	Email     string `gorm:"column:email;NOT NULL;comment:'邮箱'" json:"email"`
	Mobile    string `gorm:"column:mobile;NOT NULL;comment:'手机号'" json:"mobile"`
	State     uint   `gorm:"column:state;default:0;comment:'使用状态0未启用1启用'" json:"state"`
	DepIds    string `gorm:"column:depIds;type:json;default:(json_array());comment:'部门id集合'" json:"depIds"`
	Remark    string `gorm:"column:remark;comment:'备注'" json:"remark"`
	Sex       uint   `gorm:"column:sex;default:0;comment:'性别1男2女'" json:"sex"`
	Avatar    string `gorm:"column:avatar;comment:'头像'" json:"avatar"`
	Super     int    `gorm:"column:super;default:0;comment:'是否超级管理员1'" json:"super"`
	IsDel     int    `gorm:"column:isDel;default:0;comment:'删除状态1已删除'" json:"isDel"`
	CreatedAt uint64 `gorm:"column:createdAt;default:0;NOT NULL;comment:'创建时间'" json:"createdAt"`
	UpdatedAt uint64 `gorm:"column:updatedAt;default:0;NOT NULL;comment:'更新时间';comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (a Admins) ToMap() map[string]interface{} {
	return functions.StructToMap(a, "json", nil)
}

func (a Admins) Columns() []string {
	return Columns
}

func (a Admins) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (a Admins) PrimaryKey() string {
	return PrimaryId
}

func (a Admins) TableName() string {
	return "sk-admins"
}

func (a Admins) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (a Admins) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (a Admins) OnConflictColumns(_ []string) []string {
	return nil
}

func (a Admins) Correction(action orm.ActionType) interface{} {
	if a.DepIds == "" {
		a.DepIds = "[]"
	}
	if action == orm.ActionInsert {
		a.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	a.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return a
}

func (a Admins) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())
	if v, ok := data[ColumnPassword]; ok {
		if pwd, ok := v.(string); ok {
			if !functions.IsBcryptHash(pwd) {
				data[ColumnPassword], _ = functions.GeneratePwd(pwd)
			}
		}
	}

	if v, ok := data[ColumnDepIds]; ok {
		if val, ok := v.([]interface{}); ok {
			var ids []uint64
			for _, item := range val {
				id, err := functions.InterfaceToNumber[uint64](item)
				if err != nil {
					continue
				}
				ids = append(ids, id)
			}

			b, err := functions.JSONMarshal(ids)
			if err == nil {
				data[ColumnDepIds] = string(b)
			}
		}
	}

	return data
}

func (a Admins) UseCache() *orm.UseCacheAdvanced {
	return &orm.UseCacheAdvanced{
		Create:         true,
		Query:          true,
		Delete:         true,
		Update:         true,
		Row:            true,
		Raw:            true,
		CacheKeyPrefix: a.TableName(),
		Driver:         new(orm.CacheMemoryDriver),
		Expire:         60,
	}

	// return &orm.UseCacheAdvanced{
	// 	Create:         true,
	// 	Query:          true,
	// 	Delete:         true,
	// 	Update:         true,
	// 	Row:            true,
	// 	Raw:            true,
	// 	CacheKeyPrefix: a.TableName(),
	// 	Driver:         new(orm.CacheRedisDriver),
	// 	// Driver: new(orm.CacheMemoryDriver),
	// 	Expire: 600000,
	// }
}

func (a Admins) ConvToItem() (*Item, error) {
	var depIds []uint64
	if a.DepIds == "" {
		depIds = []uint64{}
	} else {
		if err := functions.ConvStringToType(a.DepIds, &depIds); err != nil {
			return nil, err
		}
	}

	var useDBCache = false
	if a.DefaultModel != nil {
		useDBCache = a.DefaultModel.UseDBCache
	}

	return &Item{
		Admins:     &a,
		UseDBCache: useDBCache,
		DepIds:     depIds,
	}, nil
}

func (a Admins) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(a)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
