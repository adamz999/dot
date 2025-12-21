package route

import (
	"crypto/sha1"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"strings"

	"reflect"

	"github.com/adamz999/dot/context"
	types "github.com/adamz999/dot/params"
	"github.com/adamz999/dot/rate"
	"github.com/adamz999/dot/registry"
	"github.com/adamz999/dot/websocket"
)

type HandlerFunc func(c *context.Context)

type Middleware func(HandlerFunc) HandlerFunc

type Router struct {
	Routes      []*Route
	middlewares []Middleware
	Registry    *registry.ServiceRegistry
	limitted    bool
	limiter     *rate.GlobalLimiter
}

type Route struct {
	ID         string
	Path       string
	Method     string
	Handler    any
	WebSocket  bool
	ParamTypes []reflect.Type
	limitted   bool
	limiter    *rate.GlobalLimiter
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

func (r *Route) initRouteID() {
	hash := sha1.Sum([]byte(r.Method + ":" + r.Path))
	r.ID = hex.EncodeToString(hash[:4])
}

func (r *Router) Get(path string, handler any) *Route {
	route := &Route{
		Path:      path,
		Method:    http.MethodGet,
		Handler:   handler,
		WebSocket: false,
	}
	route.initRouteID()
	extractTypes(route)
	r.extractParams(route)
	r.Routes = append(r.Routes, route)
	return route
}

func (r *Router) Post(path string, handler any) *Route {
	route := &Route{
		Path:      path,
		Method:    http.MethodPost,
		Handler:   handler,
		WebSocket: false,
	}
	route.initRouteID()
	extractTypes(route)
	r.extractParams(route)
	r.Routes = append(r.Routes, route)
	return route
}

func (r *Router) Put(path string, handler any) *Route {
	route := &Route{
		Path:      path,
		Method:    http.MethodPut,
		Handler:   handler,
		WebSocket: false,
	}
	route.initRouteID()
	extractTypes(route)
	r.extractParams(route)
	r.Routes = append(r.Routes, route)
	return route
}

func (r *Router) Patch(path string, handler any) *Route {
	route := &Route{
		Path:      path,
		Method:    http.MethodPatch,
		Handler:   handler,
		WebSocket: false,
	}
	route.initRouteID()
	extractTypes(route)
	r.extractParams(route)
	r.Routes = append(r.Routes, route)
	return route
}

func (r *Router) Delete(path string, handler any) *Route {
	route := &Route{
		Path:      path,
		Method:    http.MethodDelete,
		Handler:   handler,
		WebSocket: false,
	}
	route.initRouteID()
	extractTypes(route)
	r.extractParams(route)
	r.Routes = append(r.Routes, route)
	return route
}

func (r *Router) WebSocket(path string, handler any) *Route {
	route := &Route{
		Path:      path,
		Method:    http.MethodGet,
		Handler:   handler,
		WebSocket: true,
	}
	route.initRouteID()
	r.extractParams(route)
	r.Routes = append(r.Routes, route)
	return route
}

func (r *Router) extractParams(route *Route) {
	t := reflect.TypeOf(route.Handler)
	nparams := t.NumIn()
	params := make([]reflect.Type, nparams)
	for i := 0; i < nparams; i++ {
		params[i] = t.In(i)
	}
	route.ParamTypes = params
}

func extractTypes(route *Route) {
	routeParts := strings.Split(strings.Trim(route.Path, "/"), "/")
	var params []types.RouteParam
	for i := range routeParts {
		if strings.HasPrefix(routeParts[i], ":") {
			part := routeParts[i][1:]
			start := strings.Index(part, "{")
			end := strings.LastIndex(part, "}")
			dtype := "string"
			key := part
			if start != -1 && end != -1 {
				dtype = part[start+1 : end]
				key = part[0:start]
			}
			param := types.RouteParam{
				RouteID: route.ID,
				Name:    key,
				Type:    dtype,
			}
			params = append(params, param)
		}
	}
	types.GlobalRouteParams[route.ID] = params
}

func (r *Router) callHandler(route *Route, ctx *context.Context) {
	v := reflect.ValueOf(route.Handler)
	args := make([]reflect.Value, len(route.ParamTypes))
	for i, t := range route.ParamTypes {
		if t == reflect.TypeOf((*context.Context)(nil)) {
			args[i] = reflect.ValueOf(ctx)
		} else {
			dep := r.Registry.Get(t)
			if dep == nil {
				panic("missing dependency: " + t.String())
			}
			args[i] = reflect.ValueOf(dep)
		}
	}
	v.Call(args)
}

func (r *Router) match(req *http.Request, ctx *context.Context) (*Route, bool) {
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
	return nil, false
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
		w.Write([]byte("404 page not found"))
		return
	}
	ctx.RouteID = route.ID
	if !route.CheckRate(ctx) {
		return
	}
	baseHandler := func(ctx *context.Context) {
		r.callHandler(route, ctx)
	}
	handler := r.applyMiddlewares(baseHandler)
	if route.WebSocket {
		conn := websocket.UpgradeWebsocket(ctx.Res, ctx.Req)
		ctx.Connection = conn
	}
	handler(ctx)
}

func (r *Router) Health() {
	r.Get("/health", func(ctx *context.Context) {
		ctx.OK().Body(map[string]string{
			"status": "ok",
		})
	})
}

func (r *Router) ListRoutes() {
	r.Get("/routes", func(ctx *context.Context) {
		routes := []map[string]string{}
		for _, route := range r.Routes {
			routes = append(routes, map[string]string{
				"path":         route.Path,
				"method":       route.Method,
				"id":           route.ID,
				"rate_limited": strconv.FormatBool(route.limitted),
			})
		}
		ctx.Body(routes)
	})
}

func NewRouter() *Router {
	r := new(Router)
	return r
}

func (r *Route) RouteLimit(rl *rate.GlobalLimiter) {
	r.limitted = true
	r.limiter = rl
}

func (r *Route) CheckRate(ctx *context.Context) bool {

	if r.limiter == nil || !r.limitted {
		return true
	}
	limitFunc := r.limiter.LimitedFunc

	ip, _, _ := net.SplitHostPort(ctx.Req.RemoteAddr)

	for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
		found := ctx.Header().Get(header)
		if found != "" {
			ip = strings.TrimSpace(strings.Split(found, ",")[0])
		}
	}

	if !r.limiter.Take(ip) {
		if limitFunc != nil {
			limitFunc()
		} else {
			ctx.TooManyRequests().String("Too many requests")
		}
		return false
	}
	return true
}
