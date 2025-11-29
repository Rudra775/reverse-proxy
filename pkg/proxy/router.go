package proxy

import (
	"errors"
	"net/http"
	"strings"
)

type route struct {
	prefix   string
	balancer Balancer
}

type Router struct {
	routes []route
}

func NewRouter(cfg *Config) *Router {
	r := &Router{}

	for _, rc := range cfg.Routes {
		var bal Balancer

		switch rc.Strategy {
		case "round_robin":
			bal = NewRoundRobin(rc.Backends)
		default:
			bal = NewRoundRobin(rc.Backends)
		}

		r.routes = append(r.routes, route{
			prefix:   rc.PathPrefix,
			balancer: bal,
		})
	}
	return r
}

func (r *Router) Match(req *http.Request) (*Backend, error) {
	path := req.URL.Path
	var best *route

	for i := range r.routes {
		rt := &r.routes[i]
		if strings.HasPrefix(path, rt.prefix) {
			if best == nil || len(rt.prefix) > len(best.prefix) {
				best = rt
			}
		}
	}

	if best == nil {
		return nil, errors.New("no matching route")
	}

	return best.balancer.Next()
}
