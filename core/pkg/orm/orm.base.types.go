package orm

import (
	"time"

	cache "github.com/patrickmn/go-cache"
	"gorm.io/gorm"

	"skeyevss/core/repositories/redis"
)

const BulkUpdateTypeOrigin = 1

type (
	DB          = *gorm.DB
	MemoryCache = *cache.Cache

	CacheClientDriver struct {
		RedisClient *redis.Client
		MemoryCache MemoryCache
	}

	CacheItem struct {
		Sql  string      `json:"sql"`
		Data interface{} `json:"data"`
	}

	CacheDriver interface {
		Set(advanced *UseCacheAdvanced, drivers *CacheClientDriver, uniqueId string, data *CacheItem)
		Get(advanced *UseCacheAdvanced, drivers *CacheClientDriver, uniqueId string) []byte
		Delete(advanced *UseCacheAdvanced, drivers *CacheClientDriver, uniqueId string)
	}

	UseCacheAdvanced struct {
		// actions
		Create,
		Query,
		Delete,
		Update,
		UpdateDelete,
		Row,
		Raw bool

		CacheKeyPrefix string
		// implements CacheDriver
		Driver interface{}
		// 过期时间 单位/s
		Expire int64
	}

	DefaultModel struct {
		UseDBCache bool `gorm:"-" json:"-"`
	}

	Model interface {
		// ToMap 转map
		ToMap() map[string]interface{}
		// Columns 表字段集合
		Columns() []string
		// UniqueKeys 唯一索引集合
		UniqueKeys() []string
		// PrimaryKey 主键
		PrimaryKey() string
		// TableName 表名
		TableName() string
		// OnConflictColumns 更新冲突字段
		OnConflictColumns(_ []string) []string

		// QueryConditions 查询条件
		QueryConditions(conditions []*ConditionItem) []*ConditionItem
		// SetConditions 设置更新条件
		SetConditions(conditions []*ConditionItem) []*ConditionItem

		// UseCache 使用缓存配置
		UseCache() *UseCacheAdvanced
		// Correction 数据修正
		Correction(action ActionType) interface{}
		// CorrectionMap map数据纠正
		CorrectionMap(data map[string]interface{}) map[string]interface{}
	}

	DBX[T Model] struct {
		DB DB

		originalModel T
	}

	ActionType *string

	Foundation[T Model] struct {
		db *DBX[T]

		originalModel    T
		ctxCancelTimeout time.Duration
		dbType           string
	}

	Pagination struct {
		Limit  int
		Offset int
	}
)

type (
	OrderItem struct {
		Column string    `json:"column"`
		Value  OrderType `json:"value"`
	}

	VectorQueryItem struct {
		Column string    `json:"column"`
		Value  []float64 `json:"value"`
	}

	UpdateItem struct {
		Column string      `json:"column"`
		Value  interface{} `json:"value"`
	}

	// 批量更新主键值
	BulkUpdateInner struct {
		PK   interface{} `json:"pk"`            // 主键值
		Val  interface{} `json:"val"`           // 更新内容
		Type int64       `json:"type,optional"` // 更新方式 0 ?占位符 1 原始数据
	}

	BulkUpdateItemDef struct {
		Value interface{} `json:"value"`
		Type  int64       `json:"type,optional"` // 0 数字 1 字符串
	}

	BulkUpdateItem struct {
		Column  string             `json:"column"` // 更新字段
		Def     *BulkUpdateItemDef `json:"def"`
		Records []*BulkUpdateInner `json:"records"`
	}

	ReqParams struct {
		Columns        []string         `json:"columns,optional"`
		Orders         []*OrderItem     `json:"orders,optional"`
		Conditions     []*ConditionItem `json:"conditions,optional"`
		UniqueIds      []string         `json:"uniqueIds,optional"`
		Limit          int              `json:"limit,optional"`
		Page           int              `json:"page,optional"`
		Keyword        string           `json:"keyword,optional"`
		UniqueId       string           `json:"uniqueId,optional"`
		All            bool             `json:"all,optional"`
		Type           int64            `json:"type,optional"`
		IgnoreNotFound bool             `json:"ignoreNotFound,optional"`

		IgnoreUpdateColumns []string `json:"ignoreUpdateColumns,optional"`

		Data        []*UpdateItem          `json:"data,optional"`
		BulkUpdates []*BulkUpdateItem      `json:"bulkUpdates,optional"`
		DataRecord  map[string]interface{} `json:"-"` // data 转 record
		Backend     bool                   `json:"backend,optional"`
	}
)
