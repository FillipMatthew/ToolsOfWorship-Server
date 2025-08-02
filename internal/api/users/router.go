package users

import (
	"net/http"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/service"
)

func NewRouter(userService *service.UserService) *Router {
	return &Router{userService: userService}
}

type Router struct {
	userService *service.UserService
}

func (r *Router) Routes() []api.Route {
	return []api.Route{
		{Method: http.MethodPost, Pattern: "/api/user/login", Handler: loginHandler(r.userService)},
		{Method: http.MethodPost, Pattern: "/api/user/register", Handler: registerUserHandler(r.userService)},
		{Method: http.MethodGet, Pattern: "/api/user/verifyemail", Handler: verifyEmailHandler(r.userService)},
	}
}
