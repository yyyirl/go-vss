package orm

import (
	"context"
	"errors"

	"skeyevss/core/pkg/functions"
)

func (d *DBX[T]) Find(ctx context.Context, dbInstance DB, findClosure func(context.Context, DB) DB) error {
	var res = findClosure(ctx, dbInstance.Model(new(T)))
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected <= 0 {
		return NotFound
	}

	return nil
}

func (d *DBX[T]) FindWithIds(ctx context.Context, dbInstance DB, ids []any) ([]*T, error) {
	if len(ids) <= 0 {
		return make([]*T, 0), nil
	}

	var (
		model  T
		pk     = model.PrimaryKey()
		models = make([]*T, 0)
	)
	if pk == "" {
		return nil, errors.New("PrimaryKey 不能为空")
	}

	if err := d.Find(ctx, dbInstance, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(pk+" IN ?", ids).Find(&models)
	}); err != nil {
		if errors.Is(err, NotFound) {
			return make([]*T, 0), nil
		}

		return nil, err
	}

	return models, nil
}

// ---------------------------------------------------- row

func (d *DBX[T]) FindWithId(ctx context.Context, dbInstance DB, id any) (T, error) {
	var model T
	if err := d.Find(ctx, dbInstance, func(ctx context.Context, db DB) DB {
		return db.WithContext(ctx).Where(model.PrimaryKey(), id).Limit(1).Find(&model)
	}); err != nil {
		return model, err
	}

	return model, nil
}

func (d *DBX[T]) FindRow(ctx context.Context, dbInstance DB, findClosure func(context.Context, DB) DB) (T, error) {
	var (
		model T
		res   = findClosure(ctx, dbInstance.Model(new(T))).Take(&model)
	)
	if res.Error != nil {
		return model, res.Error
	}

	if res.RowsAffected <= 0 {
		return model, NotFound
	}

	return model, nil
}

// ---------------------------------------------------- value

func (d *DBX[T]) Value(ctx context.Context, dbInstance DB, column string, findClosure func(context.Context, DB) DB) (interface{}, error) {
	var (
		model T
		res   = findClosure(ctx, dbInstance.Model(new(T))).Select(column).Take(&model)
	)
	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected <= 0 {
		return nil, NotFound
	}

	return functions.GetFieldValue(model, functions.Capitalize(column))
}

// ---------------------------------------------------- count

func (d *DBX[T]) Count(ctx context.Context, dbInstance DB, findClosure func(context.Context, DB) DB) (int64, error) {
	var (
		count int64
		res   = findClosure(ctx, dbInstance.Model(new(T))).Count(&count)
	)
	if res.Error != nil {
		return 0, res.Error
	}

	return count, nil
}

// ---------------------------------------------------- 聚合

func (d *DBX[T]) aggregate(ctx context.Context, dbInstance DB, function string, column string, value interface{}, findClosure func(context.Context, DB) DB) error {
	var res = findClosure(ctx, dbInstance.Model(new(T))).Select("COALESCE(" + function + "(" + column + "), 0)").Scan(&value)
	if res.Error != nil {
		if errors.Is(res.Error, NotFound) {
			return nil
		}

		return res.Error
	}

	return nil
}

func (d *DBX[T]) Max(ctx context.Context, dbInstance DB, value interface{}, column string, findClosure func(context.Context, DB) DB) error {
	return d.aggregate(ctx, dbInstance, "max", column, value, findClosure)
}

func (d *DBX[T]) Sum(ctx context.Context, dbInstance DB, value interface{}, column string, findClosure func(context.Context, DB) DB) error {
	return d.aggregate(ctx, dbInstance, "sum", column, value, findClosure)
}

// ---------------------------------------------------- group

func (d *DBX[T]) FindGroup(ctx context.Context, dbInstance DB, selectColumn, groupColumn string, value interface{}, findClosure func(context.Context, DB) DB) error {
	return findClosure(ctx, dbInstance.Model(new(T))).Select(selectColumn).Group(groupColumn).Find(value).Error
}
