package silky

import (
	"fmt"
	"net/http"

	"regexp"
)

type Router struct {
	*http.ServeMux
	resources  []*Resource
	middleware []func(http.Handler) http.Handler
	routes     []Route
}

// Route is used to map a handler to a path and HTTP method
type Route struct {
	handler http.HandlerFunc
	method  string
	path    string
}

// Resource is a RESTful resource
type Resource struct {
	router      *Router
	path        string
	name        string
	handlers    ResourceHandlers
	middleware  []func(http.Handler) http.Handler
	constraints []Constraint
}

// ResourceHandlers is a set of handlers for a RESTful [gorapid.Resource].
type ResourceHandlers struct {
	Index  http.HandlerFunc
	Show   http.HandlerFunc
	Create http.HandlerFunc
	Update http.HandlerFunc
	Delete http.HandlerFunc
	New    http.HandlerFunc
	Edit   http.HandlerFunc
}

// Constraint is a route parameter constraint.
type Constraint struct {
	Validate func(string) bool
	Pattern  *regexp.Regexp
	Param    string
}

// New returns a new Router.
func NewRouter() *Router {
	return &Router{
		ServeMux:   http.NewServeMux(),
		resources:  make([]*Resource, 0),
		routes:     make([]Route, 0),
		middleware: make([]func(http.Handler) http.Handler, 0),
	}
}

// Resource creates a new RESTful resource for the [gorapid.Router].
func (r *Router) Resource(name string, handlers ResourceHandlers) *Resource {
	resource := &Resource{
		router:   r,
		path:     "/" + name,
		name:     name,
		handlers: handlers,
	}
	r.resources = append(r.resources, resource)
	return resource
}

func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		handler: handler,
		path:    path,
		method:  http.MethodGet,
	})
}

func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		handler: handler,
		path:    path,
		method:  http.MethodPost,
	})
}

func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		handler: handler,
		path:    path,
		method:  http.MethodPut,
	})
}

func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		handler: handler,
		path:    path,
		method:  http.MethodDelete,
	})
}

func (r *Router) Patch(path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		handler: handler,
		path:    path,
		method:  http.MethodPatch,
	})
}

func (r *Router) Options(path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		handler: handler,
		path:    path,
		method:  http.MethodOptions,
	})
}

func (r *Router) Head(path string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		handler: handler,
		path:    path,
		method:  http.MethodHead,
	})
}

func (r *Router) Use(middleware ...func(http.Handler) http.Handler) {
	r.middleware = append(r.middleware, middleware...)
}

func applyMiddleware(handler http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

func (r *Router) Namespace(name string, fn func(r *Router), middleware ...func(http.Handler) http.Handler) {
	subrouter := NewRouter()
	subrouter.Use(middleware...) // namespace-specific middleware
	fn(subrouter)

	prefix := "/" + name
	for _, route := range subrouter.routes {
		r.routes = append(r.routes, Route{
			method:  route.method,
			path:    prefix + "/" + route.path,
			handler: route.handler,
		})
	}

	// Register all resources from subrouter with prefixed paths
	for _, resource := range subrouter.resources {
		newResource := r.Resource(name+"/"+resource.name, resource.handlers).
			Constraints(resource.constraints...)

		// Add subrouter middleware to the resource
		newResource.middleware = append(newResource.middleware, subrouter.middleware...)
	}
}

// Build registers all the routes and middleware with the router. Must be called
// before the router is used.
func (r *Router) Build() {
	// Build simple routes
	for _, route := range r.routes {
		handler := applyMiddleware(route.handler, r.middleware...)
		r.HandleFunc(route.method+" "+route.path, handler.ServeHTTP)
	}

	for _, resource := range r.resources {
		allMiddleware := append(r.middleware, resource.middleware...)

		if resource.handlers.Index != nil {
			r.registerRoute(http.MethodGet, resource.path, resource.handlers.Index, allMiddleware, nil)
		}

		if resource.handlers.Show != nil {
			r.registerRoute(http.MethodGet, resource.path+"/{id}", resource.handlers.Show, allMiddleware, resource.constraints)
		}

		if resource.handlers.Create != nil {
			r.registerRoute(http.MethodPost, resource.path, resource.handlers.Create, allMiddleware, nil)
		}

		if resource.handlers.Update != nil {
			r.registerRoute(http.MethodPut, resource.path+"/{id}", resource.handlers.Update, allMiddleware, resource.constraints)
		}

		if resource.handlers.Delete != nil {
			r.registerRoute(http.MethodDelete, resource.path+"/{id}", resource.handlers.Delete, allMiddleware, resource.constraints)
		}

		if resource.handlers.New != nil {
			r.registerRoute(http.MethodGet, resource.path+"/new", resource.handlers.New, allMiddleware, nil)
		}

		if resource.handlers.Edit != nil {
			r.registerRoute(http.MethodGet, resource.path+"/{id}/edit", resource.handlers.Edit, allMiddleware, resource.constraints)
		}
	}
}

func (r *Router) registerRoute(
	method string,
	path string,
	handler http.HandlerFunc,
	middleware []func(http.Handler) http.Handler,
	constraints []Constraint,
) {
	wrappedHandler := applyMiddleware(
		applyConstraints(handler, constraints),
		middleware...,
	)
	pattern := method + " " + path
	fmt.Printf("Registering route: %s\n", pattern)
	r.HandleFunc(pattern, wrappedHandler.ServeHTTP)
}

func (res *Resource) WithMiddleware(middleware ...func(http.Handler) http.Handler) *Resource {
	res.middleware = append(res.middleware, middleware...)
	return res
}

func (res *Resource) Constraints(constraints ...Constraint) *Resource {
	res.constraints = append(res.constraints, constraints...)
	return res
}

func applyConstraints(handler http.HandlerFunc, constraints []Constraint) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		for _, c := range constraints {
			param := req.PathValue(c.Param)
			if param != "" {
				if c.Pattern != nil && !c.Pattern.MatchString(param) {
					http.Error(w, "Invalid parameter", http.StatusBadRequest)
					return
				}
				if c.Validate != nil && !c.Validate(param) {
					http.Error(w, "Invalid parameter", http.StatusBadRequest)
					return
				}
			}
		}
		handler(w, req)
	}
}
