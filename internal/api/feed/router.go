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
	listLimit := api.WithBodyLimit(512)
	postLimit := api.WithBodyLimit(65536)

	return []api.Route{
		{
			Method:  http.MethodPost,
			Pattern: "/api/feed/list",
			Handler: listLimit(http.MethodPost, "/api/feed/list", list(r.feedService)),
		},
		{
			Method:  http.MethodPost,
			Pattern: "/api/feed/post",
			Handler: postLimit(http.MethodPost, "/api/feed/post", post(r.feedService)),
		},
	}
}
