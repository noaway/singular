package singular

import (
	"net/http"
	"reflect"
)

type Context struct {
	Request *http.Request
	Writer  http.ResponseWriter
	values  Values
	Params  Params
	Keys    map[string]interface{}
}

func (ctx *Context) Set(key string, value interface{}) {
	if ctx.Keys == nil {
		ctx.Keys = make(map[string]interface{})
	}
	ctx.Keys[key] = value
}

func (ctx *Context) Get(key string) (value interface{}, exists bool) {
	if ctx.Keys != nil {
		value, exists = ctx.Keys[key]
	}
	return
}

func (ctx *Context) Apply(v interface{}) interface{} {
	return ctx.values[reflect.TypeOf(v)]
}

func (ctx *Context) MustGet(key string) interface{} {
	if value, exists := ctx.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

func (ctx *Context) Param(key string) string {
	return ctx.Params.ByName(key)
}
