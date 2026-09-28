package ratelimit_test

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/ratelimit"
)

// clock is a clock a case moves by hand. It stands still until it is moved, so
// that nothing comes back to an allowance a case did not wait for.
type clock struct{ at time.Time }

func newClock() *clock { return &clock{at: time.Date(2026, time.September, 28, 9, 0, 0, 0, time.UTC)} }

func (c *clock) now() time.Time        { return c.at }
func (c *clock) pass(by time.Duration) { c.at = c.at.Add(by) }

// limiter is a limiter of the pace given, keeping at most so many keys, on the
// clock given.
func limiter(t *testing.T, perMinute, keys int, moving *clock) ratelimit.Limiter {
	t.Helper()

	l, err := ratelimit.New(ratelimit.Settings{PerMinute: perMinute, Keys: keys, Now: moving.now})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	return l
}

// allowed is how many of so many requests of the key the limiter let in.
func allowed(l ratelimit.Limiter, key string, requests int) int {
	let := 0
	for range requests {
		if l.Take(key).Allowed {
			let++
		}
	}
	return let
}

// A key may spend a third of a minute's allowance at once, and the request
// past it is refused, told when to ask again. Only the first refusal of a run
// begins it.
func TestAKeyIsLetInForAThirdOfItsMinuteAtOnce(t *testing.T) {
	t.Parallel()

	moving := newClock()
	l := limiter(t, 30, ratelimit.MaxKeys, moving)

	if got := allowed(l, "a-user", 10); got != 10 {
		t.Fatalf("allowed of the first 10 = %d, want all 10", got)
	}
	first := l.Take("a-user")
	if first.Allowed || !first.Began {
		t.Errorf("the 11th = %+v, want refused, beginning a flood", first)
	}
	if first.RetryAfter != 2*time.Second {
		t.Errorf("RetryAfter = %v, want 2s: thirty a minute is one every two seconds", first.RetryAfter)
	}
	if again := l.Take("a-user"); again.Allowed || again.Began {
		t.Errorf("the 12th = %+v, want refused within the run already begun", again)
	}
}

// What a key spent comes back at the pace, one request at a time. A refusal
// soon after another is the same flood, even with a request let in between;
// one after a quiet stretch as long as the allowance takes to fill begins a
// new one.
func TestAnAllowanceComesBackAtThePace(t *testing.T) {
	t.Parallel()

	moving := newClock()
	l := limiter(t, 30, ratelimit.MaxKeys, moving)
	allowed(l, "a-user", 11)

	moving.pass(time.Second)
	if got := l.Take("a-user"); got.Allowed {
		t.Errorf("after a second = %+v, want refused: a request comes back every two seconds", got)
	}
	moving.pass(time.Second)
	if got := l.Take("a-user"); !got.Allowed {
		t.Errorf("after two seconds = %+v, want let in", got)
	}
	if got := l.Take("a-user"); got.Allowed || got.Began {
		t.Errorf("right after = %+v, want refused within the flood already begun", got)
	}

	moving.pass(time.Hour)
	if got := allowed(l, "a-user", 10); got != 10 {
		t.Errorf("after an hour allowed = %d, want 10: the allowance fills back to a third of a minute", got)
	}
	if got := l.Take("a-user"); got.Allowed || !got.Began {
		t.Errorf("the 11th after the quiet hour = %+v, want refused, beginning a new flood", got)
	}
}

// One key over its pace costs no other key anything.
func TestOneKeyOverItsPaceCostsNoOtherKeyAnything(t *testing.T) {
	t.Parallel()

	moving := newClock()
	l := limiter(t, 30, ratelimit.MaxKeys, moving)
	allowed(l, "a-runaway", 100)

	if got := allowed(l, "another-user", 10); got != 10 {
		t.Errorf("allowed of another key = %d, want its whole 10", got)
	}
}

// Past the most keys it keeps, a limiter forgets the key used longest ago, and
// that key starts again from a full allowance; a key used since is kept.
func TestTheKeyUsedLongestAgoIsForgotten(t *testing.T) {
	t.Parallel()

	moving := newClock()
	l := limiter(t, 3, 2, moving)

	l.Take("first")  // spent: the pace of three a minute lets one in at once
	l.Take("second") // spent
	l.Take("first")  // refused; first is now the key used last
	l.Take("third")  // a third key: second, used longest ago, is forgotten
	if got := ratelimit.Kept(t, l); got != 2 {
		t.Errorf("keys kept = %d, want 2", got)
	}
	if got := l.Take("first"); got.Allowed {
		t.Errorf("first = %+v, want still refused: it was used after second, so second went first", got)
	}
	if got := l.Take("second"); !got.Allowed {
		t.Errorf("second after it was forgotten = %+v, want let in from a full allowance", got)
	}
}

