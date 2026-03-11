package roles

var (
	ColumnId                  = "id"
	ColumnName                = "name"
	ColumnPermissionUniqueIds = "permissionUniqueIds"
	ColumnState               = "state"
	ColumnRemark              = "remark"
	ColumnIsDel               = "isDel"
	ColumnCreatedAt           = "createdAt"
	ColumnUpdatedAt           = "updatedAt"
)

var Columns = []string{
	ColumnId,
	ColumnName,
	ColumnPermissionUniqueIds,
	ColumnState,
	ColumnRemark,
	ColumnIsDel,
	ColumnCreatedAt,
	ColumnUpdatedAt,
}

const (
	PrimaryId = "id"
)
