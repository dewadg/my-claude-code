package health

import "errors"

// ErrDatabaseUnreachable wraps a failed database ping.
var ErrDatabaseUnreachable = errors.New("health.database_unreachable")
