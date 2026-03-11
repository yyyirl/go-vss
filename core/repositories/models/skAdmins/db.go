package skAdmins

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[SkAdmins]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[SkAdmins](db, SkAdmins{}, 5*time.Second)}
}
