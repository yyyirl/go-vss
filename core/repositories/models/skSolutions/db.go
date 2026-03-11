package skSolutions

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[SkSolutions]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[SkSolutions](db, SkSolutions{}, 5*time.Second)}
}
