package proxy

import (
	"net/http"
	"strings"

	"proxyserver/internal/config"
)

type Route struct {
	Prefix   string
	Balancer Balancer
}

type Router struct {
	routes []Route
}

func NewRouter(cfg *config.Config) *Router {
	r := &Router{}

	for _, rc := range cfg.Routes {
		var bal Balancer
		switch rc.Strategy {
		case "least_conn":
			bal = NewLeastConn(rc.Backends)
		default:
			bal = NewRoundRobin(rc.Backends)
		}

		r.routes = append(r.routes, Route{
			Prefix:   rc.PathPrefix,
			Balancer: bal,
		})
	}
	return r
}

// MatchRoute returns the backend for this request path (or nil if none)
func (r *Router) MatchBackend(req *http.Request) *Backend {
	path := req.URL.Path
	var best *Route
	for i := range r.routes {
		rt := &r.routes[i]
		if strings.HasPrefix(path, rt.Prefix) {
			// longest prefix match (more specific route wins)
			if best == nil || len(rt.Prefix) > len(best.Prefix) {
				best = rt
			}
		}
	}
	if best == nil {
		return nil
	}
	return best.Balancer.Next()
}
