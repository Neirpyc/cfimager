package rate

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var timeProvider = time.Now

func timeNow() time.Time {
	return timeProvider()
}

type Limiter interface {
	Consume(i ...interface{}) bool
	Clean()
}

type limiterInternal interface {
	Consume(i ...interface{}) bool
	Clean()
	undoConsume(interface{})
}

type limiterStock struct {
	updated time.Time
	stock   int
}

func (l *limiterStock) update(r *limiterRule) {
	updateLength := timeNow().Sub(l.updated).Truncate(r.refill)
	l.stock += int(updateLength.Milliseconds() / r.refill.Milliseconds())
	if l.stock > r.maxCalls {
		l.stock = r.maxCalls
	}
	l.updated = l.updated.Add(updateLength)
}

func (l *limiterStock) undoConsume() {
	l.stock++
}

func (l *limiterStock) consume() bool {
	if l.stock > 0 {
		l.stock--
		return true
	}
	return false
}

type limiterStocks struct {
	stocks []*limiterStock
	mutex  sync.RWMutex
}

func (l *limiterStocks) update(r limiterRules) {
	for i, lim := range l.stocks {
		lim.update(r[i])
	}
}

func (l *limiterStocks) undoConsume() {
	for _, lim := range l.stocks {
		lim.undoConsume()
	}
}

func (l *limiterStocks) consume() bool {
	for i, lim := range l.stocks {
		if !lim.consume() {
			for j := 0; j < i; j++ {
				l.stocks[j].undoConsume()
			}
			return false
		}
	}
	return true
}

type limiterRule struct {
	maxCalls int
	refill   time.Duration
}

func Rule(maxCalls int, window time.Duration) *limiterRule {
	return &limiterRule{
		maxCalls: maxCalls,
		refill:   time.Duration(int64(window) / int64(maxCalls)),
	}
}

type limiterRules []*limiterRule

func (l *limiterRules) stock() *limiterStocks {
	limStock := limiterStocks{
		stocks: make([]*limiterStock, len(*l)),
	}
	for i, r := range *l {
		limStock.stocks[i] = &limiterStock{
			updated: timeNow(),
			stock:   r.maxCalls,
		}
	}
	return &limStock
}

type limiter struct {
	m     sync.RWMutex
	stock map[interface{}]*limiterStocks
	rules limiterRules
}

func (l *limiter) Consume(i ...interface{}) bool {
	if len(i) != 1 {
		return false
	}
	l.m.RLock()
	if limiters, found := l.stock[i[0]]; found {
		defer l.m.RUnlock()
		limiters.update(l.rules)
		return limiters.consume()
	}
	l.m.RUnlock()
	l.m.Lock()
	defer l.m.Unlock()
	stock := l.rules.stock()
	l.stock[i[0]] = stock
	return l.stock[i[0]].consume()
}

func (l *limiter) undoConsume(i interface{}) {
	l.m.RLock()
	defer l.m.RUnlock()
	if limiters, found := l.stock[i]; found {
		limiters.undoConsume()
	}
}

func (l *limiter) Clean() {
	l.m.Lock()
	defer l.m.Unlock()
lstock:
	for k, v := range l.stock {
		for i, lim := range v.stocks {
			if time.Since(lim.updated).Milliseconds() < (2 * l.rules[i].refill).Milliseconds() {
				break lstock
			}
		}
		delete(l.stock, k)
	}
}

type CFImagerLimiter struct {
}

type CFImagerLimiters map[interface{}]CFImagerLimiter

func NewLimiter(rules ...*limiterRule) Limiter {
	return &limiter{
		stock: make(map[interface{}]*limiterStocks),
		rules: rules,
	}
}

type limiterGroup []limiterInternal

func (l *limiterGroup) Consume(i ...interface{}) bool {
	for j, lim := range *l {
		if !lim.Consume(i[j]) {
			for k := 0; k < j; k++ {
				(*l)[k].undoConsume(i[k])
			}
			return false
		}
	}
	return true
}

func (l *limiterGroup) Clean() {
	for _, lim := range *l {
		lim.Clean()
	}
}

func LimiterGroup(limiters ...Limiter) (Limiter, error) {
	limiterInternals := make(limiterGroup, len(limiters))
	for i, l := range limiters {
		if lInternal, ok := l.(limiterInternal); !ok {
			return nil, errors.New(fmt.Sprintf(
				"cannot use %+v of type %s as %s",
				l,
				reflect.TypeOf(l).String(),
				reflect.TypeOf((*limiterInternal)(nil)).String()),
			)
		} else {
			limiterInternals[i] = lInternal
		}
	}
	return Limiter(&limiterInternals), nil
}
