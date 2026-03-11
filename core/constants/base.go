package constants

var (
	ENV string
)

const (
	ENV_DEVELOPMENT = "dev"
	ENV_PRODUCTION  = "pro"

	RESPONSE_AES_KEY = "HjU9yLmPqW4sVnB7"
)

const (
	EXT_JPG   = "jpg"
	EXT_MP3   = "mp3"
	EXT_MP4   = "mp4"
	EXT_PDF   = "pdf"
	EXT_AMR   = "amr"
	EXT_KNOWN = "known"
)

const (
	API_DOC_DIR = "source/doc/api"
)

const (
	ERR_LOCAL_CACHE_NOTFOUND = "Entry not found"
)

type (
	Lang struct {
		EN  string `json:"en"`
		ZH  string `json:"zh"`
		MEN string `json:"men,omitempty,optional"`
		MZH string `json:"mzh,omitempty,optional"`
	}
	ConfItemInner struct {
		Id        uint   `json:"id"`
		Column    string `json:"column,omitempty,optional"`
		Title     *Lang  `json:"title"`
		Brief     *Lang  `json:"brief,omitempty"`
		Banner    string `json:"banner,omitempty"`
		Icon      string `json:"icon,omitempty,optional"`
		ClassName string `json:"className,omitempty,optional"`
		Types     []uint `json:"types,omitempty,optional"`
	}
)

const (
	MODE_11_SECOND = "11-second"
	MODE_10_SECOND = "10-second"
	MODE_5_SECOND  = "5-second"
	MODE_10_MINUTE = "10-minute"
	MODE_SECOND    = "second"
	MODE_MINUTE    = "minute"
	MODE_HOUR      = "hour"
	MODE_DAY       = "day"
	MODE_MONTH     = "month"
)

// ------------------------------------ log
const (
	LOG_MODE_CONSOLE = "console"
	LOG_MODE_FILE    = "file"
	LOG_MODE_VOLUME  = "volume"
)

// ------------------------------------ 短信
const (
	SMS_KEY_REGISTER          = "register"
	SMS_KEY_LOGIN             = "login"
	SMS_KEY_FIND_PWD          = "findPwd"
	SMS_KEY_BIND              = "bind"
	SMS_KEY_UNBIND            = "unbind"
	SMS_KEY_BIND_BANK_ACCOUNT = "bind-bank-account"
)

// ------------------------------------ 邮箱
const (
	EMAIL_KEY_REGISTER          = "register"
	EMAIL_KEY_LOGIN             = "login"
	EMAIL_KEY_FIND_PWD          = "findPwd"
	EMAIL_KEY_BIND              = "bind"
	EMAIL_KEY_UNBIND            = "unbind"
	EMAIL_KEY_BIND_BANK_ACCOUNT = "bind-bank-account"
)

// 队列键值
const (
	QueueCameraShareplaceUpdateKey   = "camera-share-place-update"
	QueueInteractionsActionAddKey    = "interactions-action-add"
	QueueInteractionsActionDeleteKey = "interactions-action-delete"

	// 不在消息队列中读取
	QueueCameraShareplaceLabelVectorKey = "camera-shareplace-label-vector"
)

var (
	// 合并数据
	QueueCacheSlice = []string{}
	// 不合并数据
	QueueCacheNoMergingSlice = []string{
		QueueInteractionsActionDeleteKey,
	}
	QueueTypes = []struct {
		Name  string
		Limit int
	}{
		{
			Name:  QueueCameraShareplaceUpdateKey,
			Limit: 500,
		},
		{
			Name:  QueueInteractionsActionAddKey,
			Limit: 500,
		},
		{
			Name:  QueueInteractionsActionDeleteKey,
			Limit: 50,
		},
	}
)
