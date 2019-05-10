package leif

import (
	//	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Route struct {
	Methods    []string
	Pattern    string
	Host       string
	Schemes    []string
	Queries    []string
	Headers    []string
	Handler    http.HandlerFunc
	Middleware []func(http.Handler) http.Handler
}

// Hold the references
type function struct {
	Package string
	Name    string
	Type    FuncType
	Func    interface{}
}

var registry map[string]function

type FuncType int

const (
	Handler    FuncType = 0
	Middleware FuncType = 1
)

//
func init() {
	registry = make(map[string]function)
}

// Register a function and store the ref internally for when
func RegisterHandler(fns ...http.HandlerFunc) {
	for _, fn := range fns {
		register(Handler, fn)
	}
}

func RegisterMiddleware(fns ...func(http.Handler) http.Handler) {
	for _, fn := range fns {
		register(Middleware, fn)
	}
}

func register(t FuncType, fn interface{}) {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	s := strings.Split(name, ".")
	registry[name] = function{
		Package: s[0],
		Name:    s[1],
		Func:    fn,
		Type:    t,
	}
}
