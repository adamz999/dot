package context

import (
	"encoding/json"
	"fmt"
	"net/http"

	types "github.com/adamz999/dot/params"
	"github.com/gorilla/websocket"
)

type Context struct {
	Req        *http.Request
	Res        http.ResponseWriter
	Values     map[string]any
	Params     map[string]string
	StatusCode int
	Connection *websocket.Conn
	RouteID    string
}

type headerWrapper struct {
	h http.Header
}

type cookieWrapper struct {
	req     *http.Request
	res     http.ResponseWriter
	cookies map[string]*http.Cookie
}

type websocketWrapper struct {
	Connection *websocket.Conn
}

func (c *Context) Status(code int) *Context {
	c.StatusCode = code
	return c
}

func (c *Context) OK() *Context {
	return c.Status(200)
}

func (c *Context) NotFound() *Context {
	return c.Status(404)
}

func (c *Context) BadRequest() *Context {
	return c.Status(400)
}

func (c *Context) InternalServerError() *Context {
	return c.Status(500)
}

func (c *Context) Forbidden() *Context {
	return c.Status(403)
}

func (c *Context) Body(obj any) {
	c.ToJSON(obj)
}

func (c *Context) String(text string) {
	c.ToJSON(map[string]string{
		"message": text,
	})
}

func (c *Context) Error(msg string, code ...int) {
	status := 500
	if len(code) > 0 {
		status = code[0]
	}
	c.Status(status).ToJSON(map[string]string{"error": msg})
}

func (c *Context) ToJSON(obj any) {
	c.Res.Header().Set("Content-Type", "application/json")
	if c.StatusCode == 0 {
		c.StatusCode = 200
	}
	c.Res.WriteHeader(c.StatusCode)
	json.NewEncoder(c.Res).Encode(obj)
}

func (c *Context) Set(key string, val any) {
	if c.Values == nil {
		c.Values = make(map[string]any)
	}
	c.Values[key] = val
}

func Get[T any](c *Context, key string) (T, bool) {
	val, ok := c.Values[key]
	if !ok {
		var null T
		return null, false
	}
	tval, ok := val.(T)
	return tval, ok
}

func (c *Context) Redirect(url string, codes ...int) {
	code := 302
	if len(codes) > 0 {
		code = codes[0]
	}
	http.Redirect(c.Res, c.Req, url, code)
}

func (c *Context) Param(param string) any {
	return types.GetParsedParam(c.RouteID, param, c.Params[param])
}

func (h *headerWrapper) Set(key, val string) {
	h.h.Set(key, val)
}

func (h *headerWrapper) Get(key string) string {
	return h.h.Get(key)
}

func (c *Context) Header() *headerWrapper {
	return &headerWrapper{
		h: c.Req.Header,
	}
}

func (c *Context) Cookie() *cookieWrapper {
	cw := &cookieWrapper{
		req:     c.Req,
		res:     c.Res,
		cookies: make(map[string]*http.Cookie),
	}

	for _, ck := range c.Req.Cookies() {
		cw.cookies[ck.Name] = ck
	}

	return cw

}

func (cw *cookieWrapper) Get(name string) (*http.Cookie, bool) {
	ck, ok := cw.cookies[name]
	return ck, ok
}

func (cw *cookieWrapper) Set(cookie *http.Cookie) {
	http.SetCookie(cw.res, cookie)
	cw.cookies[cookie.Name] = cookie
}

func (c *Context) JSON(status int, obj any) {
	c.Status(status).Body(obj)
}

func (c *Context) Fatal(msg string, codes ...int) {
	code := 500
	if len(codes) > 0 {
		code = codes[0]
	}
	c.Status(code).ToJSON(map[string]string{"error": msg})
	panic("handler aborted")

}

func (c *Context) DecodeBody(obj any) error {
	if err := json.NewDecoder(c.Req.Body).Decode(obj); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func (c *Context) WebSocket() *websocketWrapper {
	return &websocketWrapper{
		Connection: c.Connection,
	}
}

func (ws *websocketWrapper) Read() ([]byte, error) {
	_, message, err := ws.Connection.ReadMessage()
	return message, err
}

func (ws *websocketWrapper) Write(msg string) error {
	return ws.Connection.WriteMessage(websocket.TextMessage, []byte(msg))
}
