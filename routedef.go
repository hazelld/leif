package leif

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

// Internal structure to represent the parsed json under the "routes" part of the
// tree. As the json tree is recursively decended down, the route.Merge() defines how
// the parent route should be merged onto the new child route. Only parts of the tree
// that define 'function' will be converted into the concrete Route type that is
// returned by sherpa.
//
// If you need extended functionality, this structure should be embedded within the
// custom struct and then custom.RouteDef.Merge() should be called in the custom
// merge. See docs for more.
type RouteDef struct {
	Methods     []string
	Host        string
	Schemes     []string
	Headers     []string
	Queries     []string
	Function    string
	Middlewares MiddlewareDef
	Exclude     []string
	Package     string
	Pattern     string

	// Global Parser context is added to the RouteDef for use by the build method
	Context map[string]interface{}

	// Hold the parent node for access to it's fields for merging
	RouteDefParent *RouteDef `sherpa:"parent"`
}

// Internal structure to represent middlewares. Again, middlewares.Merge() defines
// how parents should be merged onto children. Essentially:
// - Union of parent functions and child functions (use exclude-middleware to prevent
// 	 certain middlewares from being used at a given level
// - If child defines package the parent is overwritten
type MiddlewareDef struct {
	Package   string
	Functions []string
	Exclude   []string
}

// Merge each RouteDef with parent. Full rules for merging is outlined in README
func (r *RouteDef) Merge() error {

	if len(r.Methods) == 0 {
		r.Methods = r.RouteDefParent.Methods
	}
	if len(r.Host) == 0 {
		r.Host = r.RouteDefParent.Host
	}
	if len(r.Schemes) == 0 {
		r.Schemes = r.RouteDefParent.Schemes
	}
	if len(r.Headers) == 0 {
		r.Headers = r.RouteDefParent.Headers
	}
	if len(r.Queries) == 0 {
		r.Queries = r.RouteDefParent.Queries
	}
	if len(r.Package) == 0 {
		r.Package = r.RouteDefParent.Package
	}

	// Since there is no real "pattern" field in the JSON, rather it is the parent's
	// key, it must be passed through the Context, since ParseRoute() only does work
	// on interfaces not structs.
	pattern, ok := r.Context["route"].(string)
	if len(pattern) == 0 || !ok {
		r.Pattern = r.RouteDefParent.Pattern
	} else {

		// Don't double up on '/', so remove if it exists
		if r.RouteDefParent.Pattern != "/" {
			r.Pattern = r.RouteDefParent.Pattern + pattern
		} else {
			r.Pattern = pattern
		}
	}

	return nil
}

// Merge the middlewares with the following rules:
// - Remove excluded functions from parent
// - Union between child + parent
// - Apply CHILD exclude
// - If child defines package, overwrite parent
//
// Note that the excludes are merged down because if the parent level has excluded a
// function, it will not be passed to the child. However, the child can always re-add
// this function by including it in the 'functions' array.
func (child *MiddlewareDef) Merge(parent MiddlewareDef) error {
	parentFunctions := difference(parent.Functions, parent.Exclude)
	child.Functions = union(child.Functions, parentFunctions)
	child.Functions = difference(child.Functions, child.Exclude)
	if len(child.Package) == 0 {
		child.Package = parent.Package
	}
	return nil
}

// Build the RouteDef into a Route
/*
func (r *RouteDef) Build() (Route, error) {
}
*/

// Method that will attempt to load the functions defined by a middlewares, from the
// functions that have been registered through RegisterMiddlewares(). The function
// type holds the reference to the actual function location.
//
// Refs are expanded here, by looking them up in the internal index of function
// collections. Note the functions from the index have already been checked to ensure
// they have been registered, so that check isn't done two times.
//
// Note that expanded refs ALWAYS keep their package defined in the ref. It is never
// overwritten by defining a package at the 'middlewares' level. For example:
//
// 'middlewares': { 'api': { 'functions': 'Validate', 'package': 'mw' } }
// ...
// '/route': {
// 		...
// 		'middlewares': {
//			'functions': [ '$api', 'OtherMW' ],
//			'package': 'other',
//		}
// }
//
// Will yield the following middlewares being applied to '/route': mw.Validate,
// other.OtherMW
//
// Also note that if you define a function with the package applied (ie.
// package.function), then the 'package' field will be ignored for that specific
// function. This is mostly for excluding functions that may not be in the same
// package as middlewares being added (not sure why this would be the case, but just
// incase). Excluding functions will by default also take on the package defined at
// that level.
func (m *MiddlewareDef) LoadFunctions(index map[string][]function) ([]function, error) {
	var funcs []function

	for _, f := range m.Functions {

		// Expand references
		if ref, ok := isRef(f); ok && index != nil {
			mws, _ := index[ref]
			funcs = append(funcs, mws...)
			continue
		}

		// Get the full function in form 'package.function'
		funcName := f
		if _, ok := hasPackage(f); !ok {
			funcName = m.Package + "." + f
		}

		// Ensure there is a registered function
		if fn, ok := registry[funcName]; ok {
			if fn.Type != Middleware {
				err := errors.New("Function " + funcName + " doesn't match middleware signature")
				return []function{}, err
			}
			funcs = append(funcs, fn)
		} else {
			err := errors.New("Function " + funcName + " has not been registered")
			return []function{}, err
		}
	}
	return funcs, nil
}

// Implement the Buildable interface for RouteDef

// Want to build the routes when we we have a defined function at that given level
func (rd *RouteDef) ShouldBuild() bool {
	return len(rd.Function) != 0
}

func (rd *RouteDef) Build() ([]Route, error) {
	var handler http.HandlerFunc

	// Lookup if function name exists, throw error if it doesn't
	funcName := rd.Package + "." + rd.Function
	if fn, ok := registry[funcName]; ok {
		if fn.Package != rd.Package {
			panic("Function doesn't exist in package")
		}
		handler, _ = fn.Func.(http.HandlerFunc)
	} else {
		panic("Function doesn't exist")
	}

	// Handle middlewares
	var middlewares []func(http.Handler) http.Handler
	ctx := rd.Context["middlewares"].(map[string][]function)
	m, _ := rd.Middlewares.LoadFunctions(ctx)
	for _, mw := range m {
		i := mw.Func.(func(http.Handler) http.Handler)
		middlewares = append(middlewares, i)
	}

	//fmt.Printf("%+v\n", rd)
	r := Route{
		Methods:    rd.Methods,
		Pattern:    rd.Pattern,
		Host:       rd.Host,
		Schemes:    rd.Schemes,
		Queries:    rd.Queries,
		Headers:    rd.Headers,
		Middleware: middlewares,
		Handler:    handler,
	}
	//fmt.Println(r)
	fmt.Println("")
	return []Route{r}, nil
}

// Test if a given string is a ref to a collection of functions. Currently this is
// defined by: '$var' where var is the key of the other part of the JSON
func isRef(ref string) (string, bool) {
	if len(ref) > 0 {
		if []rune(ref)[0] == '$' {
			return ref[1:], true
		}
	}
	return "", false
}

// Test if string contains a package & function, or just a function. Package +
// function strings will have the form 'package.function'. Returns the split package
// and function if this is the case
func hasPackage(function string) ([]string, bool) {
	split := strings.Split(function, ".")

	if len(split) > 1 {
		return split, true
	}
	return split, false
}
