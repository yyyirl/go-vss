package orm

import (
	"os"
	"path"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/repositories/redis"
	"skeyevss/core/tps"
)

func NewDBX[T Model](DB DB, model T) *DBX[T] {
	return &DBX[T]{
		DB:            DB,
		originalModel: model,
	}
}

type DBConnectConfig struct {
	Mode        string
	Log         logx.LogConf
	Databases   tps.YamlDatabases
	RedisClient *redis.Client
}

func newConnect(c *DBConnectConfig, dialector gorm.Dialector, Type string) DB {
	var logLevel = Info
	if c.Mode == service.ProMode {
		logLevel = Error
	}

	_ = functions.MakeDir(c.Log.Path + "/" + Type)

	var (
		ticker = functions.TickerWithDuration(2, func() {
			functions.LogError("数据库链接超时")
			functions.PrintStyle("red", "Error: ", "数据库链接超时")
			os.Exit(1)
		})
		// 日志输出
		logFilePath = ""
	)

	defer ticker.Stop()
	if c.Databases.SaveSqlDir != "" {
		logFilePath = c.Databases.SaveSqlDir + "/" + Type
		_ = functions.MakeDir(logFilePath)
	}
	// var loggerInstance = logger.Default.LogMode(logger.Info)
	var loggerInstance = NewStdFileLogger(
		FileLogConfig{
			Config: logger.Config{
				SlowThreshold: 200 * time.Millisecond,
				Colorful:      true,
			},
		},
		logFilePath,
	).LogMode(logLevel)

	db, err := gorm.Open(
		dialector,
		&gorm.Config{
			Logger:      loggerInstance,
			PrepareStmt: true,
		},
	)
	if err != nil {
		panic(err)
	}

	if err := db.Use(
		NewCachePlugin(
			c.RedisClient,
			cache.New(1*time.Minute, 5*time.Minute),
		),
	); err != nil {
		panic(err)
	}

	_db, err := db.DB()
	if err != nil {
		panic("failed to get underlying sql.DB")
	}

	// 连接池
	_db.SetMaxIdleConns(10)
	_db.SetMaxOpenConns(2000)
	_db.SetConnMaxLifetime(30 * time.Second)

	return db
}

func NewMysqlConnect(c *DBConnectConfig) DB {
	if c.Databases.MysqlBase == "" {
		panic("Databases.MysqlBase 不能为空")
	}

	return newConnect(c, mysql.Open(c.Databases.MysqlBase), "mysql")
}

func NewSqliteConnect(c *DBConnectConfig) DB {
	if c.Databases.SqliteBase == "" {
		panic("Databases.SqliteBase 不能为空")
	}

	var (
		data = strings.Split(c.Databases.SqliteBase, "?")
		dir  = path.Dir(data[0])
	)
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}

	return newConnect(c, sqlite.Open(c.Databases.SqliteBase), "sqlite")
}
