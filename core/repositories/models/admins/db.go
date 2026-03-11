package admins

import (
	"time"

	"gorm.io/gorm"

	orm2 "skeyevss/core/pkg/orm"
)

type DB struct {
	*orm2.Foundation[Admins]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm2.NewFoundation[Admins](db, Admins{}, 5*time.Second)}
}
