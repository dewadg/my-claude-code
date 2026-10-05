// Package buildutil exposes build-time metadata.
package buildutil

// Name is the service name, used for the CLI, the log "service" attribute and
// the default OTEL_SERVICE_NAME.
const Name = "app"

// Version is overridden at build time:
//
//	-ldflags "-X 'example.com/app/internal/pkg/buildutil.Version=<value>'"
var Version = "dev"
