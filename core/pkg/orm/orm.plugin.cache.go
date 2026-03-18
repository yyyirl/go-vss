package orm

// 数据库缓存
import (
	"context"
	"errors"
	"reflect"
	"strings"

	cache "github.com/patrickmn/go-cache"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/repositories/redis"
)

var (
	_ gorm.Plugin = &CachePlugin{}

	hitCacheError = errors.New("hit cache")
)

const (
	registerNameBeforeCreate = "beforeCreate"
	registerNameBeforeQuery  = "beforeQuery"
	registerNameBeforeDelete = "beforeDelete"
	registerNameBeforeUpdate = "beforeUpdate"
	registerNameBeforeRow    = "beforeRow"
	registerNameBeforeRaw    = "beforeRaw"
	registerNameAfterCreate  = "afterCreate"
	registerNameAfterQuery   = "afterQuery"
	registerNameAfterDelete  = "afterDelete"
	registerNameAfterUpdate  = "afterUpdate"
	registerNameAfterRow     = "afterRow"
	registerNameAfterRaw     = "afterRaw"

	cacheContextCacheExists = "cacheExists"
)

type CachePlugin struct {
	Drivers *CacheClientDriver
}

func NewCachePlugin(redisClient *redis.Client, memoryCacheClient *cache.Cache) *CachePlugin {
	return &CachePlugin{
		Drivers: &CacheClientDriver{
			RedisClient: redisClient,
			MemoryCache: memoryCacheClient,
		},
	}
}

func (c *CachePlugin) Name() string {
	return "cachePlugin"
}

func (c *CachePlugin) Initialize(db *gorm.DB) error {
	if err := db.Callback().Create().Before("gorm:before_create").Register(registerNameBeforeCreate, c.beforeCreate); err != nil {
		return err
	}

	if err := db.Callback().Query().Before("gorm:query").Register(registerNameBeforeQuery, c.beforeQuery); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("gorm:before_delete").Register(registerNameBeforeDelete, c.beforeDelete); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("gorm:setup_reflect_value").Register(registerNameBeforeUpdate, c.beforeUpdate); err != nil {
		return err
	}

	if err := db.Callback().Row().Before("gorm:row").Register(registerNameBeforeRow, c.beforeRow); err != nil {
		return err
	}

	if err := db.Callback().Raw().Before("gorm:raw").Register(registerNameBeforeRaw, c.beforeRaw); err != nil {
		return err
	}

	if err := db.Callback().Create().After("gorm:after_create").Register(registerNameAfterCreate, c.afterCreate); err != nil {
		return err
	}

	if err := db.Callback().Query().After("gorm:after_query").Register(registerNameAfterQuery, c.afterQuery); err != nil {
		return err
	}

	if err := db.Callback().Delete().After("gorm:after_delete").Register(registerNameAfterDelete, c.afterDelete); err != nil {
		return err
	}

	if err := db.Callback().Update().After("gorm:after_update").Register(registerNameAfterUpdate, c.afterUpdate); err != nil {
		return err
	}

	if err := db.Callback().Row().After("gorm:row").Register(registerNameAfterRow, c.afterRow); err != nil {
		return err
	}

	if err := db.Callback().Raw().After("gorm:raw").Register(registerNameAfterRaw, c.afterRaw); err != nil {
		return err
	}

	return nil
}

func (c *CachePlugin) beforeCreate(db *gorm.DB) {
	c.before(registerNameBeforeCreate, db)
}

func (c *CachePlugin) beforeQuery(db *gorm.DB) {
	c.before(registerNameBeforeQuery, db)
}

func (c *CachePlugin) beforeDelete(db *gorm.DB) {
	c.before(registerNameBeforeDelete, db)
}

func (c *CachePlugin) beforeUpdate(db *gorm.DB) {
	c.before(registerNameBeforeUpdate, db)
}

func (c *CachePlugin) beforeRow(db *gorm.DB) {
	c.before(registerNameBeforeRow, db)
}

