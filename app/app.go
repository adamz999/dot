package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/adamz999/dot/registry"
	router "github.com/adamz999/dot/router"
)

type App struct {
	router     *router.Router
	server     *http.Server
	startHooks []func()
	errorHooks []func()
	stopHooks  []func()
}

func (a *App) OnServerStart(hook func()) {
	a.startHooks = append(a.startHooks, hook)
}

func (a *App) OnServerError(hook func()) {
	a.errorHooks = append(a.errorHooks, hook)
}

func (a *App) OnServerStop(hook func()) {
	a.stopHooks = append(a.stopHooks, hook)
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

	server := &http.Server{
		Addr:    strPort,
		Handler: a.router,
	}

	runStartHooks(a)

	printStartupBanner(strPort, a)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			runErrorHooks(a)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)

	runStopHooks(a)
}

func runStartHooks(a *App) {
	for _, hook := range a.startHooks {
		hook()
	}
}

func runErrorHooks(a *App) {
	for _, hook := range a.errorHooks {
		hook()
	}
}

func runStopHooks(a *App) {
	for _, hook := range a.stopHooks {
		hook()
	}
}

func (a *App) Register(dep any) {
	a.router.Registry.Add(dep)
}

func printStartupBanner(port string, app *App) {
	banner := `

	Server started
	Listening on %s
	Routes registered: %d

	`
	fmt.Printf(banner, port, len(app.router.Routes))
}
