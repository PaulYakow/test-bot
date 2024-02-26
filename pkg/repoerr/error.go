package repoerr

import "errors"

var (
	ErrRecordNotFound         = errors.New("record not found")
	ErrRecordAlreadyExist     = errors.New("record already exist")
	ErrRecordNotAffected      = errors.New("record(-s) not affected")
	ErrRecordNotModifiedSince = errors.New("record not modified since last query")
	ErrConflict               = errors.New("conflict: cannot be completed due to some kind of mismatch")
)
