package crontab

import "skeyevss/core/pkg/functions"

var (
	ColumnUniqueId    = "uniqueId"
	ColumnTitle       = "title"
	ColumnInterval    = "interval"
	ColumnCounter     = "counter"
	ColumnBlockStatus = "blockStatus"
	ColumnStatus      = "status"
	ColumnTimeout     = "timeout"
	ColumnLogs        = "logs"
	ColumnReadonly    = "readonly"
	ColumnCreatedAt   = "createdAt"
	ColumnUpdatedAt   = "updatedAt"
)

var Columns = []string{
	ColumnUniqueId,
	ColumnTitle,
	ColumnInterval,
	ColumnCounter,
	ColumnBlockStatus,
	ColumnStatus,
	ColumnTimeout,
	ColumnLogs,
	ColumnReadonly,
	ColumnCreatedAt,
	ColumnUpdatedAt,
}

const (
	PrimaryUniqueId = "uniqueId"
)

const (
	UniqueIdVideoProject = "video-project"
)

var UniqueIds = []string{
	UniqueIdVideoProject,
}

var (
	now         = uint64(functions.NewTimer().NowMilli())
	InitRecords = []Crontab{
		{
			UniqueId:    UniqueIdVideoProject,
			Title:       "录像计划",
			Interval:    1,
			Counter:     40,
			BlockStatus: 1,
			Status:      1,
			Readonly:    1,
			Timeout:     10,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
)