// A shared limiter counts every key against one allowance: the instance's.
func TestASharedLimiterCountsEveryKeyTogether(t *testing.T) {
	t.Parallel()

	moving := newClock()
	l, err := ratelimit.NewShared(3, moving.now)
	if err != nil {
		t.Fatalf("NewShared() error = %v, want nil", err)
	}

	if got := l.Take("one-user"); !got.Allowed {
		t.Fatalf("the first request = %+v, want let in", got)
	}
	if got := l.Take("another-user"); got.Allowed || !got.Began {
		t.Errorf("another key's request = %+v, want refused, beginning a run: the allowance is shared", got)
	}
	if got := l.Take("one-user"); got.Allowed || got.Began {
		t.Errorf("the first key again = %+v, want refused within the run already begun", got)
	}
	if got := l.Take("one-user"); got.RetryAfter != 20*time.Second {
		t.Errorf("RetryAfter = %v, want 20s: three a minute is one every twenty seconds", got.RetryAfter)
	}
}

// Requests that arrive at once are each counted once: however many race for
// an allowance, no more of them are let in than it holds.
func TestRequestsAtOnceAreEachCountedOnce(t *testing.T) {
	t.Parallel()

	moving := newClock()
	for _, l := range []ratelimit.Limiter{limiter(t, 30, ratelimit.MaxKeys, moving), shared(t, 30, moving)} {
		var let atomic.Int32
		var racing sync.WaitGroup
		for range 100 {
			racing.Go(func() {
				if l.Take("one-key").Allowed {
					let.Add(1)
				}
			})
		}
		racing.Wait()
		if got := let.Load(); got != 10 {
			t.Errorf("%T let in %d of 100 at once, want 10", l, got)
		}
	}
}

// shared is a limiter every key shares, of the pace given, on the clock given.
func shared(t *testing.T, perMinute int, moving *clock) ratelimit.Limiter {
	t.Helper()

	l, err := ratelimit.NewShared(perMinute, moving.now)
	if err != nil {
		t.Fatalf("NewShared() error = %v, want nil", err)
	}
	return l
}

// Settings nothing could be counted with are refused when the limiter is
// built, not at the first request.
func TestSettingsNothingCouldBeCountedWithAreRefused(t *testing.T) {
	t.Parallel()

	now := newClock().now
	for _, tc := range []struct {
		name     string
		settings ratelimit.Settings
	}{
		{"no request a minute", ratelimit.Settings{PerMinute: 0, Keys: 1, Now: now}},
		{"no key kept", ratelimit.Settings{PerMinute: 1, Keys: 0, Now: now}},
		{"no clock", ratelimit.Settings{PerMinute: 1, Keys: 1}},
	} {
		if _, err := ratelimit.New(tc.settings); !errors.Is(err, ratelimit.ErrSettings) {
			t.Errorf("New() with %s: error = %v, want ErrSettings", tc.name, err)
		}
	}
	if _, err := ratelimit.NewShared(0, now); !errors.Is(err, ratelimit.ErrSettings) {
		t.Errorf("NewShared() with no request a minute: error = %v, want ErrSettings", err)
	}
	if _, err := ratelimit.NewShared(1, nil); !errors.Is(err, ratelimit.ErrSettings) {
		t.Errorf("NewShared() with no clock: error = %v, want ErrSettings", err)
	}
}

// request is one request of a sequence a property plays: whose it is, and how
// long after the one before it arrives.
type request struct {
	key string
	gap time.Duration
}

// sequenceOf zips the keys and the gaps a property was given into requests.
func sequenceOf(keys, gaps []int) []request {
	played := make([]request, min(len(keys), len(gaps)))
	for i := range played {
		played[i] = request{key: "key-" + strconv.Itoa(keys[i]), gap: time.Duration(gaps[i]) * time.Millisecond}
	}
	return played
}

// play takes every request of a sequence, in order and on a clock of its own,
// and says what was decided about each, and when.
func play(t *testing.T, perMinute, keys int, sequence []request) ([]ratelimit.Verdict, []time.Time) {
	t.Helper()

	moving := newClock()
	l := limiter(t, perMinute, keys, moving)
	verdicts := make([]ratelimit.Verdict, len(sequence))
	times := make([]time.Time, len(sequence))
	for i, r := range sequence {
		moving.pass(r.gap)
		times[i] = moving.now()
		verdicts[i] = l.Take(r.key)
	}
	return verdicts, times
}

