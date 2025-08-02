package feed

import (
	"net/http"

	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/api"
	"github.com/FillipMatthew/ToolsOfWorship-Server/internal/service"
)

func NewRouter(feedService *service.FeedService) *Router {
	return &Router{feedService: feedService}
}

type Router struct {
	feedService *service.FeedService
}

func (r *Router) Routes() []api.Route {
	return []api.Route{
		{Method: http.MethodPost, Pattern: "/api/feed/list", Handler: list(r.feedService)},
		{Method: http.MethodPost, Pattern: "/api/feed/post", Handler: post(r.feedService)},
	}
}
