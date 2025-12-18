package app

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/adamz999/dot/registry"
	router "github.com/adamz999/dot/router"
)

type App struct {
	router *router.Router
}

func New(router *router.Router) *App {
	reg := registry.NewServiceRegistry()
	router.Registry = reg

	return &App{
		router: router,
	}
}

func (a *App) Start(port int) {

	strPort := ":" + strconv.Itoa(port)

	if a.router == nil {
		panic("router must be created before starting server")
	}

	printStartupBanner(strPort, a)

	if err := http.ListenAndServe(strPort, a.router); err != nil {
		panic(fmt.Sprintf("server startup failed %v", err))
	}
}

func (a *App) Register(dep any) {
	a.router.Registry.Add(dep)
}

func printStartupBanner(port string, app *App) {
	banner := `

	Server started
	Listening on :%s
	Routes registered: %d

	`
	fmt.Printf(banner, port, len(app.router.Routes))
}
