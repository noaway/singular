package singular

import (
	// "fmt"
	"net/http"
	"reflect"
)

//处理是一个函数,可以注册一个途径来处理HTTP
//请求。像http。HandlerFunc,但第三个参数的值
//通配符(变量)。
type (
	Handle     func(*Context)
	Decorator  func(Handle) Handle
	Decorators []Decorator

	//参数是一个URL参数,一个键和一个值组成。
	Param struct {
		Key   string
		Value string
	}

	// Params is a Param-slice, as returned by the router.
	// The slice is ordered, the first URL parameter is also the first slice value.
	// It is therefore safe to read values by the index.
	Params []Param

	// Router is a http.Handler which can be used to dispatch requests to different
	// handler functions via configurable routes
	Router struct {
		values   map[reflect.Type]interface{}
		Handlers Decorators
		singular *Singular
		basePath string
		root     bool
	}
)

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

func (r *Router) Group(path string, ds ...Decorator) *Router {
	return &Router{
		basePath: r.basePath + path,
		Handlers: append(r.Handlers, ds...),
		singular: r.singular,
		values:   make(Values),
		root:     false,
	}
}

func (r *Router) Map(v interface{}) {
	r.values[reflect.TypeOf(v)] = v
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, f Handle, ds ...Decorator) {
	r.singular.Handle("GET", r.basePath+path, r.Decorate(f, append(r.Handlers, ds...)...))
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, f Handle, ds ...Decorator) {
	r.singular.Handle("HEAD", r.basePath+path, r.Decorate(f, append(r.Handlers, ds...)...))
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, f Handle, ds ...Decorator) {
	r.singular.Handle("OPTIONS", r.basePath+path, r.Decorate(f, append(r.Handlers, ds...)...))
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, f Handle, ds ...Decorator) {
	r.singular.Handle("POST", r.basePath+path, r.Decorate(f, append(r.Handlers, ds...)...))
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, f Handle, ds ...Decorator) {
	r.singular.Handle("PUT", r.basePath+path, r.Decorate(f, append(r.Handlers, ds...)...))
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, f Handle, ds ...Decorator) {
	r.singular.Handle("PATCH", r.basePath+path, r.Decorate(f, append(r.Handlers, ds...)...))
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, f Handle, ds ...Decorator) {
	r.singular.Handle("DELETE", r.basePath+path, r.Decorate(f, append(r.Handlers, ds...)...))
}

// Handler is an adapter which allows the usage of an http.Handler as a
// request handle.
func (r *Router) Handler(method, path string, handler http.Handler) {
	r.singular.Handle(method, path,
		func(ctx *Context) {
			handler.ServeHTTP(ctx.Writer, ctx.Request)
		},
	)
}

// HandlerFunc is an adapter which allows the usage of an http.HandlerFunc as a
// request handle.
func (r *Router) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Handler(method, path, handler)
}

func (r *Router) Decorate(f Handle, ds ...Decorator) Handle {
	decorated := f
	handles := []Handle{decorated}
	for _, decorate := range ds {
		decorated = decorate(decorated)
		handles = append(handles, decorated)
	}
	return func(ctx *Context) {
		ctx.values = r.values
		ctx.Handles = handles
		decorated(ctx)
	}
}

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	fileServer := http.FileServer(root)

	// r.GET(path, func(w http.ResponseWriter, req *http.Request, ps Params) {
	// 	req.URL.Path = ps.ByName("filepath")
	// 	fileServer.ServeHTTP(w, req)
	// })

	r.GET(path, func(ctx *Context) {
		// ctx.Request.URL.Path = ps.ByName("filepath")
		fileServer.ServeHTTP(ctx.Writer, ctx.Request)
	})
}

func (r *Router) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		r.singular.PanicHandler(w, req, rcv)
	}
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string) (Handle, Params, bool) {
	if root := r.singular.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (r *Router) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide
		for method := range r.singular.trees {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	} else { // specific path
		for method := range r.singular.trees {
			// Skip the requested method - we already tried this one
			if method == reqMethod || method == "OPTIONS" {
				continue
			}

			handle, _, _ := r.singular.trees[method].getValue(path)
			if handle != nil {
				// add request method to list of allowed methods
				if len(allow) == 0 {
					allow = method
				} else {
					allow += ", " + method
				}
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}
