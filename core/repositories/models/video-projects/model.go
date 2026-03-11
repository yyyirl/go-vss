package videoProjects

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*VideoProjects)(nil)

// 录像计划
type VideoProjects struct {
	ID               uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;COMMENT:'主键'" json:"id"`
	Name             string `gorm:"column:name;type:varchar(50);NOT NULL;default:'';COMMENT:'计划名称'" json:"name"`
	State            uint   `gorm:"column:state;default:1;NOT NULL;comment:'启用状态 0 未启用 1 启用'" json:"state"`
	ChannelUniqueIds string `gorm:"column:channelUniqueIds;type:json;default:(json_array());comment:'通道id集合'" json:"channelUniqueIds"`
	Plans            string `gorm:"column:plans;NOT NULL" json:"plans"`
	CreatedAt        uint64 `gorm:"column:createdAt;default:0;comment:'创建时间'" json:"createdAt"`
	UpdatedAt        uint64 `gorm:"column:updatedAt;default:0;comment:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (v VideoProjects) ToMap() map[string]interface{} {
	return functions.StructToMap(v, "json", nil)
}

func (v VideoProjects) Columns() []string {
	return Columns
}

func (v VideoProjects) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (v VideoProjects) PrimaryKey() string {
	return PrimaryId
}

func (v VideoProjects) TableName() string {
	return "sk-video-projects"
}

func (v VideoProjects) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (v VideoProjects) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (v VideoProjects) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (v VideoProjects) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		v.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}

	if v.ChannelUniqueIds == "" {
		v.ChannelUniqueIds = "[]"
	}

	v.UpdatedAt = uint64(functions.NewTimer().NowMilli())
	return v
}

// CorrectionMap map数据修正
func (v VideoProjects) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())
	if v, ok := data[ColumnChannelUniqueIds]; ok {
		if val, ok := v.([]interface{}); ok {
			var ids []uint64
			for _, item := range val {
				if v, ok := item.(uint64); ok {
					ids = append(ids, v)
				}
			}

			b, err := functions.JSONMarshal(ids)
			if err == nil {
				data[ColumnChannelUniqueIds] = string(b)
			}
		}
	}

	return data
}

// UseCache 数据库缓存
func (v VideoProjects) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (v VideoProjects) ConvToItem() (*Item, error) {
	var useDBCache = false
	if v.DefaultModel != nil {
		useDBCache = v.DefaultModel.UseDBCache
	}

	var channelUniqueIds []uint64
	if v.ChannelUniqueIds != "" {
		if err := functions.JSONUnmarshal([]byte(v.ChannelUniqueIds), &channelUniqueIds); err != nil {
			return nil, err
		}
	}

	if channelUniqueIds == nil {
		channelUniqueIds = []uint64{}
	}

	return &Item{
		VideoProjects:    &v,
		ChannelUniqueIds: channelUniqueIds,
		UseDBCache:       useDBCache,
	}, nil
}

func (v VideoProjects) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(v)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
