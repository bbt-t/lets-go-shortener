// Package that creates Router interface.

package controller

import "context"

// HTTPServer - Router interface.
type HTTPServer interface {
	UP() error
	Stop(ctx context.Context) error
}