func (c *CachePlugin) beforeRaw(db *gorm.DB) {
	c.before(registerNameBeforeRaw, db)
}

func (c *CachePlugin) afterCreate(db *gorm.DB) {
	c.after(registerNameAfterCreate, db)
}

func (c *CachePlugin) afterQuery(db *gorm.DB) {
	c.after(registerNameAfterQuery, db)
}

func (c *CachePlugin) afterDelete(db *gorm.DB) {
	c.after(registerNameAfterDelete, db)
}

func (c *CachePlugin) afterUpdate(db *gorm.DB) {
	c.after(registerNameAfterUpdate, db)
}

func (c *CachePlugin) afterRow(db *gorm.DB) {
	c.after(registerNameAfterRow, db)
}

func (c *CachePlugin) afterRaw(db *gorm.DB) {
	c.after(registerNameAfterRaw, db)
}

func (c *CachePlugin) before(name string, db *gorm.DB) {
	model, ok := (db.Statement.Model).(Model)
	if !ok {
		return
	}

	var cacheAdvance = model.UseCache()
	if cacheAdvance == nil {
		return
	}

	switch name {
	// case registerNameBeforeCreate:
	// 	if !cacheAdvance.Create || cacheAdvance.Driver == nil {
	// 		return
	// 	}

	// case registerNameBeforeDelete:
	// 	if !cacheAdvance.Delete || cacheAdvance.Driver == nil {
	// 		return
	// 	}
	//
	// case registerNameBeforeUpdate:
	// 	if !cacheAdvance.Update || cacheAdvance.Driver == nil {
	// 		return
	// 	}

	// case registerNameBeforeRaw:
	// 	if !cacheAdvance.Raw || cacheAdvance.Driver == nil {
	// 		return
	// 	}

	case registerNameBeforeRow:
		if !cacheAdvance.Row || cacheAdvance.Driver == nil {
			return
		}

	case registerNameBeforeQuery:
		if !cacheAdvance.Query || cacheAdvance.Driver == nil {
			return
		}

	default:
		return
	}

	// 获取缓存
	if driver, ok := (cacheAdvance.Driver).(CacheDriver); ok {
		var res = driver.Get(cacheAdvance, c.Drivers, c.getCacheId(db))
		if len(res) <= 0 {
			return
		}

		if err := functions.JSONUnmarshal(res, &db.Statement.Dest); err != nil {
			functions.LogError(err)
			return
		}

		_ = functions.ModifyField(db.Statement.Dest, "DefaultModel", &DefaultModel{
			UseDBCache: true,
		})

		var ref = reflect.TypeOf(db.Statement.Dest)
		if ref.Kind() == reflect.Array {
			db.Statement.RowsAffected = int64(ref.Len())
		} else {
			db.Statement.RowsAffected = 1
		}

		db.RowsAffected = db.Statement.RowsAffected
		db.Statement.Context = context.WithValue(db.Statement.Context, cacheContextCacheExists, true)
		// 阻止数据库查询
		db.Statement.Error = hitCacheError
	}
}

