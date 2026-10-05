package infra

import "go.uber.org/fx"

// Module wires infra. StartRuntimeMetrics is an Invoke because it is a side
// effect nothing consumes, and fx never builds an unconsumed provider; it
// also pulls the logger and meter provider up before any domain constructor.
var Module = fx.Module("infra",
	fx.Provide(NewBaseLogger),
	fx.Provide(fx.Annotate(
		NewLogger,
		fx.ParamTags(``, ``, `name:"base"`),
	)),
	fx.Provide(NewMeterProvider),
	fx.Provide(NewTracerProvider),
	fx.Provide(NewDB),
	fx.Invoke(StartRuntimeMetrics),
)
