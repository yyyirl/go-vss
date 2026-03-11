package orm

func (d *DBX[T]) corrections(action ActionType, records []T) []T {
	for key, item := range records {
		records[key] = d.correction(action, item)
	}

	return records
}

func (d *DBX[T]) correction(action ActionType, record T) T {
	return record.Correction(action).(T)
}

func (d *DBX[T]) updateColumns() []string {
	var (
		columns    = d.originalModel.Columns()
		primaryKey = d.originalModel.PrimaryKey()
	)

	if primaryKey == "" {
		return columns
	}

	var (
		i             = 0
		updateColumns = make([]string, len(columns)-1)
	)
	for _, item := range columns {
		if item == primaryKey {
			continue
		}

		updateColumns[i] = item
		i += 1
	}

	return updateColumns
}

func isString[T comparable](value T) bool {
	_, ok := (interface{})(value).(string)
	return ok
}
