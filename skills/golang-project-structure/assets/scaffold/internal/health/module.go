package health

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
)

// Module wires the health domain.
var Module = fx.Module("health",
	fx.Provide(fx.Annotate(NewRepository, fx.As(new(healthStore)))),
	fx.Provide(NewService),
	fx.Invoke(RegisterRoutes),
)

type routeParams struct {
	fx.In

	Group   *echo.Group `name:"public"`
	Service *Service
}

// RegisterRoutes is the fx adapter over Register.
func RegisterRoutes(p routeParams) {
	Register(p.Group, Deps{Service: p.Service})
}
