// Implementation of errors.

package storage

import "errors"

// Errors for storage response.
var (
	ErrExists   = errors.New("url is already exists")
	ErrDeleted  = errors.New("url is deleted")
	ErrNotFound = errors.New("not found")
)
