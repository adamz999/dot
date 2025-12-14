package route

import (
	"net/http"
	"strings"

	context "github.com/adamz999/dot/context"
)

type HandlerFunc func(c *context.Context)

type Middleware func(HandlerFunc) HandlerFunc

type Router struct {
	Routes      []Route
	middlewares []Middleware
}

type Route struct {
	Path    string
	Method  string
	Handler HandlerFunc
}

func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

func (r *Router) applyMiddlewares(h HandlerFunc) HandlerFunc {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

func (r *Router) Get(path string, handler HandlerFunc) {
	r.Routes = append(r.Routes,
		Route{
			Path:    path,
			Method:  http.MethodGet,
			Handler: handler,
		},
	)
}

func (r *Router) Post(path string, handler HandlerFunc) {
	r.Routes = append(r.Routes,
		Route{
			Path:    path,
			Method:  http.MethodPost,
			Handler: handler,
		},
	)
}

func (r *Router) Put(path string, handler HandlerFunc) {
	r.Routes = append(r.Routes,
		Route{
			Path:    path,
			Method:  http.MethodPut,
			Handler: handler,
		},
	)
}

func (r *Router) Patch(path string, handler HandlerFunc) {
	r.Routes = append(r.Routes,
		Route{
			Path:    path,
			Method:  http.MethodPatch,
			Handler: handler,
		},
	)
}

func (r *Router) Delete(path string, handler HandlerFunc) {
	r.Routes = append(r.Routes,
		Route{
			Path:    path,
			Method:  http.MethodDelete,
			Handler: handler,
		},
	)
}

func (r *Router) match(req *http.Request, ctx *context.Context) (Route, bool) {
	reqParts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")

	for _, route := range r.Routes {

		if route.Method != req.Method {
			continue
		}

		routeParts := strings.Split(strings.Trim(route.Path, "/"), "/")

		if len(reqParts) != len(routeParts) {
			continue
		}

		params := map[string]string{}

		matched := true

		for i := range routeParts {
			if strings.HasPrefix(routeParts[i], ":") {
				key := routeParts[i][1:]
				params[key] = reqParts[i]
			} else if routeParts[i] != reqParts[i] {
				matched = false
				break
			}

		}

		if matched {
			ctx.Params = params
			return route, true
		}

	}

	return Route{}, false
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	ctx := &context.Context{
		Req:    req,
		Res:    w,
		Values: make(map[string]any),
		Params: make(map[string]string),
	}

	route, found := r.match(req, ctx)

	if !found {
		w.WriteHeader(404)
		w.Write([]byte("not found"))
		return
	}

	handler := r.applyMiddlewares(route.Handler)
	handler(ctx)
}

func (r *Router) Health() {
	r.Get("/health", func(ctx *context.Context) {
		ctx.OK().Body(map[string]string{
			"status": "ok",
		})
	})
}

func NewRouter() *Router {
	return new(Router)
}
