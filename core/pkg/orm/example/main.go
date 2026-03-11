/**
 * @Author:         yi
 * @Description:    orm
 * @Version:        1.0.0
 * @Date:           2024/6/18 18:19
 */
package main

import (
	"flag"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"skeyevss/core/constants"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/repositories/models/admins"
	"skeyevss/core/repositories/redis"
	"skeyevss/core/tps"
)

type Config struct {
	rest.RestConf
	tps.YamlAuth

	AesKey      string
	LoginExpire int64

	SavePath tps.YamlSavePath

	Elasticsearch tps.YamlElasticsearch
	Mysql         tps.YamlDatabases
	Redis         tps.YamlRedis
	Email         tps.YamlEmail
	AliCloud      tps.YamlAli

	RedisHost,
	MysqlHost,
	ElasticsearchHost string
}

var configFile = flag.String(
	"f",
	"etc/.backend-api.yaml",
	"the config file",
)

func dbTest(userModel *admins.DB, adminsModel *admins.DB, page int) {
	go func() {
		if _, err := adminsModel.Exists(1); err != nil {
			panic(err)
		}
	}()

	var (
		now = time.Now()
		// 今年
		year = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).UnixMilli()
		// 上月1号
		lastMonth = now.AddDate(0, -1, 1).UnixMilli()
		// 当月1号
		month = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).UnixMilli()
		// 今天凌晨
		today = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
		// 昨日凌晨
		yesterday = now.AddDate(0, 0, -1).UnixMilli()
	)

	// 获取列表
	go func() {
		if _, err := userModel.List(&orm.ReqParams{
			Limit: 1,
			Page:  page,
			Orders: []*orm.OrderItem{
				{Column: admins.ColumnId, Value: "desc"},
			},
		}); err != nil {
			panic(err)
		}
	}()

	// 今日注册
	go func() {
		if _, err := userModel.Count(&orm.ReqParams{Conditions: []*orm.ConditionItem{{Column: admins.ColumnCreatedAt, Value: today, Operator: ">="}}}); err != nil {
			panic(err)
		}
	}()

	// 昨日注册
	go func() {
		if _, err := userModel.Count(
			&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: admins.ColumnCreatedAt, Value: yesterday, Operator: ">="},
					{Column: admins.ColumnCreatedAt, Value: today, Operator: "<="},
				},
			},
		); err != nil {
			panic(err)
		}
	}()
	// 本月注册
	go func() {
		if _, err := userModel.Count(
			&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: admins.ColumnCreatedAt, Value: month, Operator: ">="},
				},
			},
		); err != nil {
			panic(err)
		}
	}()
	// 上月注册
	go func() {
		if _, err := userModel.Count(
			&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: admins.ColumnCreatedAt, Value: lastMonth, Operator: ">="},
					{Column: admins.ColumnCreatedAt, Value: month, Operator: "<="},
				},
			},
		); err != nil {
			panic(err)
		}
	}()
	// 今年
	go func() {
		if _, err := userModel.Count(
			&orm.ReqParams{
				Conditions: []*orm.ConditionItem{
					{Column: admins.ColumnCreatedAt, Value: year, Operator: ">="},
				},
			},
		); err != nil {
			panic(err)
		}
	}()
	// 总用户
	go func() {
		if _, err := userModel.Count(new(orm.ReqParams)); err != nil {
			panic(err)
		}
	}()
	// 获取总数
	go func() {
		if _, err := userModel.Count(new(orm.ReqParams)); err != nil {
			panic(err)
		}
	}()
	// 获取列表
	go func() {
		if _, err := userModel.List(&orm.ReqParams{
			Limit: 1,
			Page:  page,
			Orders: []*orm.OrderItem{
				{Column: admins.ColumnId, Value: "desc"},
			},
		}); err != nil {
			panic(err)
		}
	}()
	// 查询
	go func() {
		if _, err := userModel.RowWithParams(&orm.ReqParams{
			Conditions: []*orm.ConditionItem{{Column: admins.ColumnId, Value: 4}},
		}); err != nil {
			// panic(err)
		}
	}()
	// 查询
	go func() {
		if _, err := userModel.RowWithParams(&orm.ReqParams{
			Conditions: []*orm.ConditionItem{
				{Column: admins.ColumnId, Value: 5},
				// {Column: admins.ColumnPurchaseStatus, Value: 1},
			},
		}); err != nil {
			// panic(err)
		}
	}()
	// 更新
	go func() {
		if err := userModel.UpdateWithParams(map[string]interface{}{admins.ColumnState: 1}, &orm.ReqParams{
			Conditions: []*orm.ConditionItem{{Column: admins.ColumnId, Value: 1}},
		}); err != nil {
			// panic(err)
		}
	}()
}

