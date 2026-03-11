package cascade

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Cascade)(nil)

// 平台级联
type Cascade struct {
	ID       uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;COMMENT:'主键'" json:"id"`
	Name     string `gorm:"column:name;type:varchar(50);NOT NULL;default:'';COMMENT:'通道名称'" json:"name"`
	UniqueId string `gorm:"column:uniqueId;type:char(70);uniqueIndex:Cascade_UniqueId;NOT NULL;COMMENT:'uid'" json:"uniqueId"`

	Protocol          uint   `gorm:"column:protocol;type:tinyint(4);default:1;NOT NULL;COMMENT:'信令传输协议 1 TCP 2 UDP'" json:"protocol"`
	SipId             string `gorm:"column:sipId;type:varchar(100);NOT NULL;COMMENT:'SIP服务国标编码'" json:"sipId"`
	SipDomain         string `gorm:"column:sipDomain;type:varchar(100);NOT NULL;COMMENT:'SIP服务国标域'" json:"sipDomain"`
	SipIp             string `gorm:"column:sipIp;type:varchar(100);NOT NULL;COMMENT:'SIP服务IP'" json:"sipIp"`
	SipPort           uint   `gorm:"column:sipPort;type:int(6);NOT NULL;COMMENT:'SIP服务端口'" json:"sipPort"`
	Username          string `gorm:"column:username;uniqueIndex:Cascade_Username;type:varchar(100);NOT NULL;COMMENT:'SIP认证用户'" json:"username"`
	Password          string `gorm:"column:password;type:varchar(100);NOT NULL;COMMENT:'SIP认证密码'" json:"password"`
	LocalIp           string `gorm:"column:localIp;type:varchar(100);NOT NULL;COMMENT:'本地级联IP'" json:"localIp"`
	KeepaliveInterval uint   `gorm:"column:keepaliveInterval;type:int(6);default:60;NOT NULL;COMMENT:'心跳间隔(秒)'" json:"keepaliveInterval"`
	RegisterInterval  uint   `gorm:"column:registerInterval;type:int(6);default:60;NOT NULL;COMMENT:'注册间隔(秒)'" json:"registerInterval"`
	RegisterTimeout   uint   `gorm:"column:registerTimeout;type:int(6);default:3600;NOT NULL;COMMENT:'注册有效期(秒)'" json:"registerTimeout"`
	CommandTransport  uint   `gorm:"column:commandTransport;type:tinyint(4);default:1;NOT NULL;COMMENT:'信令传输 1 TCP 2 UDP'" json:"commandTransport"`
	State             uint   `gorm:"column:state;type:tinyint(4);default:1;NOT NULL;COMMENT:'启用状态 0 未启用 1 启用'" json:"state"`
	Online            uint   `gorm:"column:online;type:tinyint(4);default:0;NOT NULL;COMMENT:'在线状态 0 不在线 1 在线'" json:"online"`
	Relations         string `gorm:"column:relations;type:json;default:(json_array());comment:'分组/通道集合'" json:"relations"`
	CatalogGroupSize  uint   `gorm:"column:catalogGroupSize;type:tinyint(4);default:4;comment:'目录分组大小'" json:"catalogGroupSize"`

	CreatedAt uint64 `gorm:"column:createdAt;type:bigint;default:0;NOT NULL;COMMENT:'创建时间'" json:"createdAt"`
	UpdatedAt uint64 `gorm:"column:updatedAt;type:bigint;default:0;NOT NULL;COMMENT:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (c Cascade) ToMap() map[string]interface{} {
	return functions.StructToMap(c, "json", nil)
}

func (c Cascade) Columns() []string {
	return Columns
}

func (c Cascade) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (c Cascade) PrimaryKey() string {
	return PrimaryId
}

func (c Cascade) TableName() string {
	return "sk-cascade"
}

func (c Cascade) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (c Cascade) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (c Cascade) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (c Cascade) Correction(action orm.ActionType) interface{} {
	if c.Relations == "" {
		c.Relations = "[]"
	}

	if action == orm.ActionInsert {
		c.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	c.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return c
}

// CorrectionMap map数据修正
func (c Cascade) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())

	if v, ok := data[ColumnRelations]; ok {
		var screenshots []string
		if err := functions.ConvInterface(v, &screenshots); err != nil {
			return data
		}

		b, err := functions.JSONMarshal(screenshots)
		if err != nil {
			return data
		}

		data[ColumnRelations] = string(b)
	}
	return data
}

// UseCache 数据库缓存
func (c Cascade) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (c Cascade) ConvToItem() (*Item, error) {
	var useDBCache = false
	if c.DefaultModel != nil {
		useDBCache = c.DefaultModel.UseDBCache
	}

	var relation []RelationItem
	if c.Relations == "" {
		relation = []RelationItem{}
	} else {
		if err := functions.ConvStringToType(c.Relations, &relation); err != nil {
			return nil, err
		}
	}

	return &Item{
		Relations:  relation,
		Cascade:    &c,
		UseDBCache: useDBCache,
	}, nil
}

func (c Cascade) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(c)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
