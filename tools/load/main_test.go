package main

import (
	"context"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/tools/load/servicetest"
)

// A lesson against a service that answers it as asked ends clean, with the
// report on stdout.
func TestALessonTheServiceAnswersEndsClean(t *testing.T) {
	t.Parallel()

	// Walked without a pause, a lesson is faster than an account's pace allows.
	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600")
	var stdout, stderr strings.Builder
	code := run(t.Context(), []string{
		"-scenario", "lesson", "-url", target.URL, "-host", target.Host, "-pace", "0s", "-tasks", "2",
	}, &stdout, &stderr)

	if code != exitClean {
		t.Fatalf("exit code = %d, want %d; stderr: %s; stdout: %s", code, exitClean, stderr.String(), stdout.String())
	}
	for _, want := range []string{"# Load: lesson", "**Verdict:** clean", "over 2 × accepted task"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("the report has no %q:\n%s", want, stdout.String())
		}
	}
}

// An attack named on the command line hands in the variants named there, each
// followed by its recovery, and ends clean against a service that takes it.
func TestAnAttackRunsTheVariantsItIsGiven(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t,
		"MATHTRAIL_SOLVER_STEPS=200000", "MATHTRAIL_SOLVER_TIMEOUT=300ms", "MATHTRAIL_RATE_USER_PER_MIN=600")
	var stdout, stderr strings.Builder
	code := run(t.Context(), []string{
		"-scenario", "adversarial", "-url", target.URL, "-host", target.Host, "-variants", "loop, product",
		"-rate", "5", "-duration", "500ms", "-children", "1", "-steps", "200000",
	}, &stdout, &stderr)

	if code != exitClean {
		t.Fatalf("exit code = %d, want %d; stderr: %s; stdout: %s", code, exitClean, stderr.String(), stdout.String())
	}
	var headings []string
	for line := range strings.Lines(stdout.String()) {
		if heading, found := strings.CutPrefix(line, "# Load: "); found {
			headings = append(headings, strings.TrimSpace(heading))
		}
	}
	want := []string{"adversarial: loop", "adversarial: loop: recovery", "adversarial: product", "adversarial: product: recovery"}
	if strings.Join(headings, "; ") != strings.Join(want, "; ") {
		t.Errorf("the report has the runs %q, want %q", headings, want)
	}
}

// A lesson the service holds back ends as a run that found something, with
// the report that says what.
func TestALessonTheServiceHoldsBackEndsHard(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=3")
	var stdout, stderr strings.Builder
	code := run(t.Context(), []string{
		"-scenario", "lesson", "-url", target.URL, "-host", target.Host, "-pace", "0s",
	}, &stdout, &stderr)

	if code != exitHard {
		t.Fatalf("exit code = %d, want %d; stderr: %s", code, exitHard, stderr.String())
	}
	if !strings.Contains(stdout.String(), "**Verdict:** failed") {
		t.Errorf("the report does not say the run failed:\n%s", stdout.String())
	}
}

// What cannot run is not run: nothing is sent, the reason is on stderr, and
// the exit code says the run never happened.
func TestWhatCannotRunIsNotRun(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		args []string
		says string
	}{
		{"a scenario nobody wrote", []string{"-scenario", "stampede", "-url", "http://localhost:8080"}, "lesson"},
		{"no scenario at all", []string{"-url", "http://localhost:8080"}, "no scenario"},
		{"no service", []string{"-scenario", "lesson"}, "-url"},
		{"more tasks than written", []string{"-scenario", "lesson", "-url", "http://localhost:8080", "-tasks", "9"}, "tasks"},
		{"a memory with no unit", []string{"-scenario", "lesson", "-url", "http://localhost:8080", "-memory", "512"}, "unit"},
		{"a flag nobody declared", []string{"-scenario", "lesson", "-url", "http://localhost:8080", "-speed", "5"}, "speed"},
		{"no processor", []string{"-scenario", "lesson", "-url", "http://localhost:8080", "-cpus", "0"}, "-cpus"},
		{"a processor that is not a number", []string{"-scenario", "lesson", "-url", "http://localhost:8080", "-cpus", "NaN"}, "-cpus"},
		{"a memory that is not a number", []string{"-scenario", "lesson", "-url", "http://localhost:8080", "-memory", "NaNm"}, "not a size"},
		{"an option the lesson does not read", []string{"-scenario", "lesson", "-url", "http://localhost:8080", "-steps", "1000"}, "lesson reads no -steps"},
		{"an option saturation does not read", []string{"-scenario", "saturation", "-url", "http://localhost:8080", "-variants", "pairs"}, "saturation reads no -variants"},
		{"no rate", []string{"-scenario", "saturation", "-url", "http://localhost:8080", "-rate", "0"}, "rate"},
		{"an attack of no time", []string{"-scenario", "saturation", "-url", "http://localhost:8080", "-duration", "0s"}, "no attack"},
		{"no child", []string{"-scenario", "saturation", "-url", "http://localhost:8080", "-children", "0"}, "children"},
		{"a ceiling no solver is sized to", []string{"-scenario", "adversarial", "-url", "http://localhost:8080", "-steps", "1000"}, "steps"},
		{"a variant nobody wrote", []string{"-scenario", "adversarial", "-url", "http://localhost:8080", "-variants", "pairs,sleep"}, "sleep"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr strings.Builder
			if code := run(t.Context(), test.args, &stdout, &stderr); code != exitCannot {
				t.Errorf("exit code = %d, want %d", code, exitCannot)
			}
			if !strings.Contains(stderr.String(), test.says) {
				t.Errorf("stderr = %q, want it to say %q", stderr.String(), test.says)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want no report of a run that never happened", stdout.String())
			}
		})
	}
}

// A memory is read as docker writes it, in GiB.
func TestAMemoryIsReadAsDockerWritesIt(t *testing.T) {
	t.Parallel()

	for size, want := range map[string]float64{"512m": 0.5, "1g": 1, "1024M": 1, "2G": 2, "524288k": 0.5} {
		if got, err := gibibytes(size); err != nil || got != want {
			t.Errorf("gibibytes(%q) = %v, %v, want %v", size, got, err, want)
		}
	}
	for _, size := range []string{"", "512", "m", "-1g", "0m", "one g", "5\u212a", "NaNm", "Infg"} {
		if got, err := gibibytes(size); err == nil {
			t.Errorf("gibibytes(%q) = %v, want an error", size, got)
		}
	}
}

// A run stopped before its end writes the report of the part that ran, with no
// verdict, and says on stderr that it was stopped: what it did proves nothing
// either way.
func TestAStoppedRunReportsThePartThatRan(t *testing.T) {
	t.Parallel()

	target := servicetest.Start(t, "MATHTRAIL_RATE_USER_PER_MIN=600")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	var stdout, stderr strings.Builder
	code := run(ctx, []string{"-scenario", "lesson", "-url", target.URL, "-host", target.Host}, &stdout, &stderr)

	if code != exitCannot {
		t.Errorf("exit code = %d, want %d", code, exitCannot)
	}
	if !strings.Contains(stdout.String(), "**Stopped** before its end") || strings.Contains(stdout.String(), "**Verdict:**") ||
		!strings.Contains(stderr.String(), "stopped") {
		t.Errorf("stdout %q, stderr %q, want the report of what ran with no verdict and a word that it was stopped",
			stdout.String(), stderr.String())
	}
}
