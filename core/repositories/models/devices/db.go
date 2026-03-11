package devices

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
	*orm.Foundation[Devices]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Devices](db, Devices{}, 5*time.Second)}
}

func (db *DB) XList(ctx context.Context, accessProtocols []uint64) ([]*XListItem, error) {
	var (
		list []*XListItem
		data = NewXList()
	)
	if len(accessProtocols) > 0 {
		if err := db.GetDB().DB.WithContext(ctx).Table(new(Devices).TableName()).Where(
			fmt.Sprintf("%s in ?", ColumnAccessProtocol), functions.SliceToSliceAny(accessProtocols),
		).Select(data.columns()).Scan(&list).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.GetDB().DB.WithContext(ctx).Table(new(Devices).TableName()).Select(data.columns()).Scan(&list).Error; err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (db *DB) OnlineStateList(ctx context.Context, req *orm.ReqParams) ([]*OnlineStateListItem, error) {
	var (
		list          []*OnlineStateListItem
		data          = NewOnlineStateList()
		originalModel = Devices{}
	)
	where, placeholder, err := orm.NewConditionBuild[Devices](originalModel.QueryConditions(req.Conditions), originalModel, db.GetDatabaseType()).Do(true)
	if err != nil {
		return nil, err
	}

	if where == "" {
		if err := db.GetDB().DB.WithContext(ctx).Table(new(Devices).TableName()).Select(data.columns()).Scan(&list).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.GetDB().DB.WithContext(ctx).Table(new(Devices).TableName()).Where(
			where, placeholder...,
		).Select(data.columns()).Scan(&list).Error; err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (db *DB) SList(ctx context.Context, req *orm.ReqParams) ([]*SimpleItem, error) {
	var (
		list          []*SimpleItem
		data          = NewSList()
		originalModel = Devices{}
	)
	where, placeholder, err := orm.NewConditionBuild[Devices](originalModel.QueryConditions(req.Conditions), originalModel, db.GetDatabaseType()).Do(true)
	if err != nil {
		return nil, err
	}

	if where == "" {
		if err := db.GetDB().DB.WithContext(ctx).Table(new(Devices).TableName()).Select(data.columns()).Scan(&list).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.GetDB().DB.WithContext(ctx).Table(new(Devices).TableName()).Where(
			where, placeholder...,
		).Select(data.columns()).Scan(&list).Error; err != nil {
			return nil, err
		}
	}

	return list, nil
}

func (db *DB) MSMaps(ctx context.Context, req *orm.ReqParams) (map[string][]uint64, error) {
	var (
		list          []*MSimpleItem
		data          = NewMSList()
		originalModel = Devices{}
	)
	where, placeholder, err := orm.NewConditionBuild[Devices](originalModel.QueryConditions(req.Conditions), originalModel, db.GetDatabaseType()).Do(true)
	if err != nil {
		return nil, err
	}

	if where == "" {
		return nil, errors.New("conditions is empty")

	}

	if err := db.GetDB().DB.WithContext(ctx).Table(new(Devices).TableName()).Where(
		where, placeholder...,
	).Select(data.columns()).Scan(&list).Error; err != nil {
		return nil, err
	}

	var maps = make(map[string][]uint64)
	for _, item := range list {
		var msIds []uint64
		if err := functions.JSONUnmarshal([]byte(item.MSIds), &msIds); err != nil {
			return nil, err
		}

		maps[item.DeviceUniqueId] = msIds
	}

	return maps, nil
}

func (db *DB) GroupByAccessProtocol() (map[uint]uint, error) {
	var list []*AccessProtocolGroup
	if err := db.Group(
		new(orm.ReqParams),
		NewAccessProtocol().columns(),
		ColumnAccessProtocol,
		&list,
	); err != nil {
		return nil, err
	}

	var maps = make(map[uint]uint)
	for _, item := range list {
		if item.Cnt > 0 {
			maps[item.AccessProtocol] = item.Cnt
		}
	}

	return maps, nil
}
