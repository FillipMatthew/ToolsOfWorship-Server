package api

type Route struct {
	Method  string
	Pattern string
	Handler Handler
}

type Router interface {
	Routes() []Route
}

type RouterFunc func() []Route

func (r RouterFunc) Routes() []Route {
	return r()
}

func ComposeRouters(rr ...Router) Router {
	return RouterFunc(func() []Route {
		var rtt []Route
		for _, rt := range rr {
			rtt = append(rtt, rt.Routes()...)
		}

		return rtt
	})
}

type MiddlewareFunc func(method, pattern string, h Handler) Handler

func ChainMiddleware(mm ...MiddlewareFunc) MiddlewareFunc {
	return func(method, pattern string, h Handler) Handler {
		for i := range mm {
			h = mm[len(mm)-i-1](method, pattern, h)
		}

		return h
	}
}
