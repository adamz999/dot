package main

import (
	"github.com/adamz999/dot/app"
	"github.com/adamz999/dot/context"
	route "github.com/adamz999/dot/router"
)

type Dep struct {
	Value string
}

func main() {
	r := route.NewRouter()
	app := app.New(r)

	dep := &Dep{Value: "A"}

	app.Register(dep)

	r.Get("/test", func(c *context.Context, dep *Dep) {
		c.String(dep.Value)
	})

	app.Start(8080)
}
