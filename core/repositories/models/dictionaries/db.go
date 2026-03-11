package dictionaries

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[Dictionaries]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Dictionaries](db, Dictionaries{}, 5*time.Second)}
}
