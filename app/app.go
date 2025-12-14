package app

import (
	"fmt"
	"net/http"
	"strconv"

	router "github.com/adamz999/dot/router"
)

type App struct {
	router *router.Router
}

func New(router *router.Router) *App {
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

func printStartupBanner(port string, app *App) {
	banner := `
	========================================
	____   ____ _________
	|  _ \ |  _ \__   __/
	| | | || | | | / / 
	| |_| || |_| |/ /_ 
	|____/ |____//____|

		Server started!
		Listening on :%s
		Routes registered: %d
	========================================`
	fmt.Printf(banner, port, len(app.router.Routes))
}

// easier json encoding for responses
// middleware
// path vartiable and req params