// Whatever the pace and however requests arrive: no key is let in more often
// than a third of a minute's allowance and the pace since allow; a key that
// keeps to the pace is never refused; no key's requests change what another
// key is told; and a refusal begins a flood exactly when none of the key's
// requests was refused for as long as its allowance takes to fill.
func TestALimiterHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	pace := gen.IntRange(1, 120)
	keys := gen.SliceOf(gen.IntRange(0, 3))
	gaps := gen.SliceOf(gen.IntRange(0, 5000))

	properties.Property("no key is let in more often than its allowance", prop.ForAll(
		func(perMinute int, keys, gaps []int) bool {
			sequence := sequenceOf(keys, gaps)
			verdicts, times := play(t, perMinute, ratelimit.MaxKeys, sequence)
			return withinAllowance(perMinute, sequence, verdicts, times)
		},
		pace, keys, gaps,
	))
	properties.Property("a key that keeps to the pace is never refused", prop.ForAll(
		func(perMinute int, late []int) bool {
			verdicts, _ := play(t, perMinute, ratelimit.MaxKeys, steady(perMinute, late))
			for _, verdict := range verdicts {
				if !verdict.Allowed {
					return false
				}
			}
			return true
		},
		pace, gaps,
	))
	properties.Property("no key's requests change what another key is told", prop.ForAll(
		func(perMinute int, keys, gaps []int) bool {
			sequence := sequenceOf(keys, gaps)
			together, _ := play(t, perMinute, ratelimit.MaxKeys, sequence)
			for _, key := range []string{"key-0", "key-1", "key-2", "key-3"} {
				if !sameAsAlone(t, perMinute, key, sequence, together) {
					return false
				}
			}
			return true
		},
		pace, keys, gaps,
	))
	properties.Property("a refusal begins a flood exactly after a quiet stretch", prop.ForAll(
		func(perMinute int, keys, gaps []int) bool {
			sequence := sequenceOf(keys, gaps)
			verdicts, times := play(t, perMinute, ratelimit.MaxKeys, sequence)
			return floodsBeginAfterQuiet(perMinute, sequence, verdicts, times)
		},
		pace, keys, gaps,
	))
	properties.TestingRun(t)
}

// steady is one key's requests, each at least one request of the pace after
// the one before it, and later by as many milliseconds as late says.
func steady(perMinute int, late []int) []request {
	sequence := make([]request, len(late))
	for i, after := range late {
		sequence[i] = request{key: "steady", gap: time.Minute/time.Duration(perMinute) + time.Duration(after)*time.Millisecond}
	}
	return sequence
}

// floodsBeginAfterQuiet says whether a refusal begins a flood exactly when none
// of its key's requests was refused for as long as the allowance takes to fill
// back whole: a third of a minute's requests, each a minute over the number.
func floodsBeginAfterQuiet(perMinute int, sequence []request, verdicts []ratelimit.Verdict, times []time.Time) bool {
	fill := time.Duration(max(1, perMinute/3)) * (time.Minute / time.Duration(perMinute))
	refused := map[string]time.Time{}
	for i, verdict := range verdicts {
		key := sequence[i].key
		if verdict.Allowed {
			if verdict.Began {
				return false
			}
			continue
		}
		last, seen := refused[key]
		if verdict.Began != (!seen || times[i].Sub(last) >= fill) {
			return false
		}
		refused[key] = times[i]
	}
	return true
}

// withinAllowance says whether, between any two requests of a key, no more of
// its requests were let in than the allowance held at the first and what came
// back since.
func withinAllowance(perMinute int, sequence []request, verdicts []ratelimit.Verdict, times []time.Time) bool {
	burst := float64(max(1, perMinute/3))
	perSecond := float64(perMinute) / 60
	for from := range sequence {
		let := 0
		for to := from; to < len(sequence); to++ {
			if sequence[to].key != sequence[from].key || !verdicts[to].Allowed {
				continue
			}
			let++
			if float64(let) > burst+perSecond*times[to].Sub(times[from]).Seconds()+1e-6 {
				return false
			}
		}
	}
	return true
}

// sameAsAlone says whether a key was told the same, among every other key's
// requests, as it is told when its requests are played alone.
func sameAsAlone(t *testing.T, perMinute int, key string, sequence []request, together []ratelimit.Verdict) bool {
	t.Helper()

	var alone []request
	var told []ratelimit.Verdict
	var carried time.Duration
	for i, r := range sequence {
		carried += r.gap
		if r.key != key {
			continue
		}
		alone = append(alone, request{key: key, gap: carried})
		told = append(told, together[i])
		carried = 0
	}
	replayed, _ := play(t, perMinute, ratelimit.MaxKeys, alone)
	for i := range replayed {
		if replayed[i] != told[i] {
			return false
		}
	}
	return true
}

