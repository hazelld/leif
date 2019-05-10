package builder

import (
	"fmt"
	"github.com/gorilla/mux"
	"sherpa"
)

// Create a mux router from a definition, assuming all handlers/middlewares have
// been defined
func Mux(def []byte) (*mux.Router, error) {
	r := mux.NewRouter()
	routes, _ := sherpa.Parse(def)

	for _, r := range routes {

	}
	return r, nil
}

func MuxRouteBuilder(r sherpa.Route) *mux.Route {
	route := mux.NewRoute()

	route.Path(r.Pattern).
		Headers(r.Headers...).
		Host(r.Host).
		Methods(r.Methods...).
		Queries(r.Queries...).
		Schemes(r.Schemes...).
		Handler(r.Handler)
}
