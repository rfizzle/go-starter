package storage

type Storage struct {
}

func Connect(uri string) (*Storage, error) {
	return &Storage{}, nil
}

func (s *Storage) Close() error {
	return nil
}
