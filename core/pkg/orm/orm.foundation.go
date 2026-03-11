package orm

// 数据库基础操作
import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"skeyevss/core/pkg/functions"
)

func NewFoundation[T Model](db *gorm.DB, model T, ctxCancelTimeout time.Duration) *Foundation[T] {
	return &Foundation[T]{
		db:               NewDBX[T](db, model),
		originalModel:    model,
		ctxCancelTimeout: ctxCancelTimeout,
	}
}

func (d *Foundation[T]) ctxTrace(ctx context.Context) context.Context {
	return context.WithValue(ctx, callerFileCtxName, functions.CallerFileFull(3))
}

func (d *Foundation[T]) withDBType(dbType string) *Foundation[T] {
	d.dbType = dbType
	return d
}

func (d *Foundation[T]) GetDB() *DBX[T] {
	return d.db
}

func (d *Foundation[T]) GetDatabaseType() string {
	if d.db.DB == nil {
		if d.dbType == "" {
			panic("dbType does not set")
		}

		return d.dbType
	}

	switch d.db.DB.Dialector.Name() {
	case "mysql":
		return "mysql"

	case "sqlite":
		return "sqlite"

	case "postgres":
		return "postgres"

	case "sqlserver":
		return "sqlserver"

	default:
		return "unknown"
	}
}

// MARK ---------------------------------------------------- find row

func (d *Foundation[T]) Row(pk any) (*T, error) {
	var tx = dbSession(d.db.DB, false)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	row, err := d.db.FindWithId(d.ctxTrace(ctx), tx, pk)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (d *Foundation[T]) RowWithParams(params *ReqParams) (*T, error) {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.QueryConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(false)
	if err != nil {
		return nil, err
	}

	order, err := d.db.Order(params)
	if err != nil {
		return nil, err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	row, err := d.db.FindRow(d.ctxTrace(ctx), tx, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(where, placeholder...).Order(order).Limit(1)
	})

	if err != nil {
		return nil, err
	}

	return &row, nil
}

// MARK ---------------------------------------------------- find value

func (d *Foundation[T]) ValueWithParams(column string, params *ReqParams) (interface{}, error) {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.QueryConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(false)
	if err != nil {
		return nil, err
	}

	order, err := d.db.Order(params)
	if err != nil {
		return nil, err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.Value(d.ctxTrace(ctx), tx, column, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(where, placeholder...).Order(order).Limit(1)
	})
}

// MARK ---------------------------------------------------- find aggregate

func (d *Foundation[T]) MaxWithParams(column string, value interface{}, params *ReqParams) error {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.QueryConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(false)
	if err != nil {
		return err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	err = d.db.Max(d.ctxTrace(ctx), tx, value, column, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(where, placeholder...)
	})

	if params.IgnoreNotFound && errors.Is(err, NotFound) {
		return nil
	}

	return err
}

func (d *Foundation[T]) SumWithParams(column string, value interface{}, params *ReqParams) error {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.QueryConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(false)
	if err != nil {
		return err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	err = d.db.Sum(d.ctxTrace(ctx), tx, value, column, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(where, placeholder...)
	})

	if params.IgnoreNotFound && errors.Is(err, NotFound) {
		return nil
	}

	return err
}

// MARK ---------------------------------------------------- find exists

func (d *Foundation[T]) Exists(id uint) (bool, error) {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](
		[]*ConditionItem{
			{
				Column: d.originalModel.PrimaryKey(),
				Value:  id,
			},
		},
		d.originalModel,
		d.dbType,
	).Do(false)
	if err != nil {
		return false, err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	val, err := d.db.Count(d.ctxTrace(ctx), tx, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(where, placeholder...)
	})

	return val > 0, err
}

func (d *Foundation[T]) ExistsWithParams(params *ReqParams) (bool, error) {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.QueryConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(false)
	if err != nil {
		return false, err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	val, err := d.db.Count(d.ctxTrace(ctx), tx, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(where, placeholder...)
	})
	return val > 0, err
}

// MARK ---------------------------------------------------- find count

func (d *Foundation[T]) Count(params *ReqParams) (int64, error) {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.QueryConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(true)
	if err != nil {
		return 0, err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.Count(d.ctxTrace(ctx), tx, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(where, placeholder...)
	})
}

// MARK ---------------------------------------------------- find list

func (d *Foundation[T]) WithIds(ids []uint) ([]*T, error) {
	var tx = dbSession(d.db.DB, false)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.FindWithIds(d.ctxTrace(ctx), tx, functions.SliceToSliceAny(ids))
}

