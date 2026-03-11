/**
 * @Author:         yi
 * @Description:    auto_migrate
 * @Version:        1.0.0
 * @Date:           2025/4/25 16:33
 */

package svc

import (
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"skeyevss/core/app/sev/db/internal/config"
	"skeyevss/core/common/source/permissions"
	"skeyevss/core/pkg/functions"
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
)

func setDB(c config.Config) {
	if !c.SevBase.UseMysql {
		return
	}
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/mysql",
			c.SevBase.MysqlUsername,
			c.SevBase.MysqlPassword,
			c.SevBase.MysqlHost,
			c.SevBase.MysqlPort,
		),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err = db.Ping(); err != nil {
		panic(err)
	}

	if _, err = db.Exec("CREATE DATABASE IF NOT EXISTS `" + c.Databases.BaseDBName + "` CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;"); err != nil {
		panic(err)
	}
}

// 初始化更新密码
// 测试打包 redis mysql 启动状态

func backendAccounts(c config.Config, adminsModel *admins.DB) {
	list, err := adminsModel.List(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{
				Column: admins.ColumnUsername,
				Values: []interface{}{
					c.Accounts.BackendUsername,
					c.Accounts.BackendSuperUsername,
				},
			},
		},
	})
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(err)
		}
	}

	if len(list) <= 0 {
		password1, _ := functions.GeneratePwd(c.Accounts.BackendPassword)
		password2, _ := functions.GeneratePwd(c.Accounts.BackendSuperPassword)
		if err := adminsModel.Upsert(
			[]admins.Admins{
				{ID: 1, Username: c.Accounts.BackendUsername, Password: password1, Super: 1},
				{ID: 2, Username: c.Accounts.BackendSuperUsername, Password: password2, Super: 1},
			},
			[]string{admins.ColumnId, admins.ColumnUsername},
		); err != nil {
			panic(err)
		}
	}

	if c.UseShowcaseAccount {
		password3, _ := functions.GeneratePwd(c.Accounts.BackendShowcasePassword)
		if err := adminsModel.Upsert(
			[]admins.Admins{
				{ID: 3, Username: c.Accounts.BackendShowcaseUsername, Password: password3, Super: 1},
			},
			[]string{admins.ColumnId, admins.ColumnUsername},
		); err != nil {
			panic(err)
		}
	}
}

func autoMigrate(db *gorm.DB) {
	functions.PrintStyle("cyan-underline", "------>>>>>> 初始化数据表 ...")

	// 基础数据 ---------------------------------------------
	if err := db.AutoMigrate(new(admins.Admins)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(roles.Roles)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(systemOperationLogs.SystemOperationLogs)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(dictionaries.Dictionaries)); err != nil {
		panic(err)
	}

	// if !db.Migrator().HasIndex(new(crontab.Crontab), "Crontab_uniqueId") {
	//
	// }

	_ = db.Migrator().DropIndex(new(crontab.Crontab), "Crontab_uniqueId")

	if err := db.AutoMigrate(new(crontab.Crontab)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(mediaServers.MediaServers)); err != nil {
		panic(err)
	}

	// if !db.Migrator().HasIndex(new(dictionaries.Dictionaries), dictionaries.ColumnUniqueId) {
	// 	if err := db.Migrator().CreateIndex(new(dictionaries.Dictionaries), dictionaries.ColumnUniqueId); err != nil {
	// 		panic(err)
	// 	}
	// }

	if err := db.AutoMigrate(new(departments.Departments)); err != nil {
		panic(err)
	}
	// 基础数据 ---------------------------------------------

	// data ---------------------------------------------

	if err := db.AutoMigrate(new(devices.Devices)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(channels.Channels)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(settings.Settings)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(alarms.Alarms)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(videoProjects.VideoProjects)); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(new(cascade.Cascade)); err != nil {
		panic(err)
	}

	// data ---------------------------------------------

	functions.PrintStyle("cyan-underline", "------>>>>>> 数据表初始化完成")
}

func initTableRecords(
	c config.Config,
	dictionariesModel *dictionaries.DB,
	settingModel *settings.DB,
	crontabModel *crontab.DB,
	roleModel *roles.DB,
	smsModel *mediaServers.DB,
) {
	var now = uint64(functions.NewTimer().NowMilli())
	if err := dictionariesModel.Upsert(
		dbInitTableRecords.Dictionaries,
		[]string{
			dictionaries.ColumnUniqueId,
			dictionaries.ColumnId,
		},
	); err != nil {
		panic(err)
	}

	_ = settingModel.Insert([]settings.Settings{{ID: 1}})
	_ = crontabModel.Insert(crontab.InitRecords)
	_ = crontabModel.DeleteBy(&orm.ReqParams{
		Conditions: []*orm.ConditionItem{
			{Column: crontab.ColumnUniqueId, Values: functions.SliceToSliceAny(crontab.UniqueIds), Operator: "NOTIN"},
		},
	})

	_ = roleModel.Insert(
		[]roles.Roles{
			{
				ID:                  1,
				Name:                "设备操作员",
				PermissionUniqueIds: permissions.EquipmentOperator(),
				State:               1,
			},
		},
	)

	_ = smsModel.Upsert(
		[]mediaServers.MediaServers{
			{
				ID:                       1,
				Name:                     "default",
				IP:                       c.InternalIp,
				ExtIP:                    c.ExternalIp,
				Port:                     uint(c.SevBase.MediaServerPort),
				MediaServerStreamPortMin: c.Sip.MediaServerStreamPortMin,
				MediaServerStreamPortMax: c.Sip.MediaServerStreamPortMax,
				IsDef:                    1,
				State:                    1,
				CreatedAt:                now,
				UpdatedAt:                now,
			},
		},
		[]string{
			dictionaries.ColumnUniqueId,
			dictionaries.ColumnId,
		},
	)
}
