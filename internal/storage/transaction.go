package storage

type Tx interface {
	InsertMany(queries map[string][]interface{}) error
	Client
}