// However many keys come, a limiter keeps no more of them than it may.
func TestALimiterKeepsNoMoreKeysThanItMay(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("the keys kept never pass the most a limiter may keep", prop.ForAll(
		func(most int, keys []int) bool {
			l := limiter(t, 30, most, newClock())
			for _, key := range keys {
				l.Take("key-" + strconv.Itoa(key))
				if ratelimit.Kept(t, l) > most {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 8), gen.SliceOf(gen.IntRange(0, 50)),
	))
	properties.TestingRun(t)
}

// A request a later ceiling turns away is given back to the ceilings before
// it: the key is not then refused by its own pace for a request that never
// ran. The ceiling named is the one that refused.
func TestARequestTurnedAwayLaterCostsTheEarlierCeilingsNothing(t *testing.T) {
	t.Parallel()

	moving := newClock()
	own := limiter(t, 3, ratelimit.MaxKeys, moving)
	everybody := shared(t, 3, moving)
	everybody.Take("somebody else")

	refusedBy, verdict := ratelimit.Admit("a-user",
		ratelimit.Ceiling{Name: "own", Limiter: own},
		ratelimit.Ceiling{Name: "everybody", Limiter: everybody},
	)
	if verdict.Allowed || refusedBy != "everybody" {
		t.Fatalf("Admit() = %q, %+v, want refused by everybody", refusedBy, verdict)
	}
	if got := own.Take("a-user"); !got.Allowed {
		t.Errorf("the key's own pace after = %+v, want its request given back and let in", got)
	}
}

// Admit stops at the first ceiling that refuses: the ones after it are not
// asked, and spend nothing. A request every ceiling lets in is let in.
func TestAdmitStopsAtTheFirstCeilingThatRefuses(t *testing.T) {
	t.Parallel()

	moving := newClock()
	own := limiter(t, 3, ratelimit.MaxKeys, moving)
	everybody := shared(t, 6, moving)
	ceilings := []ratelimit.Ceiling{{Name: "own", Limiter: own}, {Name: "everybody", Limiter: everybody}}

	if refusedBy, verdict := ratelimit.Admit("a-user", ceilings...); !verdict.Allowed || refusedBy != "" {
		t.Fatalf("the first request: Admit() = %q, %+v, want it let in", refusedBy, verdict)
	}
	if refusedBy, verdict := ratelimit.Admit("a-user", ceilings...); verdict.Allowed || refusedBy != "own" {
		t.Errorf("the second request: Admit() = %q, %+v, want refused by own", refusedBy, verdict)
	}
	if got := everybody.Take("somebody else"); !got.Allowed {
		t.Errorf("everybody after = %+v, want the second of its two requests unspent by a request own refused", got)
	}
}

// A refund never gives a key more than the most it may hold, and a key
// forgotten since it was let in has nothing to give back to: the refund makes
// no key, and so forgets none of the keys kept.
func TestARefundGivesBackNoMoreThanWasTaken(t *testing.T) {
	t.Parallel()

	moving := newClock()
	l := limiter(t, 30, ratelimit.MaxKeys, moving)
	l.Take("a-user")
	l.Refund("a-user")
	l.Refund("a-user")
	if got := allowed(l, "a-user", 11); got != 10 {
		t.Errorf("allowed after two refunds of one request = %d, want the 10 it may hold and no more", got)
	}

	few := limiter(t, 3, 2, moving)
	for _, key := range []string{"forgotten", "kept", "latest"} {
		few.Take(key)
	}
	few.Refund("forgotten")
	if got := few.Take("kept"); got.Allowed {
		t.Errorf("a key kept, after a refund to one forgotten = %+v, want still refused: the refund forgot it", got)
	}
}

// A refund to a shared limiter goes back to the one allowance every key
// shares, whoever's request it was.
func TestASharedLimiterTakesARefundBackForEverybody(t *testing.T) {
	t.Parallel()

	l := shared(t, 3, newClock())
	l.Take("one-user")
	l.Refund("one-user")
	if got := l.Take("another-user"); !got.Allowed {
		t.Errorf("another key after the refund = %+v, want the request given back let in", got)
	}
}

// A clock that goes back gives a key nothing back and takes nothing away: the
// requests it held are still there.
func TestAClockThatGoesBackTakesNothingAway(t *testing.T) {
	t.Parallel()

	moving := newClock()
	l := limiter(t, 30, ratelimit.MaxKeys, moving)
	allowed(l, "a-user", 5)

	moving.pass(-10 * time.Second)
	if got := allowed(l, "a-user", 10); got != 5 {
		t.Errorf("allowed after the clock went back = %d, want the 5 the key still held", got)
	}
}
