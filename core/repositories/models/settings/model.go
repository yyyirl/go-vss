package settings

import (
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Settings)(nil)

// 设置表

type Content struct {
	// 默认播放方式
	MediaServerVideoPlayAddressType string `gorm:"column:-" json:"mediaServerVideoPlayAddressType"`
	// 黑名单 IP
	BanIp string `gorm:"column:-" json:"banIp"`
	// logo
	Logo string `gorm:"column:-" json:"logo"`
	// 管理后台标题
	WebManageTitle string `gorm:"column:-" json:"webManageTitle"`
	// 官网地址
	Website string `gorm:"column:-" json:"website"`
	// 地图地址
	MapTiles string `gorm:"column:-" json:"mapTiles"`
	// 地图中心点
	MapCenterPoints string `gorm:"column:-" json:"mapCenterPoints"`
	// 地图放大倍数
	MapZoom int `gorm:"column:-" json:"mapZoom"`
}

type Settings struct {
	ID      uint   `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'主键'" json:"id"`
	Content string `gorm:"column:content;type:json;NOT NUL;default:(json_object())" json:"content"`

	*orm.DefaultModel
}

func (s Settings) ToMap() map[string]interface{} {
	return functions.StructToMap(s, "json", nil)
}

func (s Settings) Columns() []string {
	return Columns
}

func (s Settings) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (s Settings) PrimaryKey() string {
	return PrimaryId
}

func (s Settings) TableName() string {
	return "sk-settings"
}

func (s Settings) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s Settings) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (s Settings) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (s Settings) Correction(_ orm.ActionType) interface{} {
	return s
}

// CorrectionMap map数据修正
func (s Settings) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	if v, ok := data[ColumnContent]; ok {
		b, err := functions.JSONMarshal(v)
		if err != nil {
			data[ColumnContent] = "{}"
		} else {
			data[ColumnContent] = string(b)
		}

	}
	return data
}

// UseCache 数据库缓存
func (s Settings) UseCache() *orm.UseCacheAdvanced {
	return nil
}

// ConvToItem 数据转换
func (s Settings) ConvToItem() (*Item, error) {
	var useDBCache = false
	if s.DefaultModel != nil {
		useDBCache = s.DefaultModel.UseDBCache
	}

	var content Content
	if s.Content == "" {
		content = Content{}
	} else {
		if err := functions.ConvStringToType(s.Content, &content); err != nil {
			return nil, err
		}
	}

	return &Item{
		Settings:   &s,
		Content:    &content,
		UseDBCache: useDBCache,
	}, nil
}

func (s Settings) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(s)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}
