package orm

func (d *DBX[T]) MakeLimit(params *ReqParams) *Pagination {
	if params.All {
		return nil
	}

	var limit = pageSize
	if params.Limit > 0 {
		limit = params.Limit
	}

	return &Pagination{
		Limit:  limit,
		Offset: d.PageOffset(params.Page, limit),
	}
}

func (d *DBX[T]) PageOffset(page, limit int) int {
	if page <= 0 {
		return 0
	}

	page = page - 1

	if limit == 0 {
		limit = 20
	}

	return page * limit
}
