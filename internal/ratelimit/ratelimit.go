// Package ratelimit holds requests to the pace an instance lets them in at,
// counted in the instance's own memory.
//
// Nothing here is shared between instances: with several of them running, a
// key may be let in as many times more often as there are instances. That is
// accepted. The limits are a ceiling on cost and a fuse against a runaway, not
// an account of anybody's use, and counting them in memory is what lets the
// service keep no store of its own.
package ratelimit

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"
)

// MaxKeys is the most keys a limiter keeps a count for. A limiter that could
// grow without bound would itself be a way to bring the instance down, so past
// it the key used longest ago is forgotten. Its next request starts from a full
// allowance: forgetting a key can only ever let a request in, never keep one
// out.
const MaxKeys = 4096

// burstShare is the part of a minute's allowance a key may spend at once after
// a pause: a third. A lesson's first calls, or the few requests of a sign-in,
// then go through together, and a runaway is still held to the minute's pace
// within seconds.
const burstShare = 3

// Limiter decides whether one more request may go ahead now.
type Limiter interface {
	// Take counts one request of the key given against the key's allowance,
	// and says whether the request may go ahead.
	Take(key string) Verdict
	// Refund gives back the request a Take let in, for a request that did not
	// go ahead after all.
	Refund(key string)
}

// Verdict is what a limiter decided about one request.
type Verdict struct {
	// Allowed is whether the request may go ahead.
	Allowed bool
	// Began marks the first refusal of a flood: none of the key's requests was
	// refused for as long as its allowance takes to fill back whole. However
	// long a flood lasts, and however many of its requests the pace lets
	// through, it is one event, to be written down once.
	Began bool
	// RetryAfter is how long a refused request should wait before it asks
	// again: the time one request of the allowance takes to come back, which
	// is at least as long as the wait for the next one.
	RetryAfter time.Duration
}

// Ceiling is one pace a request is held to, under the name a line gives it.
type Ceiling struct {
	Name    string
	Limiter Limiter
}

// Admit holds one request of the key to each ceiling in turn, and names the
// one that refused it, if one did. The request is given back to every ceiling
// before the one that refused, so that a request that does not go ahead costs
// no allowance: a key the instance turned away is not then refused by its own
// pace for requests that never ran.
func Admit(key string, ceilings ...Ceiling) (refusedBy string, verdict Verdict) {
	for i, ceiling := range ceilings {
		verdict = ceiling.Limiter.Take(key)
		if verdict.Allowed {
			continue
		}
		for _, passed := range ceilings[:i] {
			passed.Limiter.Refund(key)
		}
		return ceiling.Name, verdict
	}
	return "", Verdict{Allowed: true}
}

// Settings are what a limiter is built from.
type Settings struct {
	// PerMinute is how many requests one key may make in a minute.
	PerMinute int
	// Keys is the most keys the limiter keeps a count for: MaxKeys in the
	// service, fewer where a test has to see one forgotten.
	Keys int
	// Now is the clock the allowance comes back by.
	Now func() time.Time
}

// ErrSettings is returned when a limiter is given settings nothing could be
// counted with; callers branch on it with errors.Is.
var ErrSettings = errors.New("ratelimit: settings")

// New is a limiter that counts every key apart: one key over its pace is
// refused, and no other key is any the worse for it.
func New(settings Settings) (Limiter, error) {
	if settings.Keys < 1 {
		return nil, fmt.Errorf("%w: a limiter keeps at least one key, got %d", ErrSettings, settings.Keys)
	}
	pace, err := paceOf(settings.PerMinute, settings.Now)
	if err != nil {
		return nil, err
	}
	return &keyed{
		pace:       pace,
		now:        settings.Now,
		keys:       settings.Keys,
		recent:     list.New(),
		allowances: make(map[string]*allowance, settings.Keys),
	}, nil
}

// NewShared is a limiter every key shares one allowance of: a ceiling on the
// whole instance rather than on anybody in it. The key a request is taken for
// is not read.
func NewShared(perMinute int, now func() time.Time) (Limiter, error) {
	pace, err := paceOf(perMinute, now)
	if err != nil {
		return nil, err
	}
	return &shared{pace: pace, now: now, allowance: pace.full(now())}, nil
}

