package leif

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"reflect"
)

type Parser struct {
	NewRoute     func() Mergeable
	Validate     func([]byte) error
	Context      map[string]interface{}
	FilterRoutes func([]string) []string
}

// Default parser
var defaultParser Parser = Parser{

	// The default Route type is RouteDef. If you use a custom mergeable type, then
	// make a callback for building a new one.
	NewRoute: func() Mergeable { return new(RouteDef) },

	// Validation function for checking if the JSON is valid. By default use the
	// Validate() function in the validate.go file.
	// Validation may be turned off by using a callback that returns nil always
	Validate: func(def []byte) error { return Validate(def) },

	// Default is to find all fields that start with '/'
	FilterRoutes: func(keys []string) []string {
		var fkeys []string
		for _, k := range keys {
			if k[0] == '/' {
				fkeys = append(fkeys, k)
			}
		}
		return fkeys
	},
}

// Parse a given JSON file, validating it against the current schema definition using
// the default parser. Set up the context map to hold the defined middlewares
func Parse(def []byte) ([]Route, error) {
	defaultParser.Context = make(map[string]interface{})
	defaultParser.Context["middlewares"] = make(map[string][]function)
	return defaultParser.Parse(def)
}

func (p *Parser) Parse(def []byte) ([]Route, error) {
	var result map[string]interface{}
	json.Unmarshal(def, &result)

	// Use default schema validator
	if err := p.Validate(def); err != nil {
		return []Route{}, err
	}

	if err := p.ParseMiddlewares(result); err != nil {
		return []Route{}, err
	}

	return p.ParseRoutes(result)
}

// Parse the middlewares part of the JSON tree, storing them in the Parser structs
// context for later use
func (p *Parser) ParseMiddlewares(tree map[string]interface{}) error {
	if mwMap, ok := tree["middlewares"]; ok {
		for key, val := range mwMap.(map[string]interface{}) {
			var fns MiddlewareDef
			mapstructure.Decode(val, &fns)
			functions, _ := fns.LoadFunctions(nil)
			ctx := p.Context["middlewares"].(map[string][]function)
			ctx[key] = functions
		}
	}
	return nil
}

func (p *Parser) ParseRoutes(tree map[string]interface{}) ([]Route, error) {
	if routeMap, ok := tree["routes"]; ok {
		rootMap, _ := routeMap.(map[string]interface{})
		routes, _ := p.ParseRoute(rootMap, &RouteDef{})
		return routes, nil
	} else {
		panic("No routes defined")
	}
	return []Route{}, errors.New("Could not parse routes")
}

//
func (p *Parser) ParseRoute(node map[string]interface{}, parent Mergeable) ([]Route, error) {

	// Collect all keys from node, determine which are routes / definitions
	var keys []string
	for k := range node {
		keys = append(keys, k)
	}
	routeNames := p.FilterRoutes(keys)

	// Get new type
	m := p.NewRoute()

	// Decode level to the struct
	err := mapstructure.Decode(node, &m)
	if err != nil {
		panic(err.Error())
	}

	// Link to parent using the struct tag
	err = linkParent(m, parent)
	if err != nil {
		panic(err.Error())
	}

	// Merge with parent
	applyContext(p.Context, m)
	fmt.Println(m)
	err = m.Merge()

	// Call down next level
	var routes []Route
	for _, route := range routeNames {
		if nlevel, ok := node[route].(map[string]interface{}); !ok {
			panic("Route can't be recursed down")
		} else {
			p.Context["route"] = route
			nroutes, _ := p.ParseRoute(nlevel, m)
			routes = append(routes, nroutes...)
		}
	}

	// Build if we should & return
	if buildable, ok := m.(Buildable); ok {
		if buildable.ShouldBuild() {
			applyContext(p.Context, buildable)
			built, _ := buildable.Build()
			routes = append(routes, built...)
		}
	} else {
		panic("Doesn't implement buildable interface")
	}
	return routes, nil
}

func linkParent(child Mergeable, parent Mergeable) error {
	var pfield /*, nested*/ reflect.Value

	retChild := child
	t := reflect.TypeOf(retChild).Elem()
	v := reflect.ValueOf(retChild).Elem()

	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag
		if len(tag) > 0 && tag.Get("sherpa") == "parent" {
			pfield = v.Field(i).Addr()
			pfield.Elem().Set(reflect.ValueOf(parent))
		}
	}
	return nil
}

func applyContext(c map[string]interface{}, m interface{}) error {
	field := reflect.ValueOf(m).Elem().FieldByName("Context").Addr()
	field.Elem().Set(reflect.ValueOf(c))
	return nil
}
