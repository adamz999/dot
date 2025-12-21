package rate

import (
	"math"
	"sync"
	"time"
)

type GlobalLimiter struct {
	mu          sync.Mutex
	limiters    map[string]*Limiter
	cap         float64
	rate        float64
	LimitedFunc func()
}

type Limiter struct {
	mu         sync.Mutex
	tokens     float64
	rate       float64
	cap        float64
	lastRefill time.Time
}

func newClientLimiter(cap float64, rates ...float64) *Limiter {
	rate := float64(5)
	if len(rates) > 0 {
		rate = rates[0]
	}
	return &Limiter{
		tokens:     cap,
		rate:       rate,
		cap:        cap,
		lastRefill: time.Now(),
	}
}

func NewLimiter(cap, rate float64) *GlobalLimiter {
	return &GlobalLimiter{
		limiters: make(map[string]*Limiter),
		cap:      cap,
		rate:     rate,
	}
}

func (g *GlobalLimiter) Take(ip string) bool {
	g.mu.Lock()

	lim, ok := g.limiters[ip]

	if !ok {
		g.limiters[ip] = newClientLimiter(g.cap, g.rate)
		lim = g.limiters[ip]
	}

	g.mu.Unlock()

	return lim.take()

}

func (lim *Limiter) take() bool {

	lim.mu.Lock()
	defer lim.mu.Unlock()

	now := time.Now()

	elapsed := time.Since(lim.lastRefill)

	lim.tokens += elapsed.Seconds() * lim.rate
	lim.tokens = math.Min(lim.tokens, lim.cap)
	lim.lastRefill = now

	if lim.tokens >= 1 {
		lim.tokens--
		return true
	}

	return false

}

func (b *GlobalLimiter) OnError(errorFunc func()) {
	b.LimitedFunc = errorFunc
}
