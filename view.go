package silky

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/a-h/templ"
)

type ViewRenderer[T any] interface {
	Render(ctx context.Context, w io.Writer, component T) error
	RenderWithLayout(ctx context.Context, w io.Writer, component T, layout func(T) T) error
	RenderError(ctx context.Context, w http.ResponseWriter, err error, status int) error
}

type TemplRenderer struct {
	errorHandler func(error, int) templ.Component
}

func NewTemplRenderer(errorHandler func(error, int) templ.Component) *TemplRenderer {
	return &TemplRenderer{errorHandler: errorHandler}
}

func (r *TemplRenderer) Render(ctx context.Context, w io.Writer, c templ.Component) error {
	fmt.Println("i am here")
	return c.Render(ctx, w)
}

func (r *TemplRenderer) RenderWithLayout(
	ctx context.Context,
	w io.Writer,
	component templ.Component,
	layout func(templ.Component) templ.Component,
) error {
	return layout(component).Render(ctx, w)
}

func (r *TemplRenderer) RenderError(ctx context.Context, w http.ResponseWriter, err error, status int) error {
	if r.errorHandler != nil {
		return r.errorHandler(err, status).Render(ctx, w)
	}
	// Fallback error handling if no error handler is configured
	http.Error(w, err.Error(), status)
	return nil
}
