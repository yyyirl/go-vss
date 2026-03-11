package channels

var (
	ColumnID                      = "id"
	ColumnUniqueId                = "uniqueId"
	ColumnDeviceUniqueId          = "deviceUniqueId"
	ColumnOriginalChannelUniqueId = "originalChannelUniqueId"
	ColumnCascadeChannelUniqueId  = "cascadeChannelUniqueId"
	ColumnCascadeDepUniqueId      = "cascadeDepUniqueId"
	ColumnIsCascade               = "isCascade"
	ColumnName                    = "name"
	ColumnLabel                   = "label"
	ColumnPtzType                 = "ptzType"
	ColumnStreamUrl               = "streamUrl"
	ColumnCdnState                = "cdnState"
	ColumnCdnUrl                  = "cdnUrl"
	ColumnLongitude               = "longitude"
	ColumnLatitude                = "latitude"
	ColumnOnDemandLiveState       = "onDemandLiveState"
	ColumnAudioState              = "audioState"
	ColumnTransCodedState         = "transCodedState"
	ColumnRecordingState          = "recordingState"
	ColumnOnline                  = "online"
	ColumnOnlineAt                = "onlineAt"
	ColumnStreamState             = "streamState"
	ColumnStreamMSId              = "streamMSId"
	ColumnOriginal                = "original"
	ColumnVideos                  = "videos"
	ColumnScreenshots             = "screenshots"
	ColumnSnapshot                = "snapshot"
	ColumnDepIds                  = "depIds"
	ColumnParental                = "parental"
	ColumnParentID                = "parentID"
	ColumnCreatedAt               = "createdAt"
	ColumnUpdatedAt               = "updatedAt"
)

var Columns = []string{
	ColumnID,
	ColumnUniqueId,
	ColumnDeviceUniqueId,
	ColumnOriginalChannelUniqueId,
	ColumnCascadeChannelUniqueId,
	ColumnCascadeDepUniqueId,
	ColumnIsCascade,
	ColumnName,
	ColumnLabel,
	ColumnPtzType,
	ColumnStreamUrl,
	ColumnCdnState,
	ColumnCdnUrl,
	ColumnLongitude,
	ColumnLatitude,
	ColumnOnDemandLiveState,
	ColumnAudioState,
	ColumnTransCodedState,
	ColumnRecordingState,
	ColumnOnline,
	ColumnOnlineAt,
	ColumnStreamState,
	ColumnStreamMSId,
	ColumnOriginal,
	ColumnVideos,
	ColumnScreenshots,
	ColumnSnapshot,
	ColumnDepIds,
	ColumnParentID,
	ColumnParental,
	ColumnCreatedAt,
	ColumnUpdatedAt,
}

const (
	PrimaryId = "id"
)

// 摄像机云台类型
const (
	PTXType_0 uint = iota
	PTXType_1
	PTXType_2
	PTXType_3
	PTXType_4
)

var PTXTypes = map[uint]string{
	PTXType_0: "未知",
	PTXType_1: "球机",
	PTXType_2: "半球",
	PTXType_3: "固定枪机",
	PTXType_4: "遥控枪机",
}
