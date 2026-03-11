package departments

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[Departments]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Departments](db, Departments{}, 5*time.Second)}
}
