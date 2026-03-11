package roles

import (
	"time"

	"gorm.io/gorm"

	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[Roles]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Roles](db, Roles{}, 5*time.Second)}
}
