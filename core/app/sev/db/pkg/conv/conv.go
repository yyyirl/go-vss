package conv

import (
	"errors"
	"strings"

	"skeyevss/core/app/sev/db/db"
	"skeyevss/core/pkg/functions"
	"skeyevss/core/pkg/orm"
	"skeyevss/core/tps"
)

type Converter struct {
	pbAny     *pbParams
	ormParams *ormParams
	mode      string // dev prod
}

func New(mode string) *Converter {
	return &Converter{
		pbAny:     &pbParams{mode: mode},
		ormParams: &ormParams{mode: mode},
		mode:      mode,
	}
}

func makeErr(mode string, message string) error {
	if mode == "dev" {
		return tps.NewErrWithSkip(2, message)
	}
	return errors.New(message)
}

// orm.ReqParams转db pb XRequestParams
func (c *Converter) ToPBParams(req *orm.ReqParams) (*db.XRequestParams, error) {
	var (
		err  error
		data = &db.XRequestParams{
			Columns:             req.Columns,
			UniqueId:            req.UniqueId,
			UniqueIds:           req.UniqueIds,
			Limit:               int64(req.Limit),
			Page:                int64(req.Page),
			Keyword:             req.Keyword,
			Type:                req.Type,
			All:                 req.All,
			IgnoreNotFound:      req.IgnoreNotFound,
			Backend:             req.Backend,
			IgnoreUpdateColumns: req.IgnoreUpdateColumns,
		}
	)

	if len(req.Orders) > 0 {
		for _, item := range req.Orders {
			var val = db.XSortDirection_ASC
			if strings.ToLower(string(item.Value)) == strings.ToLower(string(orm.OrderDesc)) {
				val = db.XSortDirection_DESC
			}
			data.Orders = append(data.Orders, &db.XOrderItem{
				Column: item.Column,
				Value:  val,
			})
		}
	}

	if len(req.Conditions) > 0 {
		data.Conditions, err = c.pbAny.conditions(req.Conditions)
		if err != nil {
			return nil, makeErr(c.mode, err.Error())
		}
	}

	if len(req.Data) > 0 {
		for _, item := range req.Data {
			var record = &db.XUpdateItem{Column: item.Column}
			if item.Value != nil {
				record.Value, err = c.pbAny.interfaceToPBAny(item.Value)
				if err != nil {
					return nil, makeErr(c.mode, err.Error())
				}
			}

			data.Data = append(data.Data, record)
		}
	}

	if len(req.BulkUpdates) > 0 {
		bulkUpdates, err := functions.JSONMarshal(req.BulkUpdates)
		if err != nil {
			return nil, makeErr(c.mode, err.Error())
		}

		data.BulkUpdates = bulkUpdates
	}

	req.DataRecord = req.DataRecord

	return data, nil
}

// db pb XRequestParams转orm.ReqParams
func (c *Converter) ToOrmParams(req *db.XRequestParams) (*orm.ReqParams, error) {
	var (
		err  error
		data = &orm.ReqParams{
			Columns:             req.Columns,
			UniqueId:            req.UniqueId,
			UniqueIds:           req.UniqueIds,
			Limit:               int(req.Limit),
			Page:                int(req.Page),
			Keyword:             req.Keyword,
			Type:                req.Type,
			All:                 req.All,
			IgnoreNotFound:      req.IgnoreNotFound,
			Backend:             req.Backend,
			IgnoreUpdateColumns: req.IgnoreUpdateColumns,
		}
	)
	if len(req.Orders) > 0 {
		for _, item := range req.Orders {
			var val = orm.OrderAsc
			if item.Value == db.XSortDirection_DESC {
				val = orm.OrderDesc
			}
			data.Orders = append(data.Orders, &orm.OrderItem{
				Column: item.Column,
				Value:  val,
			})
		}
	}

	if len(req.Conditions) > 0 {
		data.Conditions, err = c.ormParams.conditions(req.Conditions)
		if err != nil {
			return nil, makeErr(c.mode, err.Error())
		}
	}

	if len(req.Data) > 0 {
		for _, item := range req.Data {
			var record = &orm.UpdateItem{Column: item.Column}
			if item.Value != nil {
				record.Value, err = c.ormParams.pbAnyToInterface(item.Value)
				if err != nil {
					return nil, makeErr(c.mode, err.Error())
				}
			}

			data.Data = append(data.Data, record)
		}
	}

	if len(req.BulkUpdates) > 0 {
		var bulkUpdates []*orm.BulkUpdateItem
		if err := functions.JSONUnmarshal(req.BulkUpdates, &bulkUpdates); err != nil {
			return nil, makeErr(c.mode, err.Error())
		}

		data.BulkUpdates = bulkUpdates
	}

	if len(req.Data) > 0 {
		data.DataRecord = make(map[string]interface{})
		for _, item := range req.Data {
			data.DataRecord[item.Column], err = c.ormParams.pbAnyToInterface(item.Value)
			if err != nil {
				return nil, makeErr(c.mode, err.Error())
			}
		}
	}

	return data, nil
}