func (d *Foundation[T]) List(params *ReqParams) ([]*T, error) {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.QueryConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(true)
	if err != nil {
		return nil, err
	}

	var list []*T
	order, err := d.db.Order(params)
	if err != nil {
		return nil, err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	if err := d.db.Find(d.ctxTrace(ctx), tx, func(ctx context.Context, db DB) DB {
		if pagination := d.db.MakeLimit(params); pagination != nil {
			return db.WithContext(ctx).Where(where, placeholder...).Limit(pagination.Limit).Offset(pagination.Offset).Order(order).Find(&list)
		}
		return db.WithContext(ctx).Where(where, placeholder...).Order(order).Find(&list)
	}); err != nil {
		if errors.Is(err, NotFound) {
			return nil, nil
		}

		return nil, err
	}

	return list, nil
}

func (d *Foundation[T]) ListWithClosure(call func(ctx context.Context, db DB) DB) ([]*T, error) {
	var (
		tx   = dbSession(d.db.DB, false)
		list []*T
	)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	if err := d.db.Find(d.ctxTrace(ctx), tx, func(ctx context.Context, db DB) DB {
		return call(ctx, tx).Find(&list)
	}); err != nil {
		if errors.Is(err, NotFound) {
			return nil, nil
		}

		return nil, err
	}

	return list, nil
}

// MARK ---------------------------------------------------- find group

func (d *Foundation[T]) Group(params *ReqParams, selectColumn, groupColumn string, value interface{}) error {
	var tx = dbSession(d.db.DB, false)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.QueryConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(true)
	if err != nil {
		return err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.FindGroup(d.ctxTrace(ctx), tx, selectColumn, groupColumn, value, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(where, placeholder...)
	})
}

// MARK ---------------------------------------------------- delete

func (d *Foundation[T]) Delete(uniqueIds []string) error {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.DeleteWithIds(d.ctxTrace(ctx), tx, functions.SliceToSliceAny(uniqueIds))
}

func (d *Foundation[T]) DeleteBy(params *ReqParams) error {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	where, placeholder, err := NewConditionBuild[T](d.originalModel.SetConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(false)
	if err != nil {
		return err
	}

	order, err := d.db.Order(params)
	if err != nil {
		return err
	}

	return d.db.Delete(d.ctxTrace(ctx), tx, func(db DB) DB {
		return db.Where(where, placeholder...).Order(order)
	})
}

// MARK ---------------------------------------------------- insert

func (d *Foundation[T]) Insert(records []T) error {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.Insert(d.ctxTrace(ctx), tx, records)
}

func (d *Foundation[T]) Add(record T) (*T, error) {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.Add(d.ctxTrace(ctx), tx, record)
}

// MARK ---------------------------------------------------- update

func (d *Foundation[T]) Upsert(records []T, onConflictColumns []string) error {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.Upsert(d.ctxTrace(ctx), tx, records, onConflictColumns)
}

func (d *Foundation[T]) UpsertWithExcludeColumns(records []T, onConflictColumns []string, excludeColumns []string) error {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.UpsertWithExcludeColumns(d.ctxTrace(ctx), tx, records, onConflictColumns, excludeColumns)
}

func (d *Foundation[T]) Update(record T) error {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.Update(d.ctxTrace(ctx), tx, record)
}

func (d *Foundation[T]) UpdateWithColumns(pkColumn string, pkValue any, maps map[string]interface{}, params *ReqParams) error {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.UpdateByColumns(d.ctxTrace(ctx), tx, params, maps, func(ctx context.Context, db DB) DB {
		return db.Where(pkColumn, pkValue)
	})
}

func (d *Foundation[T]) UpdateWithParams(maps map[string]interface{}, params *ReqParams) error {
	var tx = dbSession(d.db.DB, true)
	where, placeholder, err := NewConditionBuild[T](d.originalModel.SetConditions(params.Conditions), d.originalModel, d.GetDatabaseType()).Do(false)
	if err != nil {
		return err
	}

	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.UpdateByColumns(d.ctxTrace(ctx), tx, params, maps, func(ctx context.Context, db DB) DB {
		return db.Where(where, placeholder...)
	})
}

func (d *Foundation[T]) BulkUpdate(primaryKey string, updateAllowedColumns []string, records []*BulkUpdateItem) error {
	var tx = dbSession(d.db.DB, true)
	ctx, cancel := makeContext(tx, d.ctxCancelTimeout)
	defer cancel()

	return d.db.BulkUpdate(ctx, tx, primaryKey, updateAllowedColumns, records)
}

// MARK ---------------------------------------------------- context

func makeContext(db DB, timeout time.Duration) (context.Context, context.CancelFunc) {
	var ctx = context.Background()
	if dbCtx, ok := db.Get("context"); ok {
		if ctxVal, ok := dbCtx.(context.Context); ok {
			ctx = ctxVal
		}
	}

	return context.WithTimeout(ctx, timeout)
}

// MARK ---------------------------------------------------- 事务

func Transaction(db DB, options ...func(db DB) error) error {
	var tx = dbSession(db, true)
	ctx, cancel := makeContext(tx, 5*time.Second)
	defer cancel()

	return transaction(context.WithValue(ctx, callerFileCtxName, functions.CallerFileFull(3)), tx, options...)
}

func transaction(ctx context.Context, db DB, options ...func(db DB) error) error {
	if _, ok := db.Get("gorm:started_transaction"); ok {
		return errors.New("nested transactions are not allowed")
	}

	var tx = db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var committed bool
	defer func() {
		if !committed {
			if err := tx.Rollback().Error; err != nil {
				functions.LogError("Transaction rollback failed: %v", err)
			}
		}
	}()

	for _, fn := range options {
		if err := fn(tx); err != nil {
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	committed = true
	return nil
}

// MARK ---------------------------------------------------- session

func dbSession(db DB, writeState bool) DB {
	// DisableNestedTransaction: true, // 写操作设为true，纯查询可false
	// SkipDefaultTransaction:   true, // 写操作false，批量操作true 不设置SkipDefaultTransaction让gorm自动管理
	// NewDB:                    true,
	// PrepareStmt:              true,
	// Initialized:           	 true,
	// QueryFields: true, // 明确查询字段
	// DryRun:      true, // 许在不实际执行数据库操作的情况下查看GORM将要执行的SQL语句 Statement.SQL, Statement.Vars
	if writeState {
		return db.Session(
			&gorm.Session{
				DisableNestedTransaction: true,

				NewDB:       true,
				PrepareStmt: true,
			},
		)
	}

	return db.Session(
		&gorm.Session{
			DisableNestedTransaction: false,
			QueryFields:              true,

			NewDB:       true,
			PrepareStmt: true,
		},
	)
}

// MARK ---------------------------------------------------- conditions

// 匹配数字
func (d *Foundation[T]) CaseNumberCondition(column string) *ConditionOriginalItem {
	if d.GetDatabaseType() == DBTypeSqlite {
		return &ConditionOriginalItem{
			Query:  fmt.Sprintf("%s <> '' AND %s NOT GLOB ?", column, column),
			Values: []interface{}{"^[0-9]+$"},
		}
	}

	return &ConditionOriginalItem{
		Query:  fmt.Sprintf("%s REGEXP ?", column),
		Values: []interface{}{"^[0-9]+$"},
	}
}

// 字符串指定索引长度匹配
func (d *Foundation[T]) CaseSubstrCondition(column string, start, end int, subs []interface{}) *ConditionOriginalItem {
	if len(subs) <= 0 {
		return nil
	}

	var (
		length       = len(subs)
		placeholders = strings.Trim(strings.Repeat("?,", length), ",")
	)
	if d.GetDatabaseType() == DBTypeSqlite {
		var (
			item       = fmt.Sprintf("SUBSTR(`%s`, %d, %d) = ?", strings.Trim(column, "`"), start, end)
			conditions []string
		)
		for i := 0; i < length; i++ {
			conditions = append(conditions, item)
		}
		return &ConditionOriginalItem{
			Query:  fmt.Sprintf("(%s)", strings.Join(conditions, "OR")),
			Values: subs,
		}
	}

	return &ConditionOriginalItem{
		Query:  fmt.Sprintf("SUBSTR(`%s`, %d, %d) IN (%s)", strings.Trim(column, "`"), start, end, placeholders),
		Values: subs,
	}
}

// 匹配字段是否包含数组中的元素
func (d *Foundation[T]) CaseJSONContainsCondition(column string, data []interface{}) *ConditionOriginalItem {
	if len(data) <= 0 {
		return nil
	}

	if d.GetDatabaseType() == DBTypeSqlite {
		return &ConditionOriginalItem{
			Query: fmt.Sprintf(
				"EXISTS (SELECT 1 FROM json_each(%s) WHERE value IN (%s))",
				column,
				strings.Trim(strings.Repeat("?,", len(data)), ","),
			),
			Values: data,
		}
	}

	var conditions []string
	for range data {
		conditions = append(conditions, fmt.Sprintf("? MEMBER OF(%s)", column))
	}

	return &ConditionOriginalItem{
		Query:  fmt.Sprintf("(%s)", strings.Join(conditions, " OR ")),
		Values: data,
	}
}

// json array 全包含
func (d *Foundation[T]) CaseJSONContainsAllArrCondition(column string, arr []interface{}) *ConditionOriginalItem {
	/*
		Original: &orm.ConditionOriginalItem{
			Query:  fmt.Sprintf("JSON_CONTAINS(`%s`, ?, '$')", fodItems.ColumnMergedIds),
			Values: []interface{}{fmt.Sprintf("[%d]", row.ID)},
		}
	*/
	if len(arr) <= 0 {
		return nil
	}

	var data = &ConditionOriginalItem{
		Query: fmt.Sprintf("JSON_CONTAINS(`%s`, ?, '$')", column),
	}
	var (
		placeholder []string
		res         []interface{}
	)
	for _, item := range arr {
		if _, ok := item.(string); ok {
			placeholder = append(placeholder, "'%s'")
		} else {
			placeholder = append(placeholder, "%v")
		}

		res = append(res, item)
	}
	data.Values = []interface{}{fmt.Sprintf("["+strings.Join(placeholder, " ,")+"]", res...)}

	return data
}
