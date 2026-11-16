// Package dirlock excludes cooperating filesystem mutations across processes.
package dirlock

import "errors"

// ErrBusy means another helper operation owns the resource. No work was started.
var ErrBusy = errors.New("another Millennium operation is using this directory; retry after it finishes")
