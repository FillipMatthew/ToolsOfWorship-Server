package users

import (
	"net/http"
	"time"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api/middleware"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/service"
)

func NewRouter(userService *service.UserService) *Router {
	return &Router{userService: userService}
}

type Router struct {
	userService *service.UserService
}

func (r *Router) Routes() []api.Route {
	authRateLimit := middleware.RateLimitMiddleware(5, 1*time.Minute)

	return []api.Route{
		{Method: http.MethodPost, Pattern: "/api/user/login", Handler: authRateLimit(http.MethodPost, "/api/user/login", loginHandler(r.userService))},
		{Method: http.MethodPost, Pattern: "/api/user/register", Handler: authRateLimit(http.MethodPost, "/api/user/register", registerUserHandler(r.userService))},
		{Method: http.MethodGet, Pattern: "/api/user/verifyemail", Handler: verifyEmailHandler(r.userService)},
	}
}
