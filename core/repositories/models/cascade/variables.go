package cascade

var (
	ColumnId                = "id"
	ColumnUniqueId          = "uniqueId"
	ColumnName              = "name"
	ColumnProtocol          = "protocol"
	ColumnSipId             = "sipId"
	ColumnSipDomain         = "sipDomain"
	ColumnSipIp             = "sipIp"
	ColumnSipPort           = "sipPort"
	ColumnUsername          = "username"
	ColumnPassword          = "password"
	ColumnLocalIp           = "localIp"
	ColumnKeepaliveInterval = "keepaliveInterval"
	ColumnRegisterInterval  = "registerInterval"
	ColumnRegisterTimeout   = "registerTimeout"
	ColumnCommandTransport  = "commandTransport"
	ColumnState             = "state"
	ColumnOnline            = "online"
	ColumnCatalogGroupSize  = "catalogGroupSize"
	ColumnRelations         = "relations"
	ColumnCreatedAt         = "createdAt"
	ColumnUpdatedAt         = "updatedAt"
)

var Columns = []string{
	ColumnId,
	ColumnUniqueId,
	ColumnName,
	ColumnProtocol,
	ColumnSipId,
	ColumnSipDomain,
	ColumnSipIp,
	ColumnSipPort,
	ColumnUsername,
	ColumnPassword,
	ColumnLocalIp,
	ColumnKeepaliveInterval,
	ColumnRegisterInterval,
	ColumnRegisterTimeout,
	ColumnCommandTransport,
	ColumnState,
	ColumnOnline,
	ColumnCatalogGroupSize,
	ColumnRelations,
	ColumnCreatedAt,
	ColumnUpdatedAt,
}

const (
	PrimaryId = "id"
)

const (
	_ uint = iota
	Protocol_UDP
	Protocol_TCP
)

var ProtocolMaps = map[uint]string{
	Protocol_UDP: "UDP",
	Protocol_TCP: "TCP",
}

const (
	RegisterStateDef RegisterState = iota
	RegisterStateSuccess
	RegisterStateUnauthorized
	RegisterStateForbidden
	RegisterStateOffline
	RegisterStateOther
)
