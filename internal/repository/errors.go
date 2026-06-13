package repository

import "errors"

var ErrNotFound = errors.New("not found")

type scannable interface {
	Scan(dest ...any) error
}
