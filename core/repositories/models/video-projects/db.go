package videoProjects

import (
	"time"

	"gorm.io/gorm"

	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[VideoProjects]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[VideoProjects](db, VideoProjects{}, 5*time.Second)}
}