// pace is how often a key's requests come back, and how many of them a key
// may hold at once.
type pace struct {
	// interval is what one request costs: the time it takes to come back.
	interval time.Duration
	// most is the most a key may hold: a third of a minute's requests, in
	// the time they take to come back.
	most time.Duration
}

// paceOf is the pace of so many requests a minute, refusing a number nothing
// could be let in at and a missing clock.
func paceOf(perMinute int, now func() time.Time) (pace, error) {
	switch {
	case perMinute < 1:
		return pace{}, fmt.Errorf("%w: a limiter lets in at least one request a minute, got %d", ErrSettings, perMinute)
	case now == nil:
		return pace{}, fmt.Errorf("%w: a limiter needs a clock", ErrSettings)
	}
	interval := time.Minute / time.Duration(perMinute)
	return pace{interval: interval, most: time.Duration(max(1, perMinute/burstShare)) * interval}, nil
}

// full is an allowance nothing has been spent of yet.
func (p pace) full(now time.Time) *allowance {
	return &allowance{held: p.most, at: now}
}

// allowance is what a key holds, as of a moment, and when a request of the key
// was last refused. It is held as time: a request costs one interval, and every
// moment that passes gives a moment back, up to the most a key may hold. Kept
// in whole nanoseconds, a request that comes exactly one interval after the
// last one always finds its interval there.
type allowance struct {
	held      time.Duration
	at        time.Time
	refusedAt time.Time
	// place is where the key stands in the order the keys were last used in.
	place *list.Element
}

// take spends one request of the allowance when there is one to spend.
func (a *allowance) take(now time.Time, p pace) Verdict {
	a.catchUp(now, p)
	if a.held >= p.interval {
		a.held -= p.interval
		return Verdict{Allowed: true}
	}
	began := a.refusedAt.IsZero() || now.Sub(a.refusedAt) >= p.most
	a.refusedAt = now
	return Verdict{Began: began, RetryAfter: p.interval}
}

// refund gives one request back, never past the most a key may hold.
func (a *allowance) refund(p pace) {
	a.held = min(p.most, a.held+p.interval)
}

// catchUp gives back the time that passed since the allowance was last looked
// at. A clock that has gone back gives nothing back, and takes nothing away.
func (a *allowance) catchUp(now time.Time, p pace) {
	if !now.After(a.at) {
		return
	}
	if passed := now.Sub(a.at); passed < p.most-a.held {
		a.held += passed
	} else {
		a.held = p.most
	}
	a.at = now
}

// keyed keeps an allowance for each key it has seen lately. recent holds the
// keys in the order they were last used in, the latest at the front, so that
// the key to forget is always at its back.
type keyed struct {
	mu         sync.Mutex
	pace       pace
	now        func() time.Time
	keys       int
	recent     *list.List
	allowances map[string]*allowance
}

func (k *keyed) Take(key string) Verdict {
	k.mu.Lock()
	defer k.mu.Unlock()
	now := k.now()
	return k.allowanceOf(key, now).take(now, k.pace)
}

// Refund gives the key its request back. A key forgotten since it was let in
// has nothing to give back to: it starts from a full allowance anyway.
func (k *keyed) Refund(key string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if kept, seen := k.allowances[key]; seen {
		kept.refund(k.pace)
	}
}

// allowanceOf is the key's allowance, a full one for a key not seen lately.
// Making room for it forgets the key used longest ago, whose allowance has had
// the longest to fill back.
func (k *keyed) allowanceOf(key string, now time.Time) *allowance {
	if kept, seen := k.allowances[key]; seen {
		k.recent.MoveToFront(kept.place)
		return kept
	}
	if k.recent.Len() >= k.keys {
		if oldest, isKey := k.recent.Remove(k.recent.Back()).(string); isKey {
			delete(k.allowances, oldest)
		}
	}
	fresh := k.pace.full(now)
	fresh.place = k.recent.PushFront(key)
	k.allowances[key] = fresh
	return fresh
}

// shared is one allowance for every key.
type shared struct {
	mu        sync.Mutex
	pace      pace
	now       func() time.Time
	allowance *allowance
}

func (s *shared) Take(string) Verdict {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.allowance.take(s.now(), s.pace)
}

func (s *shared) Refund(string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.allowance.refund(s.pace)
}
