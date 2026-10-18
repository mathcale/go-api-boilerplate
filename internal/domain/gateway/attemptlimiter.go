package gateway

type AttemptLimiter interface {
	Allow(key string) bool
}
