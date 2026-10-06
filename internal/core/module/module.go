package module

import "net/http"

const (
	Public       = false
	Protected    = true
	NotThrottled = false
	RateLimited  = true
)

type Route struct {
	Method        string
	Path          string
	Protected     bool
	RateLimited   bool
	RequiredRoles []string
	Handler       http.HandlerFunc
}

type Module interface {
	Name() string
	Routes() []Route
}
