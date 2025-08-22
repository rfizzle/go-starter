package storage

type Scanner interface {
	ScanOne(dst interface{}) error
	ScanAll(dst interface{}) error
}

type Client interface {
	Query(sql string, args ...interface{}) (Scanner, error)
	Execute(sql string, args ...interface{}) error
}
