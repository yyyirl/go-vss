package channels

import (
	"fmt"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

var _ orm.Model = (*Channels)(nil)

// 通道
type Channels struct {
	ID                      uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;COMMENT:'主键'" json:"id"`
	DeviceUniqueId          string `gorm:"column:deviceUniqueId;type:char(70);uniqueIndex:device_channel_UniqueId;NOT NULL;COMMENT:'设备id'" json:"deviceUniqueId"`
	UniqueId                string `gorm:"column:uniqueId;type:char(70);uniqueIndex:device_channel_UniqueId;NOT NULL;COMMENT:'通道id'" json:"uniqueId"`
	OriginalChannelUniqueId string `gorm:"column:originalChannelUniqueId;type:varchar(255);NOT NULL;DEFAULT:'';comment:'原始id'" json:"originalChannelUniqueId"`
	Name                    string `gorm:"column:name;type:varchar(50);NOT NULL;default:'';COMMENT:'通道名称'" json:"name"`
	Label                   string `gorm:"column:label;type:varchar(255);NOT NULL;DEFAULT:'';comment:'自定义标签'" json:"label"`

	CascadeChannelUniqueId string `gorm:"column:cascadeChannelUniqueId;type:char(70);uniqueIndex:device_CascadeChannelUniqueId;DEFAULT:NULL;COMMENT:'级联通道id'" json:"cascadeChannelUniqueId"`
	CascadeDepUniqueId     string `gorm:"column:cascadeDepUniqueId;type:char(70);device_cascadeDepUniqueId;DEFAULT:'';COMMENT:'级联父级(分组/设备/平台)id'" json:"cascadeDepUniqueId"`
	IsCascade              uint   `gorm:"column:isCascade;type:tinyint(4);default:0;NOT NULL;COMMENT:'是否是级联设备(本机级联) 1 是'" json:"isCascade"`

	PTZType   uint   `gorm:"column:ptzType;type:tinyint(4);default:0;NOT NULL;COMMENT:'摄像机云台类型：0-未知 1-球机，2-半球，3-固定枪机，4-遥控枪机 （是设备的情况下有效）'" json:"ptzType"`
	StreamUrl string `gorm:"column:streamUrl;type:varchar(255);NOT NULL;default:'';COMMENT:'接入码流地址,ONVIF和流媒体源接入设备有效'" json:"streamUrl"`

	CdnState uint   `gorm:"column:cdnState;type:tinyint(4);default:0;NOT NULL;COMMENT:'cdn开启状态 0 未开启 1 开启'" json:"cdnState"`
	CdnUrl   string `gorm:"column:cdnUrl;type:varchar(100);NOT NULL;default:'';COMMENT:'cdn地址'" json:"cdnUrl"`

	Longitude float64 `gorm:"column:longitude;type:decimal(11,8);NOT NULL;DEFAULT:0;COMMENT:'通道经度'" json:"longitude"`
	Latitude  float64 `gorm:"column:latitude;type:decimal(11,8);NOT NULL;DEFAULT:0;COMMENT:'通道纬度'" json:"latitude"`

	OnDemandLiveState uint   `gorm:"column:onDemandLiveState;type:tinyint(4);default:0;NOT NULL;COMMENT:'按需直播 0 未开启 1 开启'" json:"onDemandLiveState"`
	AudioState        uint   `gorm:"column:audioState;type:tinyint(4);default:0;NOT NULL;COMMEnNT:'开启音频: 0 未开启 1 开启'" json:"audioState"`
	TransCodedState   uint   `gorm:"column:transCodedState;type:tinyint(4);default:0;NOT NULL;COMMENT:'是否转码 0 未开启 1 开启'" json:"transCodedState"`
	OnlineAt          uint64 `gorm:"column:onlineAt;type:bigint;default:0;NOT NULL;COMMENT:'上线时间'" json:"onlineAt"`
	Online            uint   `gorm:"column:online;type:tinyint(4);default:1;NOT NULL;COMMENT:'在线状态: 0 不在线 1 在线'" json:"online"`
	StreamState       uint   `gorm:"column:streamState;type:tinyint(4);default:0;NOT NULL;COMMENT:'流状态: 0 没有流 1 正在拉流或者推流(区分推流和拉流)'" json:"streamState"`
	RecordingState    uint   `gorm:"column:recordingState;type:tinyint(4);default:0;NOT NULL;COMMENT:'视频录制中'" json:"recordingState"`
	StreamMSId        uint64 `gorm:"column:streamMSId;type:int(11);default:0;NOT NULL;COMMENT:'当前流正在使用的media server id'" json:"streamMSId"`
	ParentID          string `gorm:"column:parentID;type:char(70);default:'';NOT NULL;COMMENT:'当前通道父级目录id'" json:"parentID"`
	Parental          uint   `gorm:"column:parental;type:tinyint(4);default:0;NOT NULL;COMMENT:'设备是否有子设备,有表示是组织架构或者目录，没有表示是设备通道（1：有(目录)，0：没有(通道)）'" json:"parental"`

	Original    string `gorm:"column:original;type:json;default:(json_object());NOT NULL;COMMENT:'原始数据,对应catalog item'" json:"original"`
	Videos      string `gorm:"column:videos;type:json;default:(json_array());NOT NULL;comment:'已保存录像'" json:"videos"`
	Screenshots string `gorm:"column:screenshots;type:json;default:(json_array());NOT NULL;comment:'截图'" json:"screenshots"`
	Snapshot    string `gorm:"column:snapshot;type:varchar(255);default:'';NOT NULL;comment:'快照'" json:"snapshot"`
	DepIds      string `gorm:"column:depIds;type:json;default:(json_array());NOT NULL;comment:'部门id'" json:"depIds"`

	CreatedAt uint64 `gorm:"column:createdAt;type:bigint;default:0;NOT NULL;COMMENT:'创建时间'" json:"createdAt"`
	UpdatedAt uint64 `gorm:"column:updatedAt;type:bigint;default:0;NOT NULL;COMMENT:'更新时间'" json:"updatedAt"`

	*orm.DefaultModel
}

func (c Channels) ToMap() map[string]interface{} {
	return functions.StructToMap(c, "json", nil)
}

func (c Channels) Columns() []string {
	return Columns
}

func (c Channels) UniqueKeys() []string {
	return []string{
		PrimaryId,
	}
}

func (c Channels) PrimaryKey() string {
	return PrimaryId
}

func (c Channels) TableName() string {
	return "sk-channels"
}

func (c Channels) QueryConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (c Channels) SetConditions(conditions []*orm.ConditionItem) []*orm.ConditionItem {
	return conditions
}

func (c Channels) OnConflictColumns(_ []string) []string {
	return nil
}

// Correction 数据修正
func (c Channels) Correction(action orm.ActionType) interface{} {
	if action == orm.ActionInsert {
		c.CreatedAt = uint64(functions.NewTimer().NowMilli())
	}

	if c.Original == "" {
		c.Original = "{}"
	}

	if c.Videos == "" {
		c.Videos = "[]"
	}

	if c.DepIds == "" {
		c.DepIds = "[]"
	}

	c.UpdatedAt = uint64(functions.NewTimer().NowMilli())

	return c
}

// CorrectionMap map数据修正
func (c Channels) CorrectionMap(data map[string]interface{}) map[string]interface{} {
	data[ColumnUpdatedAt] = uint64(functions.NewTimer().NowMilli())

	if v, ok := data[ColumnOriginal]; ok {
		if val, ok := v.(map[string]interface{}); ok {
			b, err := functions.JSONMarshal(val)
			if err == nil {
				data[ColumnOriginal] = string(b)
			}
		}

		if v, ok := data[ColumnVideos]; ok {
			var videos []*VideoItem
			if err := functions.ConvInterface(v, &videos); err != nil {
				return data
			}

			for _, item := range videos {
				if item.Path == "" || item.Date == nil {
					return data
				}

				if item.Date.Start <= 0 || item.Date.End <= 0 {
					return data
				}
			}

			b, err := functions.JSONMarshal(videos)
			if err != nil {
				return data
			}

			data[ColumnVideos] = string(b)
		}

		if v, ok := data[ColumnDepIds]; ok {
			var depIds []uint64
			if err := functions.ConvInterface(v, &depIds); err != nil {
				return data
			}

			b, err := functions.JSONMarshal(depIds)
			if err != nil {
				return data
			}

			data[ColumnDepIds] = string(b)
		}

		if v, ok := data[ColumnScreenshots]; ok {
			var screenshots []string
			if err := functions.ConvInterface(v, &screenshots); err != nil {
				return data
			}

			b, err := functions.JSONMarshal(screenshots)
			if err != nil {
				return data
			}

			data[ColumnScreenshots] = string(b)
		}
	}

	return data
}

// UseCache 数据库缓存
func (c Channels) UseCache() *orm.UseCacheAdvanced {
	return &orm.UseCacheAdvanced{
		Create:       true,
		Query:        true,
		Delete:       true,
		Update:       true,
		UpdateDelete: true,
		Row:          true,
		Raw:          true,

		CacheKeyPrefix: c.TableName(),
		// Driver:         new(orm.CacheRedisDriver),
		Driver: new(orm.CacheMemoryDriver),
		Expire: 60,
	}
}

// ConvToItem 数据转换
func (c Channels) ConvToItem() (*Item, error) {
	var useDBCache = false
	if c.DefaultModel != nil {
		useDBCache = c.DefaultModel.UseDBCache
	}

	var original map[string]interface{}
	if c.Original == "" {
		original = map[string]interface{}{}
	} else {
		if err := functions.ConvStringToType(c.Original, &original); err != nil {
			return nil, err
		}
	}

	var videos []*VideoItem
	if c.Videos == "" {
		videos = []*VideoItem{}
	} else {
		if err := functions.ConvStringToType(c.Videos, &videos); err != nil {
			return nil, err
		}
	}

	var depIds []uint64
	if c.DepIds == "" {
		depIds = []uint64{}
	} else {
		if err := functions.ConvStringToType(c.DepIds, &depIds); err != nil {
			return nil, err
		}
	}

	var screenshots []string
	if c.Screenshots == "" {
		screenshots = []string{}
	} else {
		if err := functions.ConvStringToType(c.Screenshots, &screenshots); err != nil {
			return nil, err
		}
	}

	return &Item{
		Channels:    &c,
		Videos:      videos,
		Screenshots: screenshots,
		Original:    original,
		DepIds:      depIds,
		UseDBCache:  useDBCache,
	}, nil
}

func (c Channels) Conv(data interface{}) error {
	b, err := functions.JSONMarshal(c)
	if err != nil {
		return err
	}

	return functions.JSONUnmarshal(b, data)
}

type XListItem struct {
	ID             uint64   `gorm:"column:id;primary_key;AUTO_INCREMENT;COMMENT:'主键'" json:"id"`
	UniqueId       string   `gorm:"column:uniqueId;type:char(70);uniqueIndex:channels_UniqueId;NOT NULL;COMMENT:'通道id'" json:"uniqueId"`
	DeviceUniqueId string   `gorm:"column:deviceUniqueId;type:char(70);index:channels_DeviceUniqueId;NOT NULL;COMMENT:'设备id'" json:"deviceUniqueId"`
	Name           string   `gorm:"column:name;type:varchar(50);NOT NULL;default:'';COMMENT:'通道名称'" json:"name"`
	Label          string   `gorm:"column:label;type:varchar(255);NOT NULL;DEFAULT:'';comment:'自定义标签'" json:"label"`
	TDepIds        string   `gorm:"column:depIds;type:json;default:(json_array());NOT NULL;comment:'部门id'" json:"-"`
	DepIds         []uint64 `gorm:"column:depIds;type:json;default:(json_array());NOT NULL;comment:'部门id'" json:"depIds"`
	Online         uint     `gorm:"column:online;type:tinyint(4);default:1;NOT NULL;COMMENT:'在线状态: 0 不在线 1 在线'" json:"online"`
	Parental       uint     `gorm:"column:parental;type:tinyint(4);default:0;NOT NULL;COMMENT:'设备是否有子设备,有表示是组织架构或者目录，没有表示是设备通道（1：有，0：没有）'" json:"parental"`
	ParentID       string   `gorm:"column:parentID;type:char(70);default:'';NOT NULL;COMMENT:'当前通道父级目录id'" json:"parentID"`
}

func NewXList() *XListItem {
	return new(XListItem)
}

func (cs *XListItem) columns() string {
	return fmt.Sprintf(
		"`%s`, `%s`, `%s`, `%s`, `%s`, `%s`, `%s`, `%s`, `%s`",
		ColumnID,
		ColumnUniqueId,
		ColumnDeviceUniqueId,
		ColumnName,
		ColumnLabel,
		ColumnDepIds,
		ColumnOnline,
		ColumnParentID,
		ColumnParental,
	)
}

type XListItem1 struct {
	ID             uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;COMMENT:'主键'" json:"id"`
	UniqueId       string `gorm:"column:uniqueId;type:char(70);uniqueIndex:channels_UniqueId;NOT NULL;COMMENT:'通道id'" json:"uniqueId"`
	DeviceUniqueId string `gorm:"column:deviceUniqueId;type:char(70);index:channels_DeviceUniqueId;NOT NULL;COMMENT:'设备id'" json:"deviceUniqueId"`
}

func NewXList1() *XListItem1 {
	return new(XListItem1)
}

func (cs *XListItem1) columns() string {
	return fmt.Sprintf(
		"`%s`, `%s`, `%s`",
		ColumnID,
		ColumnUniqueId,
		ColumnDeviceUniqueId,
	)
}

type OnlineStateListItem struct {
	ID             uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;COMMENT:'主键'" json:"id"`
	UniqueId       string `gorm:"column:uniqueId;type:char(70);uniqueIndex:channels_UniqueId;NOT NULL;COMMENT:'通道id'" json:"uniqueId"`
	DeviceUniqueId string `gorm:"column:deviceUniqueId;uniqueIndex:devices_uniqueIndex;type:CHAR(70);NOT NULL;comment:'设备id'" json:"deviceUniqueId"`
	Online         uint   `gorm:"column:online;type:tinyint(4);default:1;NOT NULL;comment:'在线状态 0 不在线 1 在线'" json:"online"`
}

func NewOnlineStateList() *OnlineStateListItem {
	return new(OnlineStateListItem)
}

func (cs *OnlineStateListItem) columns() string {
	return fmt.Sprintf(
		"`%s`, `%s`, `%s`, `%s`",
		ColumnID,
		ColumnUniqueId,
		ColumnDeviceUniqueId,
		ColumnOnline,
	)
}
