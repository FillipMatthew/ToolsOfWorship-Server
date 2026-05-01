package fellowships

import (
	"net/http"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/service"
)

func NewRouter(fellowshipService *service.FellowshipService) *Router {
	return &Router{fellowshipService: fellowshipService}
}

type Router struct {
	fellowshipService *service.FellowshipService
}

func (r *Router) Routes() []api.Route {
	listLimit := api.WithBodyLimit(512)

	return []api.Route{
		{
			Method:   http.MethodPost,
			Pattern:  "/api/fellowships/list",
			Handler:  listLimit(http.MethodPost, "/api/fellowships/list", list(r.fellowshipService)),
		},
	}
}
