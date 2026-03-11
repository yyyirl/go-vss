package crontab

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Crontab)(nil)

// 任务
type Crontab struct {
	UniqueId string `gorm:"column:uniqueId;uniqueIndex:Crontab_uniqueId;type:CHAR(40);NOT NULL;comment:'唯一id 标注值'" json:"uniqueId"`
	Title    string `gorm:"column:title;type:varchar(100);NOT NULL" json:"title"`

	Interval uint64 `gorm:"column:interval;type:int(11);default:1;NOT NULL;comment:'执行周期单位/s'" json:"interval"`
	Status   uint   `gorm:"column:status;type:tinyint(4);default:1;NOT NULL;comment:'启用状态 1 启用'" json:"status"`
	Readonly uint   `gorm:"column:readonly;type:tinyint(4);default:0;NOT NULL;comment:'1 只读'" json:"readonly"`

	Timeout     uint `gorm:"column:timeout;type:int(11);default:10;NOT NULL;comment:'超时时间 单位/s'" json:"timeout"`
	Counter     uint `gorm:"column:counter;type:int(11);default:1;NOT NULL;comment:'每个周期 每一批次执行数量'" json:"counter"`
	BlockStatus uint `gorm:"column:blockStatus;type:tinyint(4);default:1;NOT NULL;comment:'阻塞状态 1 每个周期必须等待所有批次执行完成后再进入下一个周期 0 无需等待当前批次执行完成可进入下一个周期'" json:"blockStatus"`

	Logs      string `gorm:"column:logs;type:json;default:(json_array());comment:'logs 保留100条'" json:"logs"`
	CreatedAt uint64 `gorm:"column:createdAt;DEFAULT:0;NOT NULL;comment:'创建时间'" json:"createdAt"`
	UpdatedAt uint64 `gorm:"column:updatedAt;DEFAULT:0;NOT NULL;comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (c Crontab) ToMap() map[string]interface{} {
	return functions.StructToMap(c, "json", nil)
}

func (c Crontab) Columns() []string {
	return Columns
}

func (c Crontab) UniqueKeys() []string {
	return []string{
		PrimaryUniqueId,
	}
}

func (c Crontab) PrimaryKey() string {
	return PrimaryUniqueId
}

func (c Crontab) TableName() string {
	return "sk-crontab"
}

func (c Crontab) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (c Crontab) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (c Crontab) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (c Crontab) Correction(action orm.ActionType) interface{} {
	if c.Logs == "" {
		c.Logs = "[]"
	}

	if action == orm.ActionInsert {
		c.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}
	c.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return c
}

// CorrectionMap map数据修正
func (c Crontab) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())

	if v, ok := data[ColumnLogs]; ok {
		if val, ok := v.([]interface{}); ok {
			var logs []string
			for _, item := range val {
				record, ok := item.(string)
				if !ok {
					continue
				}
				logs = append(logs, record)
			}

			b, err := functions.JSONMarshal(logs)
			if err == nil {
				data[ColumnLogs] = string(b)
			}
		}
	}

	return data
}

// UseCache 数据库缓存
func (c Crontab) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (c Crontab) ConvToItem() (*Item, error) {
	var useDBCache = false
	if c.DefaultModel != nil {
		useDBCache = c.DefaultModel.UseDBCache
	}

	var logs []string
	if c.Logs == "" {
		logs = []string{}
	} else {
		if err := functions.ConvStringToType(c.Logs, &logs); err != nil {
			return nil, err
		}
	}

	return &Item{
		Crontab:    &c,
		UseDBCache: useDBCache,
		Logs:       logs,
	}, nil
}

func (c Crontab) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(c)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
