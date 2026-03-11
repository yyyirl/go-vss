package mediaServers

var (
	ColumnId                       = "id"
	ColumnName                     = "name"
	ColumnIp                       = "ip"
	ColumnExtIp                    = "extIP"
	ColumnPort                     = "port"
	ColumnMediaServerStreamPortMin = "mediaServerStreamPortMin"
	ColumnMediaServerStreamPortMax = "mediaServerStreamPortMax"
	ColumnState                    = "state"
	ColumnIsDef                    = "isDef"
	ColumnCreatedAt                = "createdAt"
	ColumnUpdatedAt                = "updatedAt"
)

var Columns = []string{
	ColumnId,
	ColumnName,
	ColumnIp,
	ColumnExtIp,
	ColumnPort,
	ColumnMediaServerStreamPortMin,
	ColumnMediaServerStreamPortMax,
	ColumnState,
	ColumnIsDef,
	ColumnCreatedAt,
	ColumnUpdatedAt,
}

const (
	PrimaryId = "id"
)
