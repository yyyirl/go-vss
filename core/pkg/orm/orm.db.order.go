package orm

import (
	"fmt"
	"strings"

	"skeyevss/core/pkg/functions"
)

func (d *DBX[T]) Order(params *ReqParams) (string, error) {
	var length = len(params.Orders)
	if length <= 0 {
		return "", nil
	}

	var (
		columns = d.originalModel.Columns()
		orders  = make([]string, length)
		i       = 0
	)
	for _, item := range params.Orders {
		if !functions.Contains(item.Column, columns) {
			return "", fmt.Errorf("排序字段[%s]不存在", item.Column)
		}

		item.Value = OrderType(strings.ToUpper(string(item.Value)))
		var val = OrderDesc
		if item.Value != OrderDesc {
			val = OrderAsc
		}

		orders[i] = item.Column + " " + string(val)
		i++
	}

	return strings.Join(orders, ", "), nil
}
