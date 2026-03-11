package systemOperationLogs

import (
	"time"

	"gorm.io/gorm"
	"skeyevss/core/pkg/orm"
)

type DB struct {
	*orm.Foundation[SystemOperationLogs]
}

func NewDB(db *gorm.DB) *DB {
	return &DB{orm.NewFoundation[SystemOperationLogs](db, SystemOperationLogs{}, 5*time.Second)}
}
