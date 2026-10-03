package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// task is a reference task ready to be reviewed as the service reviews a task
// handed in, compared with the reference tasks other than itself, as the
// experiment with injected defects reviews it.
type task struct {
	id       string
	prepared *reviewing.Prepared
}

// accepted are the reference tasks the checks accept as they are, with every
// check run — the tasks of the experiment with injected defects — prepared
// for review. Preparing a task, which writes it out as JSON and leaves its own
// question out of the comparison, is the harness's work and is not timed.
func accepted(ctx context.Context, shipped *content.Content, runner solver.Runner) ([]task, error) {
	var tasks []task
	for _, h := range reviewing.Hosts(shipped) {
		prepared, err := reviewing.Prepare(runner, reviewing.Without(shipped, h.Base.Task.Question), &h.Base)
		if err != nil {
			return nil, fmt.Errorf("perf: prepare %s: %w", h.ID, err)
		}
		t := task{id: h.ID, prepared: prepared}
		r, err := review(ctx, &t)
		if err != nil {
			return nil, err
		}
		if verdict := reviewing.VerdictOf(&r.outcome, r.runs); verdict.Refused() || len(verdict.Unchecked) > 0 {
			continue
		}
		tasks = append(tasks, t)
	}
	if len(tasks) == 0 {
		return nil, errors.New("perf: the checks accept none of the reference tasks")
	}
	return tasks, nil
}

// reviewed is one review of a task: what it decided, the runs of the solver,
// and how long each half took.
type reviewed struct {
	outcome        checks.Outcome
	runs           []checks.Run
	examine, judge time.Duration
}

// review is one review of a task, as the service runs it when a task is
// handed in, with its two halves timed apart: the one that reads the task and
// runs its solver, and the one that judges it.
func review(ctx context.Context, t *task) (reviewed, error) {
	start := time.Now()
	examined, err := t.prepared.Examine(ctx)
	examine := time.Since(start)
	if err != nil {
		return reviewed{}, fmt.Errorf("perf: %s: %w", t.id, err)
	}
	start = time.Now()
	outcome, err := t.prepared.Judge(examined)
	judge := time.Since(start)
	if err != nil {
		return reviewed{}, fmt.Errorf("perf: %s: %w", t.id, err)
	}
	return reviewed{outcome: outcome, runs: examined.Runs(), examine: examine, judge: judge}, nil
}

// timing is one review of one task, its two halves timed apart.
type timing struct {
	task           string
	pass           int
	examine, judge time.Duration
}

// solverRun is one run of a task's solver within a review, as the sandbox
// reports it.
type solverRun struct {
	task        string
	pass, index int
	status      solver.Status
	steps       uint64
	took        time.Duration
}

// timeReviews reviews every task once in each of a number of passes, one
// review at a time, and keeps every review's timings and solver runs.
func timeReviews(ctx context.Context, tasks []task, passes int) ([]timing, []solverRun, error) {
	timings := make([]timing, 0, len(tasks)*passes)
	var solverRuns []solverRun
	for pass := range passes {
		for i := range tasks {
			r, err := review(ctx, &tasks[i])
			if err != nil {
				return nil, nil, err
			}
			timings = append(timings, timing{task: tasks[i].id, pass: pass, examine: r.examine, judge: r.judge})
			for index, run := range r.runs {
				solverRuns = append(solverRuns, solverRun{task: tasks[i].id, pass: pass, index: index, status: run.Status, steps: run.Steps, took: run.Duration})
			}
		}
	}
	return timings, solverRuns, nil
}

// benchmarkPasses is how many times the benchmark reviews every task: enough
// that what one review allocates is an average over several passes, as few as
// keep the run short.
const benchmarkPasses = "3x"

// cost is what one review allocates, as Go's benchmark machinery counts it.
type cost struct {
	passes        int
	bytes, allocs int64
}

