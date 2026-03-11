package orm

import (
	"context"
	"fmt"
)

func (d *DBX[T]) Delete(ctx context.Context, dbInstance DB, deleteClosure func(DB) DB) error {
	return deleteClosure(dbInstance.Model(new(T))).WithContext(ctx).Delete(nil).Error
}

func (d *DBX[T]) DeleteWithIds(ctx context.Context, dbInstance DB, ids []any) error {
	return dbInstance.WithContext(ctx).Model(d.originalModel).Where(fmt.Sprintf("%s in ?", d.originalModel.PrimaryKey()), ids).Delete(nil).Error
}

// func (d *DBX[T]) Delete(deleteClosure func(DB) DB) error {
// 	var model T
// 	return deleteClosure(d.DB.Model(new(T))).Delete(&model).Error
// }
//
// func (d *DBX[T]) DeleteWithIds(ids []any) error {
// 	var model T
// 	return d.DB.Model(new(T)).Where(fmt.Sprintf("%s in ?", d.originalModel.PrimaryKey()), ids).Delete(&model).Error
// }
