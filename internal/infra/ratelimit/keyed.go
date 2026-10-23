package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	DefaultIdleTTL       = 10 * time.Minute
	DefaultSweepInterval = time.Minute
)

type (
	Config struct {
		RequestsPerMinute int
		Burst             int
		IdleTTL           time.Duration
		SweepInterval     time.Duration
		Now               func() time.Time
	}

	entry struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	KeyedLimiter struct {
		limit   rate.Limit
		burst   int
		ttl     time.Duration
		now     func() time.Time
		mu      sync.Mutex
		entries map[string]*entry
		stop    chan struct{}
		done    chan struct{}
		once    sync.Once
	}
)

func NewKeyedLimiter(cfg Config) *KeyedLimiter {
	l := newKeyedLimiter(cfg)

	interval := cfg.SweepInterval
	if interval <= 0 {
		interval = DefaultSweepInterval
	}

	go l.janitor(interval)

	return l
}

func newKeyedLimiter(cfg Config) *KeyedLimiter {
	perMinute := cfg.RequestsPerMinute
	if perMinute <= 0 {
		perMinute = 60
	}

	burst := cfg.Burst
	if burst <= 0 {
		burst = perMinute
	}

	now := cfg.Now
	if now == nil {
		now = time.Now
	}

	limit := rate.Limit(float64(perMinute) / 60.0)

	ttl := cfg.IdleTTL
	if ttl <= 0 {
		ttl = DefaultIdleTTL
	}

	// An entry is only dropped once it has fully refilled, so eviction never
	// grants a client more capacity than it would have had.
	refill := time.Duration(float64(burst) / float64(limit) * float64(time.Second))
	if refill > ttl {
		ttl = refill
	}

	return &KeyedLimiter{
		limit:   limit,
		burst:   burst,
		ttl:     ttl,
		now:     now,
		entries: make(map[string]*entry),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (l *KeyedLimiter) Allow(key string) bool {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.entries[key]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.entries[key] = e
	}

	e.lastSeen = now

	return e.limiter.AllowN(now, 1)
}

func (l *KeyedLimiter) Close() {
	l.once.Do(func() { close(l.stop) })
	<-l.done
}

func (l *KeyedLimiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.entries)
}

func (l *KeyedLimiter) evictIdle() {
	cutoff := l.now().Add(-l.ttl)

	l.mu.Lock()
	defer l.mu.Unlock()

	for key, e := range l.entries {
		if e.lastSeen.Before(cutoff) {
			delete(l.entries, key)
		}
	}
}

func (l *KeyedLimiter) janitor(interval time.Duration) {
	defer close(l.done)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.evictIdle()
		case <-l.stop:
			return
		}
	}
}
