package observability

import (
	"fmt"
	"os"

	"example.com/app/internal/pkg/buildutil"
)

// setServiceIdentityEnv defaults OTEL_SERVICE_NAME and
// OTEL_RESOURCE_ATTRIBUTES when unset. Every Setup* calls it before building
// an SDK provider because resource.Default() is memoized process-wide: the
// first provider built freezes the resource the others reuse.
func setServiceIdentityEnv(env string) {
	if os.Getenv("OTEL_SERVICE_NAME") == "" {
		_ = os.Setenv("OTEL_SERVICE_NAME", buildutil.Name)
	}
	if os.Getenv("OTEL_RESOURCE_ATTRIBUTES") == "" {
		attrs := fmt.Sprintf("service.version=%s", buildutil.Version)
		if env != "" {
			attrs += ",deployment.environment.name=" + env
		}
		_ = os.Setenv("OTEL_RESOURCE_ATTRIBUTES", attrs)
	}
}
