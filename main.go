package main

import (
	"github.com/adamz999/dot/app"
	"github.com/adamz999/dot/context"
	"github.com/adamz999/dot/rate"
	route "github.com/adamz999/dot/router"
)

type Dep struct {
	Value string
}

func main() {
	r := route.NewRouter()
	app := app.New(r)

	rl := rate.NewLimiter(5, 3)

	dep := &Dep{Value: "A"}

	app.Register(dep)

	r.Get("/test", func(c *context.Context, dep *Dep) { c.String(dep.Value) }).RouteLimit(rl)

	r.ListRoutes()

	app.Start(8080)
}

// request simulation helpers (allow devs to simulate requests to endpoints) also possibly use fmt time to see
// how long the response takes to return and show time in ms for latency test

// security like rate limitting and api keys

// Route grouping / namespaces (/api/v1/...) with shared middleware

// much better error handling

// metrics like (method, path, status, latency).

// benchmark against gin

// structured jsopn logging
