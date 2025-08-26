package postgres

import (
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type internalScanner struct {
	rows pgx.Rows
}

func (s *internalScanner) ScanOne(dst interface{}) error {
	return pgxscan.ScanOne(dst, s.rows)
}

func (s *internalScanner) ScanAll(dst interface{}) error {
	return pgxscan.ScanAll(dst, s.rows)
}
