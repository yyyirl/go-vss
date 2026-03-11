package settings

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[Settings]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[Settings](db, Settings{}, 5*time.Second)}
}