func main() {
	flag.Parse()
	var c Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	logx.DisableStat()
	logx.MustSetup(c.Log)
	// 初始化
	constants.ENV = c.Mode

	var (
		redisClient = redis.New(c.Mode, c.Log.Encoding, c.Redis, c.Log)
		dbClient    = orm.NewMysqlConnect(
			&orm.DBConnectConfig{
				Mode:        c.Mode,
				Log:         c.Log,
				Databases:   c.Mysql,
				RedisClient: redisClient,
			},
		)
		userModel   = admins.NewDB(dbClient)
		adminsModel = admins.NewDB(dbClient)
	)

	go dbTest(userModel, adminsModel, 1)
	go dbTest(userModel, adminsModel, 2)
	go dbTest(userModel, adminsModel, 3)
	go dbTest(userModel, adminsModel, 1)
	select {}

	// path := "/Users/yiyiyi/Code/golang/src/fio-server/logs" // 日志保存目录
	// name := "gormfile"                                      // 日志名字前缀
	// mylogger := orm2.NewStdFileLogger(
	// 	orm2.FileLogConfig{
	// 		Config: logger.Config{ // gorm日志原始配置项
	// 			SlowThreshold:             200 * time.Millisecond,
	// 			IgnoreRecordNotFoundError: false,
	// 			Colorful:                  true,
	// 		},
	// 	},
	// 	path,
	// 	name,
	// ).LogMode(orm2.Info)
	//
	// db, err := gorm.Open(
	// 	mysql.Open("root:a11111111@tcp(127.0.0.1:3306)/xmm?charset=utf8mb4&parseTime=True&loc=Local"),
	// 	&gorm.Config{
	// 		// Logger: logger.Default.LogMode(logger.Info),
	// 		Logger: mylogger,
	// 	},
	// )
	// if err != nil {
	// 	panic(err)
	// }
	//
	// var aaa = admins.NewDB(db)
	//
	// if err := aaa.Delete([]uint{1}); err != nil {
	// 	panic(err)
	// }
	//
	// if err := aaa.Update(admins.Admins{
	// 	ID:       5,
	// 	Nickname: functions.UniqueId(),
	// }); err != nil {
	// 	panic(err)
	// }
	//
	// if res, err := aaa.Row(1); err != nil {
	// 	functions.PrintStyle("red", "\n err: ", err)
	// } else {
	// 	if res != nil {
	// 		var useDBCache = false
	// 		if res.DefaultModel != nil {
	// 			useDBCache = true
	// 		}
	// 		fmt.Printf("\n useDBCache: %+v id: %+v nickname: %+v \n", useDBCache, res.ID, res.Nickname)
	// 	}
	// }
	//
	// if res, err := aaa.Row(5); err != nil {
	// 	functions.PrintStyle("red", "\n err: ", err)
	// } else {
	// 	if res != nil {
	// 		var useDBCache = false
	// 		if res.DefaultModel != nil {
	// 			useDBCache = true
	// 		}
	// 		fmt.Printf("\n useDBCache: %+v id: %+v nickname: %+v \n", useDBCache, res.ID, res.Nickname)
	// 	}
	// }
	//
	// os.Exit(1)
	//
	// if err := db.Use(
	// 	orm2.NewCachePlugin(
	// 		redis.New(
	// 			"dev",
	// 			tps.YamlRedis{
	// 				IP: "127.0.0.1:6379",
	// 				// Pass: "",
	// 				MaxIdle:     300,
	// 				MaxActive:   600,
	// 				IdleTimeout: 300,
	// 			},
	// 			logx.LogConf{
	// 				KeepDays:    30,
	// 				Compress:    true,
	// 				ServiceName: "test",
	// 				Path:        "logs/application/test",
	// 				Mode:        "console",
	// 				Encoding:    "plain",
	// 				Level:       "info",
	// 			},
	// 		),
	// 		cache.New(1*time.Minute, 5*time.Minute),
	// 	),
	// ); err != nil {
	// 	panic(err)
	// }
	//
	// var model = admins.NewDB(db)
	// for item := range time.NewTicker(time.Second * 1).C {
	// 	if item.Second()%2 == 0 {
	// 		if err := model.Update(admins.Admins{
	// 			ID:       5,
	// 			Nickname: functions.UniqueId(),
	// 		}); err != nil {
	// 			panic(err)
	// 		}
	// 	}
	//
	// 	if res, err := model.Row(1); err != nil {
	// 		functions.PrintStyle("red", "\n err: ", err)
	// 	} else {
	// 		if res != nil {
	// 			var useDBCache = false
	// 			if res.DefaultModel != nil {
	// 				useDBCache = true
	// 			}
	// 			fmt.Printf("\n useDBCache: %+v id: %+v nickname: %+v \n", useDBCache, res.ID, res.Nickname)
	// 		}
	// 	}
	//
	// 	if res, err := model.Row(5); err != nil {
	// 		functions.PrintStyle("red", "\n err: ", err)
	// 	} else {
	// 		if res != nil {
	// 			var useDBCache = false
	// 			if res.DefaultModel != nil {
	// 				useDBCache = true
	// 			}
	// 			fmt.Printf("\n useDBCache: %+v id: %+v nickname: %+v \n", useDBCache, res.ID, res.Nickname)
	// 		}
	// 	}
	// }

	// if err := model.Delete([]uint{2}); err != nil {
	// 	panic(err)
	// }

	// if err := model.Insert([]admins.Admins{
	// 	{
	// 		ID:       17,
	// 		Username: functions.UniqueId(),
	// 		Nickname: "asssss",
	// 		Email:    functions.UniqueId(),
	// 		Mobile:   "111",
	// 	},
	// }); err != nil {
	// 	panic(err)
	// }

	// if err := model.Insert([]admins.Admins{
	// 	{
	// 		ID:       17,
	// 		Username: functions.UniqueId(),
	// 		Nickname: "asssss",
	// 		Email:    functions.UniqueId(),
	// 		Mobile:   "111",
	// 	},
	// }); err != nil {
	// 	panic(err)
	// }

	// if err := model.Upsert([]admins.Admins{
	// 	{
	// 		ID:       17,
	// 		Username: functions.UniqueId(),
	// 		Nickname: "asssss11",
	// 		Email:    functions.UniqueId(),
	// 		Mobile:   "111",
	// 	},
	// }); err != nil {
	// 	panic(err)
	// }

	// if err := model.Update(admins.Admins{
	// 	ID:       17,
	// 	Username: functions.UniqueId(),
	// 	Nickname: "asssss19",
	// 	Email:    functions.UniqueId(),
	// 	Mobile:   "111",
	// }); err != nil {
	// 	panic(err)
	// }

	// for range time.NewTicker(time.Second * 1).C {
	// go func() {
	// 	if err := model.Delete([]uint{2}); err != nil {
	// 		panic(err)
	// 	}
	//
	// 	res, err := model.Row(2)
	// 	if err != nil {
	// 		functions.PrintStyle("red", "\n err: ", err)
	// 	}
	// 	if res != nil {
	// 		functions.PrintStyle("green", "\n res.ID: ", res.ID)
	// 	}
	// }()
	// go func() {
	// 	res, err := model.Row(1)
	// 	if err != nil {
	// 		functions.PrintStyle("red", "\n err: ", err)
	// 	}
	// 	if res != nil {
	// 		functions.PrintStyle("green", "\n res.ID: ", res.ID)
	// 	}
	// }()
	// go func() {
	// 	res, err := model.Row(5)
	// 	if err != nil {
	// 		functions.PrintStyle("red", "\n err: ", err)
	// 	}
	// 	if res != nil {
	// 		functions.PrintStyle("green", "\n res.ID: ", res.ID)
	// 	}
	// }()
	// go func() {
	// 	res, err := admins.NewDB(db).Row(5)
	// 	fmt.Printf("\n err: %+v \n", err)
	// 	fmt.Printf("\n res: %+v \n", res)
	// }()
	// go func() {
	// 	// time.Sleep(1 * time.Second)
	// 	res, err := admins.NewDB(db).List(&orm.FindParams{
	// 		Conditions: []*orm.ConditionItem{
	// 			{
	// 				Column: admins.ColumnIsDel,
	// 				Value:  0,
	// 			},
	// 		},
	// 	})
	// 	fmt.Printf("\n err: %+v \n", err)
	// 	fmt.Printf("\n res: %+v \n", len(res))
	// }()
	// }

	// data, err := admins.NewDB(db).Row(1)
	// jj, _ := json.Marshal(data)
	// var str bytes.Buffer
	// _ = json.Indent(&str, jj, "", "    ")
	// fmt.Printf("\n format: %+v \n", str.String())

	// list, err := admins.NewDB(db).List(
	// 	&orm.FindParams{
	// 		Page: 3,
	// 		Orders: []*orm.OrderItem{
	// 			{
	// 				Column: "id",
	// 				Value:  "desc",
	// 			},
	// 			{
	// 				Column: "username",
	// 			},
	// 		},
	// 		Conditions: []*orm.ConditionItem{
	// 			{
	// 				Column: "id",
	// 				Value:  "DDDDD",
	// 			},
	// 			{
	// 				LogicalOperator: "or",
	// 				Inner: []*orm.ConditionItem{
	// 					{
	// 						Column: "username",
	// 						Value: []string{
	// 							"AAAAAA", "bbbbb",
	// 						},
	// 						Operator: "in",
	// 					},
	// 					{
	// 						Column:          "password",
	// 						Value:           "AAAAAA",
	// 						LogicalOperator: "or",
	// 					},
	// 					{
	// 						LogicalOperator: "or",
	// 						Inner: []*orm.ConditionItem{
	// 							{
	// 								LogicalOperator: "or",
	// 								Inner: []*orm.ConditionItem{
	// 									{
	// 										Column: "username",
	// 										Value:  "AAAAAA",
	// 									},
	// 									{
	// 										Column:          "password",
	// 										Value:           "AAAAAA",
	// 										LogicalOperator: "or",
	// 									},
	// 									{
	// 										Column: "email",
	// 										Value:  "AAAAAA",
	// 									},
	// 								},
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// )
	// fmt.Printf("\n err: %+v \n", err)
	// fmt.Printf("\n list: %+v \n", list)

	// count, err := admins.NewDB(db).Count(
	// 	&orm.FindParams{
	// 		Conditions: []*orm.ConditionItem{
	// 			{
	// 				Column: "id",
	// 				Value:  "DDDDD",
	// 			},
	// 			{
	// 				LogicalOperator: "or",
	// 				Inner: []*orm.ConditionItem{
	// 					{
	// 						Column: "username",
	// 						Value: []string{
	// 							"AAAAAA", "bbbbb",
	// 						},
	// 						Operator: "in",
	// 					},
	// 					{
	// 						Column:          "password",
	// 						Value:           "AAAAAA",
	// 						LogicalOperator: "or",
	// 					},
	// 					{
	// 						LogicalOperator: "or",
	// 						Inner: []*orm.ConditionItem{
	// 							{
	// 								LogicalOperator: "or",
	// 								Inner: []*orm.ConditionItem{
	// 									{
	// 										Column: "username",
	// 										Value:  "AAAAAA",
	// 									},
	// 									{
	// 										Column:          "password",
	// 										Value:           "AAAAAA",
	// 										LogicalOperator: "or",
	// 									},
	// 									{
	// 										Column: "email",
	// 										Value:  "AAAAAA",
	// 									},
	// 								},
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// )
	// fmt.Printf("\n err: %+v \n", err)
	// fmt.Printf("\n count: %+v \n", count)
	//
	// row, err := admins.NewModel(db).FindRow(func(db models.DB) models.DB {
	// 	return db.Where("id = ?", 1)
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// jj1, _ := json.Marshal(row)
	// var str1 bytes.Buffer
	// _ = json.Indent(&str1, jj1, "", "    ")
	// fmt.Printf("\n format: %+v \n", str1.String())

	// if err := admins.NewModel(db).Upsert(
	// 	[]admins.Admins{
	// 		{
	// 			ID:       1,
	// 			Nickname: "zzzzzzzzzzzzzzzz",
	// 		},
	// 	},
	// ); err != nil {
	// 	panic(err)
	// }

	// if err := admins.NewModel(db).UpdateByColumns(
	// 	map[string]interface{}{
	// 		"nickname": "ssss",
	// 	},
	// 	func(db orm.DB) orm.DB {
	// 		return db.Where("id", 4)
	// 	},
	// ); err != nil {
	// 	panic(err)
	// }

	// if err := admins.NewDB(db).Delete([]uint{6, 999}); err != nil {
	// 	panic(err)
	// }

	// if err := admins.NewModel(db).Delete(func(db orm.DB) orm.DB {
	// 	return db.Where("id in ?", []uint{3333, 444})
	// }); err != nil {
	// 	panic(err)
	// }

	// if err := admins.NewDB(db).Delete([]uint{6, 999}); err != nil {
	// 	panic(err)
	// }

}

// // https://gorm.io/zh_CN/docs/
// func NewOrm(connect gorm.Dialector, level logger.LogLevel) *gorm.DB {
// 	db, err := gorm.Open(
// 		connect, &gorm.Config{
// 			Logger: logger.Default.LogMode(level),
// 		},
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	return db
// }
