package skArticles

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[SkArticles]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[SkArticles](db, SkArticles{}, 5*time.Second)}
}
