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

func (u *Router) Routes() []api.Route {
	return []api.Route{
		{Method: http.MethodPost, Pattern: "/api/user/login", Handler: loginHandler(u.userService)},
		{Method: http.MethodPost, Pattern: "/api/user/register", Handler: registerUserHandler(u.userService)},
		{Method: http.MethodGet, Pattern: "/api/user/verifyemail", Handler: verifyEmailHandler(u.userService)},
	}
}