// reviewCost benchmarks whole reviews and says what one allocates on average.
// One operation of the benchmark reviews every task once, so that the average
// is over all of them rather than over however many came first.
func reviewCost(ctx context.Context, tasks []task) (cost, error) {
	var failed error
	result := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			for i := range tasks {
				if _, err := review(ctx, &tasks[i]); err != nil {
					failed = err
					return
				}
			}
		}
	})
	if failed != nil {
		return cost{}, failed
	}
	return perReview(&result, len(tasks)), nil
}

// perReview is what one review allocates, from a benchmark whose every
// operation reviewed every task once.
func perReview(result *testing.BenchmarkResult, reviews int) cost {
	n := int64(reviews)
	return cost{passes: result.N, bytes: result.AllocedBytesPerOp() / n, allocs: result.AllocsPerOp() / n}
}

// packageAt is the package the model is handed at one point of the ladder,
// over every turn of its reference tasks: the smallest and the largest, and
// every size, each turn counted once.
type packageAt struct {
	topic             string
	point             rating.Point
	turns             int
	sizes, chars      []int
	smallest, largest int
	largestChars      int
}

// packageDay is the day the profiles packages are built for are made on.
var packageDay = time.Date(2026, time.October, 1, 9, 0, 0, 0, time.UTC)

// packages builds the package at every point of the ladder every topic is
// taught at, for a new child of the first grade of the point's level with no
// interests and no notes: the brief as the rule writes it when the model asks
// for that point, and the corridor around it. The reference tasks a package
// shows take turns by the child's count of answers, so each point is built at
// every count of a whole cycle of the turns. A parent's notes and a child's
// interests lengthen a package by their own length.
func packages(shipped *content.Content) ([]packageAt, error) {
	cycle := turnCycle(shipped)
	var found []packageAt
	for _, topic := range shipped.TopicIDs() {
		for _, point := range rating.Points(shipped.LevelsOf(topic)...) {
			at, err := packageOf(shipped, topic, point, cycle)
			if err != nil {
				return nil, err
			}
			found = append(found, at)
		}
	}
	return found, nil
}

// packageOf is the package at one point, built at every count of answers of a
// cycle.
func packageOf(shipped *content.Content, topic string, point rating.Point, cycle int) (packageAt, error) {
	p := profile.New(profile.Student{Grade: point.GradeLevel.FirstGrade(), Pseudonym: "perf"}, "perf", packageDay)
	choice := tutor.Choice{Topic: topic, GradeLevel: point.GradeLevel, Difficulty: point.Difficulty, Reason: "measured at every point"}
	brief, _, err := tutor.Next(p, shipped, choice)
	if err != nil {
		return packageAt{}, fmt.Errorf("perf: brief for %s at %s/%d: %w", topic, point.GradeLevel, point.Difficulty, err)
	}
	request := content.Request{Language: reviewing.Language, Brief: brief, Corridor: tutor.CorridorIn(p, shipped, topic), Grade: p.Student.Grade}
	at := packageAt{topic: topic, point: point, turns: cycle}
	for answers := range cycle {
		request.Answers = answers
		pack, err := shipped.Package(&request)
		if err != nil {
			return packageAt{}, fmt.Errorf("perf: package for %s at %s/%d: %w", topic, point.GradeLevel, point.Difficulty, err)
		}
		size, chars := len(pack), utf8.RuneCount(pack)
		at.sizes, at.chars = append(at.sizes, size), append(at.chars, chars)
		if answers == 0 || size < at.smallest {
			at.smallest = size
		}
		if size > at.largest {
			at.largest, at.largestChars = size, chars
		}
	}
	return at, nil
}

// turnCycle is a count of answers after which the reference tasks of every
// package have taken all their turns: the least common multiple of every
// count of reference tasks one topic has at one level and difficulty, since
// those are what take turns.
func turnCycle(shipped *content.Content) int {
	type group struct {
		topic      string
		level      rating.GradeLevel
		difficulty int
	}
	sizes := map[group]int{}
	examples := shipped.Examples()
	for i := range examples {
		sizes[group{examples[i].Topic, examples[i].GradeLevel, examples[i].Difficulty}]++
	}
	cycle := 1
	for _, size := range sizes {
		cycle = cycle / gcd(cycle, size) * size
	}
	return cycle
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
