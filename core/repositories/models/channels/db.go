package channels

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[Channels]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Channels](db, Channels{}, 5*time.Second)}
}

func (db *DB) XList(ctx context.Context) ([]*XListItem, error) {
	var (
		list []*XListItem
		data = NewXList()
	)
	if err := db.GetDB().DB.WithContext(ctx).Table(new(Channels).TableName()).Select(data.columns()).Scan(&list).Error; err != nil {
		return nil, err
	}

	for _, item := range list {
		var depIds []uint64
		if item.TDepIds == "" {
			depIds = []uint64{}
		} else {
			if err := functions.ConvStringToType(item.TDepIds, &depIds); err != nil {
				return nil, err
			}
		}

		item.DepIds = depIds
	}

	return list, nil
}

func (db *DB) DeleteWithChannelFilters(ctx context.Context, deviceUniqueId string, filters []string) error {
	var conditionItem = db.CaseSubstrCondition(
		ColumnUniqueId,
		11,
		3,
		functions.SliceToSliceAny(filters),
	)
	if conditionItem == nil {
		return nil
	}

	return db.GetDB().DB.WithContext(ctx).Table(new(Channels).TableName()).Where(
		fmt.Sprintf(
			"%s = ? AND LENGTH(%s) >= 20 AND %s",
			ColumnDeviceUniqueId,
			ColumnUniqueId,
			conditionItem.Query,
		),
		append([]interface{}{deviceUniqueId}, conditionItem.Values...)...,
	).Delete(nil).Error
}

func (db *DB) XList1(ctx context.Context, req *orm.ReqParams) ([]*XListItem1, error) {
	var (
		list          []*XListItem1
		data          = NewXList1()
		originalModel = Channels{}
	)

	where, placeholder, err := orm.NewConditionBuild[Channels](originalModel.QueryConditions(req.Conditions), originalModel, db.GetDatabaseType()).Do(true)
	if err != nil {
		return nil, err
	}

	if where == "" {
		return nil, errors.New("conditions is empty")
	}

	if err := db.GetDB().DB.WithContext(ctx).Table(new(Channels).TableName()).Where(
		where, placeholder...,
	).Select(data.columns()).Scan(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

func (db *DB) OnlineStateList(ctx context.Context, req *orm.ReqParams) ([]*OnlineStateListItem, error) {
	var (
		list          []*OnlineStateListItem
		data          = NewOnlineStateList()
		originalModel = Channels{}
	)

	where, placeholder, err := orm.NewConditionBuild[Channels](originalModel.QueryConditions(req.Conditions), originalModel, db.GetDatabaseType()).Do(true)
	if err != nil {
		return nil, err
	}

	if where == "" {
		if err := db.GetDB().DB.WithContext(ctx).Table(new(Channels).TableName()).Select(data.columns()).Scan(&list).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.GetDB().DB.WithContext(ctx).Table(new(Channels).TableName()).Where(
			where, placeholder...,
		).Select(data.columns()).Scan(&list).Error; err != nil {
			return nil, err
		}
	}

	return list, nil
}
