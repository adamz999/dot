package main

import (
	"github.com/adamz999/dot/app"
	ctx "github.com/adamz999/dot/context"
	router "github.com/adamz999/dot/router"
)

func main() {

	r := router.NewRouter()

	r.Get("/hello", func(c *ctx.Context) {
		c.OK().String("Hello, World!")
	})

	r.Post("/echo", func(c *ctx.Context) {
		var data map[string]any
		if err := c.DecodeBody(&data); err != nil {
			c.BadRequest().Error("Invalid JSON")
			return
		}
		c.OK().Body(data)
	})

	r.Get("/set", func(c *ctx.Context) {
		c.Set("user", "Alice")
		val, _ := ctx.Get[string](c, "user")
		c.String(val)
	})

	r.WebSocket("/ws", func(c *ctx.Context) {
		msg, err := c.WebSocket().Read()
		if err != nil {
			panic(err)
		}
		c.WebSocket().Write("cool response" + string(msg))
	})

	a := app.New(r)
	a.Start(8080)
}
