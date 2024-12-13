package silky

import "net/http"

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
