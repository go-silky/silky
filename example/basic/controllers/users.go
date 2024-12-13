package controllers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-silky/silky"
	"github.com/go-silky/silky/example/basic/views/layouts"
	"github.com/go-silky/silky/example/basic/views/users"
)

type UsersController struct {
	*silky.Controller[templ.Component]
}

func NewUsersController(renderer silky.ViewRenderer[templ.Component]) *UsersController {
	return &UsersController{
		Controller: silky.NewController(renderer, layouts.Application),
	}
}

func (c *UsersController) Index(w http.ResponseWriter, r *http.Request) {
	c.RenderWithLayout(w, r, users.Index())
}

func (c *UsersController) Show(w http.ResponseWriter, r *http.Request) {
	c.RenderError(w, r, silky.ErrNotFound, http.StatusNotFound)
}
