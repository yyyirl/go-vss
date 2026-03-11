package orm

import "context"

func (d *DBX[T]) Insert(ctx context.Context, dbInstance DB, records []T) error {
	return dbInstance.WithContext(ctx).CreateInBatches(d.corrections(ActionInsert, records), len(records)).Error
}

func (d *DBX[T]) Add(ctx context.Context, dbInstance DB, record T) (*T, error) {
	var row = d.correction(ActionInsert, record)
	if err := dbInstance.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, err
	}

	return &row, nil
}
