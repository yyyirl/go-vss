package svc

import (
	"skeyevss/core/app/sev/db/internal/config"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/repositories/models/admins"
	"skeyevss/core/repositories/models/alarms"
	"skeyevss/core/repositories/models/cascade"
	"skeyevss/core/repositories/models/channels"
	"skeyevss/core/repositories/models/crontab"
	"skeyevss/core/repositories/models/departments"
	"skeyevss/core/repositories/models/devices"
	"skeyevss/core/repositories/models/dictionaries"
	mediaServers "skeyevss/core/repositories/models/media-servers"
	"skeyevss/core/repositories/models/roles"
	"skeyevss/core/repositories/models/settings"
	systemOperationLogs "skeyevss/core/repositories/models/system-operation-logs"
	videoProjects "skeyevss/core/repositories/models/video-projects"
	"skeyevss/core/repositories/redis"
)

type ServiceContext struct {
	Config config.Config

	RedisClient *redis.Client
	DBClient    orm.DB

	AdminsModel              *admins.DB              // 管理员
	DictionariesModel        *dictionaries.DB        // 字典
	SettingModel             *settings.DB            // 设置
	RolesModel               *roles.DB               // 角色
	SystemOperationLogsModel *systemOperationLogs.DB // 操作日志
	DepartmentsModel         *departments.DB         // 部门
	DevicesModel             *devices.DB             // 设备
	ChannelsModel            *channels.DB            // 通道
	CrontabModel             *crontab.DB             // 任务
	MediaServersModel        *mediaServers.DB        // media server
	AlarmsModel              *alarms.DB              // 报警
	VideoProjectsModel       *videoProjects.DB       // 录像计划
	CascadeModel             *cascade.DB             // 设备级联

	DeviceDepIdSetChan chan struct{}
}

var svcCtx *ServiceContext

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库
	setDB(c)

	var (
		dbClient orm.DB
		dbConf   = &orm.DBConnectConfig{
			Mode:      c.Mode,
			Log:       c.Log,
			Databases: c.Databases,
		}
	)
	if c.SevBase.DatabaseType == "mysql" {
		dbClient = orm.NewMysqlConnect(dbConf)
	} else if c.SevBase.DatabaseType == "sqlite" {
		dbClient = orm.NewSqliteConnect(dbConf)
	} else {
		panic("unsupported db type")
	}

	var (
		adminsModel       = admins.NewDB(dbClient)
		rolesModel        = roles.NewDB(dbClient)
		dictionariesModel = dictionaries.NewDB(dbClient)
		settingsModel     = settings.NewDB(dbClient)
		crontabModel      = crontab.NewDB(dbClient)
		smsModel          = mediaServers.NewDB(dbClient)
	)

	// 初始化数据库
	autoMigrate(dbClient)
	backendAccounts(c, adminsModel)
	initTableRecords(c, dictionariesModel, settingsModel, crontabModel, rolesModel, smsModel)
	svcCtx = &ServiceContext{
		Config:   c,
		DBClient: dbClient,

		RedisClient: redis.New(c.Mode, c.Log.Encoding, c.CRedis, c.Log),

		AdminsModel:              adminsModel,
		RolesModel:               rolesModel,
		SystemOperationLogsModel: systemOperationLogs.NewDB(dbClient),
		DictionariesModel:        dictionariesModel,
		SettingModel:             settingsModel,
		DepartmentsModel:         departments.NewDB(dbClient),
		DevicesModel:             devices.NewDB(dbClient),
		ChannelsModel:            channels.NewDB(dbClient),
		CrontabModel:             crontabModel,
		MediaServersModel:        smsModel,
		AlarmsModel:              alarms.NewDB(dbClient),
		VideoProjectsModel:       videoProjects.NewDB(dbClient),
		CascadeModel:             cascade.NewDB(dbClient),
		DeviceDepIdSetChan:       make(chan struct{}),
	}
	// 任务
	proc()
	return svcCtx
}
