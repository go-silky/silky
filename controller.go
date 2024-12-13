package silky

import (
	"fmt"
	"net/http"
	"reflect"
)

type Controller[T any] struct {
	renderer ViewRenderer[T]
	layout   func(T) T
}

func (c *Controller[T]) SetLayout(layout func(T) T) {
	c.layout = layout
}

func (c *Controller[T]) Render(w http.ResponseWriter, r *http.Request, component T) error {
	return c.renderer.Render(r.Context(), w, component)
}

func (c *Controller[T]) RenderWithLayout(
	w http.ResponseWriter,
	r *http.Request,
	component T,
	layout ...func(T) T,
) error {
	l := c.layout
	if len(layout) > 0 {
		l = layout[0]
	}
	return c.renderer.RenderWithLayout(r.Context(), w, component, l)
}

func (c *Controller[T]) RenderError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	status int,
) error {
	return c.renderer.RenderError(r.Context(), w, err, status)
}

func NewController[T any](renderer ViewRenderer[T], layout ...func(T) T) *Controller[T] {
	c := &Controller[T]{renderer: renderer}
	if len(layout) > 0 {
		c.layout = layout[0]
	}
	return c
}

func MakeResourceHandlers(controller any) ResourceHandlers {
	controllerValue := reflect.ValueOf(controller)
	fmt.Printf("Controller type: %T\n", controller)

	handlers := ResourceHandlers{}

	methodMap := map[string]*http.HandlerFunc{
		"Index":  &handlers.Index,
		"Show":   &handlers.Show,
		"Create": &handlers.Create,
		"Update": &handlers.Update,
		"Delete": &handlers.Delete,
		"New":    &handlers.New,
		"Edit":   &handlers.Edit,
	}

	for methodName, handlerPtr := range methodMap {
		fmt.Printf("Looking for method: %s\n", methodName)

		method := controllerValue.MethodByName(methodName)
		if !method.IsValid() {
			fmt.Printf("Method %s not found\n", methodName)
			continue
		}

		// Create the handler function
		*handlerPtr = func(w http.ResponseWriter, r *http.Request) {
			method.Call([]reflect.Value{
				reflect.ValueOf(w),
				reflect.ValueOf(r),
			})
		}
	}

	return handlers
}
