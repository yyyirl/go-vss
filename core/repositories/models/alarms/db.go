package alarms

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[Alarms]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Alarms](db, Alarms{}, 5*time.Second)}
}
