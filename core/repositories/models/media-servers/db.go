package mediaServers

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[MediaServers]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[MediaServers](db, MediaServers{}, 5*time.Second)}
}
