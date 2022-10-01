package dbo

import "errors"

var (
	// ErrRecordNotFound record not found
	ErrRecordNotFound = errors.New("record not found")
	// ErrDuplicateRecord duplicate record
	ErrDuplicateRecord = errors.New("duplicate record")
	// ErrExceededLimit exceeded limit
	ErrExceededLimit = errors.New("exceeded limit")
)
