// @Title        pick
// @Description  main
// @Create       yiyiyi 2025/7/11 16:14

package functions

func Pick[T any, V any](records []T, call func(item T) V) []V {
	if len(records) <= 0 {
		return nil
	}

	var (
		length = len(records)
		data   = make([]V, length)
	)
	for i, item := range records {
		data[i] = call(item)
	}

	return data
}

func PickOffsetRangeWithCall(page, limit, total int, call func(start, end int)) {
	if limit >= total {
		call(0, total)
		return
	}

	var start = page * limit
	if start <= 0 {
		start = 0
	}

	if start >= total {
		return
	}

	var end = start + limit
	if end >= total {
		call(start, total)
		return
	}
	call(start, end)

	PickOffsetRangeWithCall(page+1, limit, total, call)
}

func PickWithPageOffset[T any](page, limit int, records []T) []T {
	var length = len(records)
	if length <= 0 {
		return []T{}
	}

	var offset = PageOffset(page, limit)

	// 已取完
	if offset >= length {
		return []T{}
	}

	if offset <= 0 {
		var end = limit
		if end >= length {
			end = length
		}

		return records[0:end]
	}

	var end = offset + limit
	if end >= length {
		end = length
	}

	return records[offset:end]
}
