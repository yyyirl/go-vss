package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm/clause"

	"skeyevss/core/pkg/functions"
)

func (d *DBX[T]) Update(ctx context.Context, dbInstance DB, record T) error {
	return dbInstance.WithContext(ctx).Updates(record.Correction(ActionUpdate)).Error
}

func (d *DBX[T]) UpdateByColumns(ctx context.Context, dbInstance DB, params *ReqParams, record map[string]interface{}, findClosure func(context.Context, DB) DB) error {
	var data = d.originalModel.CorrectionMap(record)
	if params != nil && len(params.IgnoreUpdateColumns) > 0 {
		var columns []string
		for column := range data {
			if functions.Contains(column, params.IgnoreUpdateColumns) {
				continue
			}
			columns = append(columns, column)
		}

		return findClosure(ctx, dbInstance.Model(new(T))).Select(columns).Updates(data).Error
	}

	return findClosure(ctx, dbInstance.Model(new(T))).Updates(data).Error
}

func (d *DBX[T]) Upsert(ctx context.Context, dbInstance DB, records []T, onConflictColumns []string) error {
	if columns := d.originalModel.OnConflictColumns(onConflictColumns); len(columns) > 0 {
		return dbInstance.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(columns),
		}).Create(d.corrections(ActionInsert, records)).Error
	}

	return dbInstance.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns(d.updateColumns()),
	}).Create(d.corrections(ActionInsert, records)).Error
}

func (d *DBX[T]) UpsertWithExcludeColumns(ctx context.Context, dbInstance DB, records []T, onConflictColumns []string, excludeColumns []string) error {
	if columns := d.originalModel.OnConflictColumns(onConflictColumns); len(columns) > 0 {
		var db = dbInstance.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(
				functions.ArrFilter(columns, func(item string) bool {
					return !functions.Contains(item, excludeColumns)
				}),
			),
		})
		if excludeColumns != nil {
			db = db.Omit(excludeColumns...)
		}

		return db.Create(d.corrections(ActionInsert, records)).Error
	}

	var db = dbInstance.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.AssignmentColumns(
			functions.ArrFilter(d.updateColumns(), func(item string) bool {
				return !functions.Contains(item, excludeColumns)
			}),
		),
	})
	if excludeColumns != nil {
		db = db.Omit(excludeColumns...)
	}

	return db.Create(d.corrections(ActionInsert, records)).Error
}

// MARK ---------------------------------------- 批量更新

func (d *DBX[T]) BulkUpdate(ctx context.Context, dbInstance DB, primaryKey string, updateAllowedColumns []string, records []*BulkUpdateItem) error {
	_sql, pretreatments, err := d.takeUpdateBulk(primaryKey, updateAllowedColumns, records)
	if err != nil {
		return err
	}

	return dbInstance.WithContext(ctx).Exec(_sql, pretreatments...).Error
}

func (d *DBX[T]) modelIsSqlNull(v interface{}) (interface{}, bool) {
	if val, ok := v.(sql.NullString); ok {
		return val, true
	} else if val, ok := v.(sql.NullByte); ok {
		return val, true
	} else if val, ok := v.(sql.NullBool); ok {
		return val, true
	} else if val, ok := v.(sql.NullFloat64); ok {
		return val, true
	} else if val, ok := v.(sql.NullInt16); ok {
		return val, true
	} else if val, ok := v.(sql.NullInt32); ok {
		return val, true
	} else if val, ok := v.(sql.NullInt64); ok {
		return val, true
	} else if val, ok := v.(sql.NullTime); ok {
		return val, true
	}

	return v, false
}

func (d *DBX[T]) takeUpdateBulk(primaryKey string, updateAllowedColumns []string, records []*BulkUpdateItem) (string, []interface{}, error) {
	primaryKey = strings.Trim(primaryKey, "`")
	if primaryKey == "" {
		return "", nil, errors.New("update primary key 不能为空")
	}

	if len(updateAllowedColumns) <= 0 {
		return "", nil, errors.New("update bulk约束字段不能为空")
	}

	if len(records) <= 0 {
		return "", nil, errors.New("update bulk字段`data`不能为空")
	}

	var (
		arr           []string
		whereValues   []interface{}
		pretreatments []interface{}
	)
	for _, val := range records {
		var (
			length = len(val.Records)
			when   = make([]string, 0, length)
		)

		for _, item := range val.Records {
			if item.PK == 0 {
				continue
			}

			whereValues = append(whereValues, item.PK)
			if item.Type == BulkUpdateTypeOrigin {
				if v, ok := item.Val.(string); ok {
					when = append(when, fmt.Sprintf("WHEN ? then %s", v))
					pretreatments = append(pretreatments, item.PK)
					continue
				}
			}

			when = append(when, "WHEN ? then ?")

			if ok, _ := functions.IsSimpleType(item.Val); ok {
				pretreatments = append(pretreatments, item.PK, item.Val)
				continue
			}

			if _v, ok := d.modelIsSqlNull(item.Val); ok {
				pretreatments = append(pretreatments, item.PK, _v)
				continue
			}

			if functions.IsMap(item.Val) || functions.IsPtrStruct(item.Val) || functions.IsStruct(item.Val) || functions.IsSlice(item.Val) {
				val, err := functions.ToString(item.Val)
				if err != nil {
					return "", nil, errors.New("BulkUpdate value序列化失败")
				}

				pretreatments = append(pretreatments, item.PK, val)
				continue
			}

			return "", nil, fmt.Errorf("BulkUpdate字段`%s`未知类型: `%T`, value: %+v", item.Val, item.PK, item.Val)
		}

		if !functions.Contains(val.Column, updateAllowedColumns) {
			return "", nil, errors.New("update bulk 字段`" + val.Column + "`非法")
		}

		if val.Def != nil {
			if val.Def.Type == 0 {
				arr = append(arr, "`"+val.Column+"` = CASE `"+primaryKey+"` "+strings.Join(when, " ")+fmt.Sprintf(" ELSE %v END", val.Def.Value))
			} else {
				arr = append(arr, "`"+val.Column+"` = CASE `"+primaryKey+"` "+strings.Join(when, " ")+fmt.Sprintf(" ELSE '%v' END", val.Def.Value))
			}
		} else {
			arr = append(arr, "`"+val.Column+"` = CASE `"+primaryKey+"` "+strings.Join(when, " ")+" ELSE '' END")
		}
	}

	var _sql = "UPDATE `" + d.originalModel.TableName() +
		"` SET " + strings.Join(arr, ", ") +
		" where `" + primaryKey + "` in(" + strings.Trim(strings.Repeat("?, ", len(whereValues)), ", ") + ")"
	pretreatments = append(pretreatments, whereValues...)
	return _sql, pretreatments, nil
}