// 只有查询语句会被缓存
// 设置 update delete updateDelete为true会删除 缓存前缀下的所有缓存
func (c *CachePlugin) after(name string, db *gorm.DB) {
	if errors.Is(db.Statement.Error, hitCacheError) {
		// 恢复
		db.Statement.Error = nil
	}

	if name != registerNameAfterDelete && name != registerNameAfterUpdate && name != registerNameAfterRaw {
		var cacheExists = db.Statement.Context.Value(cacheContextCacheExists)
		// 缓存已存在
		if v, ok := cacheExists.(bool); ok && v {
			return
		}

		if db.RowsAffected <= 0 {
			return
		}
	}

	var (
		refVal       = reflect.ValueOf(db.Statement.Model)
		model  Model = nil
		ok           = false
	)
	if refVal.Kind() == reflect.Array || (refVal.Kind() == reflect.Slice && refVal.Len() > 0) {
		model, ok = (refVal.Index(0).Interface()).(Model)
		if !ok {
			return
		}
	} else {
		model, ok = (db.Statement.Model).(Model)
		if !ok {
			return
		}
	}

	var cacheAdvance = model.UseCache()
	if cacheAdvance == nil {
		return
	}

	var (
		setCache    = false
		deleteCache = false
	)
	switch name {
	case registerNameAfterCreate:
		if !cacheAdvance.Create && !cacheAdvance.UpdateDelete || cacheAdvance.Driver == nil {
			return
		}
		deleteCache = true

	case registerNameAfterDelete:
		if !cacheAdvance.Delete && !cacheAdvance.UpdateDelete || cacheAdvance.Driver == nil {
			return
		}
		deleteCache = true

	case registerNameAfterUpdate:
		if !cacheAdvance.Update && !cacheAdvance.UpdateDelete || cacheAdvance.Driver == nil {
			return
		}
		deleteCache = true

	case registerNameAfterRaw:
		if !cacheAdvance.Raw && !cacheAdvance.UpdateDelete || cacheAdvance.Driver == nil {
			return
		}
		deleteCache = true

	case registerNameAfterQuery:
		if !cacheAdvance.Query || cacheAdvance.Driver == nil {
			return
		}
		setCache = true

	case registerNameAfterRow:
		if !cacheAdvance.Row || cacheAdvance.Driver == nil {
			return
		}
		setCache = true
	}

	if driver, ok := (cacheAdvance.Driver).(CacheDriver); ok {
		data, uniqueId, table := c.getCacheItem(db)
		// 设置缓存
		if setCache {
			driver.Set(cacheAdvance, c.Drivers, uniqueId, data)
		}
		// 删除缓存
		if deleteCache {
			c.deleteCacheByTable(cacheAdvance, table)
		}
	}
}

func (c *CachePlugin) getCacheItem(db *gorm.DB) (item *CacheItem, uniqueId string, table string) {
	callbacks.BuildQuerySQL(db)

	var (
		sql       = db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
		tableName = c.getTableName(db)
	)
	item = &CacheItem{
		Sql:  sql,
		Data: db.Statement.Dest,
	}

	// key 结构: prefix:table:md5(sql):data|sql
	table = tableName
	uniqueId = tableName + ":" + functions.Md5String(sql)

	return
}

func (c *CachePlugin) getCacheId(db *gorm.DB) string {
	callbacks.BuildQuerySQL(db)
	var (
		sql       = db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
		tableName = c.getTableName(db)
	)

	return tableName + ":" + functions.Md5String(sql)
}

func (c *CachePlugin) getTableName(db *gorm.DB) string {
	if db.Statement != nil && db.Statement.Table != "" {
		return db.Statement.Table
	}

	if db.Statement != nil && db.Statement.Model != nil {
		if model, ok := (db.Statement.Model).(Model); ok {
			return model.TableName()
		}
	}

	return "unknown"
}

func (c *CachePlugin) deleteCacheByTable(advanced *UseCacheAdvanced, table string) {
	if advanced == nil || advanced.CacheKeyPrefix == "" || table == "" {
		return
	}

	var prefix = advanced.CacheKeyPrefix + ":" + table + ":"

	if c.Drivers != nil && c.Drivers.RedisClient != nil {
		// 这里使用 Keys 做“按表清空”，简单但粗暴。适用于缓存 key 数量可控的场景。
		var keys, err = c.Drivers.RedisClient.Keys(prefix + "*")
		if err == nil && len(keys) > 0 {
			_, _ = c.Drivers.RedisClient.Del(keys...)
		}
	}

	if c.Drivers != nil && c.Drivers.MemoryCache != nil {
		for key := range c.Drivers.MemoryCache.Items() {
			if strings.HasPrefix(key, prefix) {
				c.Drivers.MemoryCache.Delete(key)
			}
		}
	}
}
