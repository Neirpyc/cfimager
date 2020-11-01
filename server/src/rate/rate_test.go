package rate

import (
	. "github.com/franela/goblin"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	g := Goblin(t)
	var currentTime time.Time
	timeProvider = func() time.Time {
		return currentTime
	}
	g.Describe("Rate limiter", func() {
		g.It("Should work with no rule", func() {
			lim := NewLimiter()
			g.Assert(lim.Consume(0)).IsTrue()
			g.Assert(lim.Consume(0)).IsTrue()
		})
		g.It("Should work with one rule", func() {
			lim := NewLimiter(Rule(5, 100*time.Millisecond))
			currentTime = time.Now()
			for i := 0; i < 5; i++ {
				g.Assert(lim.Consume(0)).IsTrue()
			}
			g.Assert(lim.Consume(0)).IsFalse()
			currentTime = currentTime.Add(100 * time.Millisecond)
			for i := 0; i < 5; i++ {
				g.Assert(lim.Consume(0)).IsTrue()
			}
			g.Assert(lim.Consume(0)).IsFalse()
		})
		g.It("Should work with multiple rule", func() {
			lim := NewLimiter(Rule(1, 100*time.Millisecond), Rule(2, 500*time.Millisecond), Rule(3, 3*time.Second))
			g.Assert(lim.Consume(0)).IsTrue()
			g.Assert(lim.Consume(0)).IsFalse()
			currentTime = currentTime.Add(100 * time.Millisecond)
			g.Assert(lim.Consume(0)).IsTrue()
			g.Assert(lim.Consume(0)).IsFalse()
			currentTime = currentTime.Add(100 * time.Millisecond)
			g.Assert(lim.Consume(0)).IsFalse()
			currentTime = currentTime.Add(300 * time.Millisecond)
			g.Assert(lim.Consume(0)).IsTrue()
			g.Assert(lim.Consume(0)).IsFalse()
			currentTime = currentTime.Add(100 * time.Millisecond)
			g.Assert(lim.Consume(0)).IsFalse()
			currentTime = currentTime.Add(400 * time.Millisecond)
			g.Assert(lim.Consume(0)).IsTrue()
		})
	})
	g.Describe("Rate limiter group", func() {
		g.It("Should work with no group", func() {
			lims, err := LimiterGroup()
			g.Assert(err).IsNil()
			currentTime = time.Now()
			for i := 0; i < 5; i++ {
				g.Assert(lims.Consume(0)).IsTrue()
			}
		})
		g.It("Should work with one group", func() {
			lims, err := LimiterGroup(NewLimiter(Rule(1, 100*time.Millisecond), Rule(2, 1*time.Second)))
			g.Assert(err).IsNil()
			currentTime = time.Now()
			g.Assert(lims.Consume(0)).IsTrue()
			g.Assert(lims.Consume(0)).IsFalse()
			currentTime = currentTime.Add(100 * time.Millisecond)
			g.Assert(lims.Consume(0)).IsTrue()
			g.Assert(lims.Consume(0)).IsFalse()
			currentTime = currentTime.Add(100 * time.Millisecond)
			g.Assert(lims.Consume(0)).IsFalse()
		})
		g.It("Should work with multiple group", func() {
			lims, err := LimiterGroup(
				NewLimiter(Rule(1, 100*time.Millisecond), Rule(2, 1*time.Second)),
				NewLimiter(Rule(1, 100*time.Millisecond), Rule(3, 1*time.Second)),
			)
			g.Assert(err).IsNil()
			currentTime = time.Now()
			g.Assert(lims.Consume(0, 0)).IsTrue()
			g.Assert(lims.Consume(0, 0)).IsFalse()
			currentTime = currentTime.Add(100 * time.Millisecond)
			g.Assert(lims.Consume(0, 0)).IsTrue()
			g.Assert(lims.Consume(0, 0)).IsFalse()
			currentTime = currentTime.Add(100 * time.Millisecond)
			g.Assert(lims.Consume(0, 0)).IsFalse()
			g.Assert(lims.Consume(1, 0)).IsTrue()
		})
	})
}
