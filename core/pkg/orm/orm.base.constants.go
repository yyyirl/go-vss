package orm

import "gorm.io/gorm"

type OrderType string

const (
	pageSize = 20

	OrderDesc OrderType = "DESC"
	OrderAsc  OrderType = "ASC"

	LogicalOperatorOrSymbol    = "||"
	LogicalOperatorOr          = "or"
	LogicalOperatorOrUppercase = "OR"

	LogicalOperatorAndSymbol    = "&&"
	LogicalOperatorAnd          = "and"
	LogicalOperatorAndUppercase = "AND"
)

var (
	LogicalOperators = []string{
		LogicalOperatorAndSymbol,
		LogicalOperatorOrSymbol,
		LogicalOperatorAnd,
		LogicalOperatorOr,
		LogicalOperatorAndUppercase,
		LogicalOperatorOrUppercase,
	}

	LogicalOperatorsOr = []string{
		LogicalOperatorOrSymbol,
		LogicalOperatorOr,
		LogicalOperatorOrUppercase,
	}
)

const callerFileCtxName = "caller-filer"

var (
	actionInsert = "insert"
	actionUpdate = "update"
	actionDelete = "delete"
	actionUpsert = "upsert"

	ActionInsert ActionType = &actionInsert
	ActionUpdate ActionType = &actionUpdate
	ActionDelete ActionType = &actionDelete
	ActionUpsert ActionType = &actionUpsert
)

var NotFound = gorm.ErrRecordNotFound

const (
	SORT_ASC  = "ASC"
	SORT_DESC = "DESC"
)

const (
	DBTypeMysql     = "mysql"
	DBTypeSqlite    = "sqlite"
	DBTypePostgres  = "postgres"
	DBTypeSqlserver = "sqlserver"
	DBTypeUnknown   = "unknown"
)
