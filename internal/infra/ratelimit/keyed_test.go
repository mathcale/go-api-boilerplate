package ratelimit

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.t
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.t = c.t.Add(d)
}

type KeyedLimiterTestSuite struct {
	suite.Suite
	clock *fakeClock
}

func (s *KeyedLimiterTestSuite) SetupTest() {
	s.clock = &fakeClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func TestKeyedLimiter(t *testing.T) {
	suite.Run(t, new(KeyedLimiterTestSuite))
}

func (s *KeyedLimiterTestSuite) newLimiter(perMinute, burst int) *KeyedLimiter {
	return newKeyedLimiter(Config{
		RequestsPerMinute: perMinute,
		Burst:             burst,
		IdleTTL:           10 * time.Minute,
		Now:               s.clock.Now,
	})
}

func (s *KeyedLimiterTestSuite) TestAllow() {
	s.Run("should allow requests up to the burst then reject", func() {
		l := s.newLimiter(60, 2)

		s.True(l.Allow("a"))
		s.True(l.Allow("a"))
		s.False(l.Allow("a"))
	})

	s.Run("should track each key independently", func() {
		l := s.newLimiter(60, 1)

		s.True(l.Allow("a"))
		s.True(l.Allow("b"))
		s.False(l.Allow("a"))
	})

	s.Run("should refill over time", func() {
		l := s.newLimiter(60, 1)

		s.True(l.Allow("a"))
		s.False(l.Allow("a"))

		s.clock.Advance(time.Second)

		s.True(l.Allow("a"))
	})

	s.Run("should apply defaults for non-positive settings", func() {
		l := newKeyedLimiter(Config{Now: s.clock.Now})

		s.Equal(60, l.burst)
		s.Equal(DefaultIdleTTL, l.ttl)
	})
}

func (s *KeyedLimiterTestSuite) TestEvictIdle() {
	s.Run("should evict entries idle for longer than the TTL", func() {
		l := s.newLimiter(60, 5)
		l.Allow("idle")

		s.clock.Advance(11 * time.Minute)
		l.evictIdle()

		s.Equal(0, l.Len())
	})

	s.Run("should keep entries seen within the TTL", func() {
		l := s.newLimiter(60, 5)
		l.Allow("idle")
		l.Allow("active")

		s.clock.Advance(9 * time.Minute)
		l.Allow("active")
		s.clock.Advance(2 * time.Minute)
		l.evictIdle()

		s.Equal(1, l.Len())
		s.True(l.Allow("active"))
	})

	s.Run("should not evict before a full burst has refilled", func() {
		l := s.newLimiter(1, 20)
		l.Allow("slow")

		s.clock.Advance(11 * time.Minute)
		l.evictIdle()

		s.Equal(1, l.Len())

		s.clock.Advance(10 * time.Minute)
		l.evictIdle()

		s.Equal(0, l.Len())
	})
}

func (s *KeyedLimiterTestSuite) TestJanitor() {
	s.Run("should evict idle entries in the background", func() {
		l := NewKeyedLimiter(Config{
			RequestsPerMinute: 60,
			Burst:             5,
			IdleTTL:           time.Minute,
			SweepInterval:     5 * time.Millisecond,
			Now:               s.clock.Now,
		})
		defer l.Close()

		l.Allow("a")
		s.clock.Advance(2 * time.Minute)

		s.Eventually(func() bool { return l.Len() == 0 }, 2*time.Second, 5*time.Millisecond)
	})
}

func (s *KeyedLimiterTestSuite) TestClose() {
	s.Run("should stop the janitor and be safe to call twice", func() {
		l := NewKeyedLimiter(Config{SweepInterval: time.Millisecond, Now: s.clock.Now})

		l.Close()
		l.Close()

		select {
		case <-l.done:
		default:
			s.Fail("janitor still running")
		}
	})
}
