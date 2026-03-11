package skReleases

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[SkReleases]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[SkReleases](db, SkReleases{}, 5*time.Second)}
}
