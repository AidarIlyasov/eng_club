package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/valyala/fasthttp"
)

// WriteJSON writes a JSON response with the given HTTP status.
func WriteJSON(ctx *fasthttp.RequestCtx, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}
	ctx.Response.Header.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(status)
	ctx.SetBody(b)
}

// WriteErr writes a JSON error object { "error": msg }.
func WriteErr(ctx *fasthttp.RequestCtx, status int, msg string) {
	WriteJSON(ctx, status, map[string]string{"error": msg})
}

// PathID parses a named path segment as int64 (e.g. router param "id").
func PathID(ctx *fasthttp.RequestCtx, key string) (int64, error) {
	v := ctx.UserValue(key)
	if v == nil {
		return 0, fmt.Errorf("missing path param %q", key)
	}
	return strconv.ParseInt(fmt.Sprint(v), 10, 64)
}
