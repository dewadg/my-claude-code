package httpserver

import "go.uber.org/fx"

// Module provides the router, the "public" and "private" route groups and
// the *http.Server, and registers middleware and lifecycle.
var Module = fx.Module("httpserver",
	fx.Provide(NewEcho),
	fx.Provide(fx.Annotate(NewPublicGroup, fx.ResultTags(`name:"public"`))),
	fx.Provide(fx.Annotate(NewPrivateGroup, fx.ResultTags(`name:"private"`))),
	fx.Provide(NewHTTPServer),
	fx.Invoke(RegisterMiddleware),
	fx.Invoke(RegisterLifecycle),
)
