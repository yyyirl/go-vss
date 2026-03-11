package departments

var (
	ColumnId                 = "id"
	ColumnName               = "name"
	ColumnRemark             = "remark"
	ColumnParentId           = "parentId"
	ColumnCascadeDepUniqueId = "cascadeDepUniqueId"
	ColumnRoleIds            = "roleIds"
	ColumnState              = "state"
	ColumnCreatedAt          = "createdAt"
	ColumnUpdatedAt          = "updatedAt"
)

var Columns = []string{
	ColumnId,
	ColumnName,
	ColumnRemark,
	ColumnParentId,
	ColumnCascadeDepUniqueId,
	ColumnRoleIds,
	ColumnState,
	ColumnCreatedAt,
	ColumnUpdatedAt,
}

const (
	PrimaryId = "id"
)
