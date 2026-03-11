package crontab

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[Crontab]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Crontab](db, Crontab{}, 5*time.Second)}
}
