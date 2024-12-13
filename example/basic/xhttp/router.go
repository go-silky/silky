package xhttp

import (
	"github.com/go-silky/silky"
	"github.com/go-silky/silky/example/basic/controllers"
	"github.com/go-silky/silky/example/basic/views"
)

func SetupRouter() *silky.Router {
	r := silky.NewRouter()

	templRenderer := silky.NewTemplRenderer(views.Error)
	usersController := controllers.NewUsersController(templRenderer)

	r.Resource("users", silky.ResourceHandlers{
		Index: usersController.Index,
		Show:  usersController.Show,
	})

	r.Build()

	return r
}
