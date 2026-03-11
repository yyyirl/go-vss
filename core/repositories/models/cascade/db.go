package cascade

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[Cascade]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Cascade](db, Cascade{}, 5*time.Second)}
}
