package egress

type StorageOption interface {
	Apply(*StorageConfig)
}

type storageTable struct {
	tableName string
}

var _ StorageOption = storageTable{}

func (e storageTable) Apply(config *StorageConfig) {
	config.TableName = e.tableName
}

func WithEgressTable(table string) StorageOption {
	return storageTable{tableName: table}
}
