package instance

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
)

// A container is asked for with an instance's processors and memory and no
// swap beside it, its port published on this machine alone, the environment
// every instance has before the spec's own — and the sealing key named, its
// value nowhere among the arguments.
func TestADockerRunAsksForAnInstancesSize(t *testing.T) {
	t.Parallel()

	spec := Spec{Image: "mathtrail:dev", CPUs: 1, Memory: "512m", Env: []string{"MATHTRAIL_SOLVER_CONCURRENCY=1"}}
	want := []string{
		"run", "--detach", "--label", "mathtrail-load=1",
		"--cpus", "1", "--memory", "512m", "--memory-swap", "512m",
		"--publish", "127.0.0.1::8080",
		"--env", "PORT=8080", "--env", "MATHTRAIL_DEV_AUTH=true", "--env", "MATHTRAIL_SEAL_KEY_CURRENT",
		"--env", "MATHTRAIL_SOLVER_CONCURRENCY=1",
		"mathtrail:dev",
	}
	if got := runArgs(&spec); !slices.Equal(got, want) {
		t.Errorf("runArgs() = %q, want %q", got, want)
	}
	spec.CPUs = 0.5
	if got := runArgs(&spec); !slices.Contains(got, "0.5") {
		t.Errorf("runArgs() = %q, want half a processor asked for as 0.5", got)
	}
}

// A spec is refused before anything is started when no instance could be
// started from it, or when it would change what every instance of a load run
// is given.
func TestASpecNoInstanceCouldBeStartedFromIsRefused(t *testing.T) {
	t.Parallel()

	good := Spec{Image: "mathtrail:dev", CPUs: 1, Memory: "512m", Up: time.Minute}
	if err := good.Check(); err != nil {
		t.Fatalf("Check(%+v) = %v, want nil", good, err)
	}
	for name, spoil := range map[string]func(*Spec){
		"no image":                 func(s *Spec) { s.Image = "" },
		"no processor":             func(s *Spec) { s.CPUs = 0 },
		"no memory":                func(s *Spec) { s.Memory = "" },
		"no time to come up":       func(s *Spec) { s.Up = 0 },
		"a variable with no value": func(s *Spec) { s.Env = []string{"MATHTRAIL_LOG_LEVEL"} },
		"a variable with no name":  func(s *Spec) { s.Env = []string{"=debug"} },
		"the port changed":         func(s *Spec) { s.Env = []string{"PORT=9090"} },
		"the sign-in changed":      func(s *Spec) { s.Env = []string{"MATHTRAIL_DEV_AUTH=false"} },
		"the key changed":          func(s *Spec) { s.Env = []string{"MATHTRAIL_SEAL_KEY_CURRENT=abc"} },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			spec := good
			spoil(&spec)
			if err := spec.Check(); err == nil {
				t.Errorf("Check(%+v) = nil, want the spec refused", spec)
			}
		})
	}
}

// The service is reached on the port docker published, under the host it
// knows itself by: the one of its public URL, when the spec gives it one.
func TestTheServiceIsReachedWhereDockerPublishedIt(t *testing.T) {
	t.Parallel()

	if port, err := portOf("127.0.0.1:32768\n[::1]:32768\n"); err != nil || port != "32768" {
		t.Errorf("portOf() = %q, %v, want 32768", port, err)
	}
	for name, published := range map[string]string{"nothing": "", "no port": "127.0.0.1\n", "an empty port": "127.0.0.1:\n"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if port, err := portOf(published); err == nil {
				t.Errorf("portOf(%q) = %q, want an error", published, port)
			}
		})
	}
	if host := publicHost(nil); host != "localhost:8080" {
		t.Errorf("publicHost() = %q, want localhost:8080", host)
	}
	if host := publicHost([]string{"MATHTRAIL_PUBLIC_URL=https://load.example:8443"}); host != "load.example:8443" {
		t.Errorf("publicHost() = %q, want the host of the public URL", host)
	}
}

// A container's process and state are read as docker inspect writes them.
func TestAContainerIsReadAsDockerInspectWritesIt(t *testing.T) {
	t.Parallel()

	pid, began, err := process("1059954 2026-09-29T00:02:32.951931289Z\n")
	if err != nil || pid != 1059954 || began.Nanosecond() != 951931289 {
		t.Errorf("process() = %d, %v, %v, want 1059954 begun at .951931289", pid, began, err)
	}
	for name, inspected := range map[string]string{
		"nothing": "", "no moment": "1059954", "no process": "x 2026-09-29T00:02:32Z", "no time": "1 yesterday",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, failed := process(inspected); failed == nil {
				t.Errorf("process(%q) = no error, want one", inspected)
			}
		})
	}
	s, err := stateOf(`{"Status":"exited","Running":false,"OOMKilled":true,"ExitCode":137}`)
	if err != nil || s != (state{Running: false, OOMKilled: true, ExitCode: 137}) {
		t.Errorf("stateOf() = %+v, %v, want an exited container killed for its memory", s, err)
	}
}

// A container that could not be reached is told with the last lines it wrote,
// when it wrote any: a service that ended as it started says why there.
func TestAContainerThatCouldNotBeReachedIsToldWithItsLastLines(t *testing.T) {
	t.Parallel()

	reached := errors.New("instance: docker port: no public port published")
	said := withLastWords(reached, `{"severity":"ERROR","message":"config: MATHTRAIL_PUBLIC_URL is no URL"}`+"\n")
	if !errors.Is(said, reached) || !strings.Contains(said.Error(), "MATHTRAIL_PUBLIC_URL is no URL") {
		t.Errorf("withLastWords() = %v, want the error with the container's last lines", said)
	}
	if said := withLastWords(reached, " \n"); said.Error() != reached.Error() {
		t.Errorf("withLastWords(nothing written) = %v, want the error as it was", said)
	}
}

// The cgroup is the path of the line of the unified hierarchy, as docker's
// own driver names it and as systemd's does; a process under the first
// version of cgroups is refused, since its counters are laid out otherwise.
func TestTheCgroupIsTheLineOfTheUnifiedHierarchy(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, listed, want string
		version1           bool
	}{
		{"docker's driver", "0::/docker/89fae65a4123\n", "/docker/89fae65a4123", false},
		{"systemd's driver", "0::/system.slice/docker-89fae65a4123.scope\n", "/system.slice/docker-89fae65a4123.scope", false},
		{"the first version", "12:memory:/docker/89fae65a4123\n11:cpu,cpuacct:/docker/89fae65a4123\n", "", true},
		{"both, as a hybrid", "12:memory:/docker/89fae65a4123\n0::/docker/89fae65a4123\n", "", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			path, err := cgroupOf(test.listed)
			if test.version1 != errors.Is(err, errVersion1) || path != test.want {
				t.Errorf("cgroupOf() = %q, %v, want %q, the first version refused: %v", path, err, test.want, test.version1)
			}
		})
	}
	if path, err := cgroupOf(""); err == nil {
		t.Errorf("cgroupOf(nothing) = %q, want an error", path)
	}
}

// The cgroup read is the container's own: a listing that names another
// container, or cgroups of the first version, leave the instance unmeasured
// rather than measured by some other process.
func TestOnlyTheContainersOwnCgroupIsRead(t *testing.T) {
	t.Parallel()

	const id = "89fae65a41238e38545578dd011c01802bdcda5afe497deb01cbf793a620305f"
	if dir := cgroupDir("0::/docker/"+id+"\n", id); dir != "/sys/fs/cgroup/docker/"+id {
		t.Errorf("cgroupDir(its own) = %q, want its directory", dir)
	}
	for name, listed := range map[string]string{
		"another container's": "0::/docker/0123456789ab\n",
		"the machine's own":   "0::/init\n",
		"the first version":   "12:memory:/docker/" + id + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if dir := cgroupDir(listed, id); dir != "" {
				t.Errorf("cgroupDir() = %q, want none", dir)
			}
		})
	}
}

// A cgroup is read from its files: the processor time from cpu.stat, the
// memory it holds and the most it held, and the kills from memory.events. A
// kernel that keeps no peak leaves the memory held now as the most.
func TestACgroupIsReadFromItsFiles(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		dir  string
		peak uint64
	}{
		{"testdata/cgroup", 35934208},
		{"testdata/cgroup-with-no-peak", 35577856},
	} {
		t.Run(test.dir, func(t *testing.T) {
			t.Parallel()

			s, err := read(test.dir)
			if err != nil {
				t.Fatalf("read() error = %v", err)
			}
			if s.cpu != 59126*time.Microsecond || s.current != 35577856 || s.peak != test.peak || s.oomKills != 1 {
				t.Errorf("read() = %+v, want the counters of the files, a peak of %d", s, test.peak)
			}
		})
	}
	if _, err := read(t.TempDir()); err == nil {
		t.Error("read(an empty directory) = no error, want one: its instance went with it")
	}
}

// What an instance spent between two moments is its processor time from the
// last reading at or before the first to the last at or before the second, and
// the most memory the readings between the two saw — a reading before them
// being what an earlier run left. A stretch the readings do not stand round —
// one shorter than their pace, or before the first — is one they cannot tell.
func TestWhatAnInstanceSpentIsCountedBetweenItsReadings(t *testing.T) {
	t.Parallel()

	at := func(ms int) time.Time { return time.Unix(0, 0).Add(time.Duration(ms) * time.Millisecond) }
	samples := []sample{
		{at: at(0), cpu: 0, current: 10},
		{at: at(250), cpu: 100 * time.Millisecond, current: 40},
		{at: at(500), cpu: 300 * time.Millisecond, current: 20},
		{at: at(750), cpu: 350 * time.Millisecond, current: 90},
	}
	for _, test := range []struct {
		name     string
		from, to int
		want     report.Spent
		told     bool
	}{
		{"the whole life", 0, 750, report.Spent{CPU: 350 * time.Millisecond, Memory: 90}, true},
		{"between readings", 300, 600, report.Spent{CPU: 200 * time.Millisecond, Memory: 20}, true},
		{"from before the first reading", -100, 250, report.Spent{CPU: 100 * time.Millisecond, Memory: 40}, true},
		{"shorter than a reading's pace", 260, 270, report.Spent{}, false},
		{"before any reading", -500, -100, report.Spent{}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got, told := spentBetween(samples, at(test.from), at(test.to)); told != test.told || got != test.want {
				t.Errorf("spentBetween() = %+v, %v, want %+v, %v", got, told, test.want, test.told)
			}
		})
	}
	if got, told := spentBetween(nil, at(0), at(750)); told {
		t.Errorf("spentBetween(no reading) = %+v, want nothing to tell", got)
	}
}

// The service's log tells a load run how many of its lines tell of a panic —
// its own lines that carry one, and the runtime's first line when one ends the
// process, the lines it indents after it not counted again — and every run of
// its sandbox.
func TestTheLogTellsItsPanicsAndItsSolverRuns(t *testing.T) {
	t.Parallel()

	written := strings.Join([]string{
		`{"severity":"INFO","message":"startup","port":"8080"}`,
		`{"severity":"INFO","message":"solver_run","status":"ok","steps":23750047,"duration_ms":96}`,
		`{"severity":"INFO","message":"solver_run","status":"timeout","steps":100011,"duration_ms":2001}`,
		`{"severity":"ERROR","message":"mcp_panic","method":"tools/call","panic":"runtime error: index out of range","stack":"..."}`,
		`{"severity":"ERROR","message":"tool_call","error":"panic","panic":"a value of type string"}`,
		`panic: runtime error: invalid memory address or nil pointer dereference [recovered]`,
		"\tpanic: a panic raised again as the first unwound",
		`goroutine 1 [running]:`,
		`fatal error: concurrent map writes`,
		`{"severity":"INFO","message":"shutdown"}`,
		`not json and not a crash`,
	}, "\n")
	panics, runs := readLog(written)
	want := []report.SolverRun{
		{Status: "ok", Steps: 23750047, Took: 96 * time.Millisecond},
		{Status: "timeout", Steps: 100011, Took: 2001 * time.Millisecond},
	}
	if panics != 4 || !slices.Equal(runs, want) {
		t.Errorf("readLog() = %d panics, runs %+v; want 4 and %+v", panics, runs, want)
	}
}

// The life of an instance is what its cgroup counted over it — its processor
// time, its peak, the kills in it — how it stood before it was stopped, and
// what its log told.
func TestTheLifeOfAnInstanceIsItsCountsItsEndingAndItsLog(t *testing.T) {
	t.Parallel()

	samples := []sample{{cpu: time.Second, peak: 100, oomKills: 0}, {cpu: 3 * time.Second, peak: 300, oomKills: 1}}
	written := `{"message":"solver_run","status":"ok","steps":10,"duration_ms":5}` + "\n"
	life := lifeOf(samples, state{Running: false, OOMKilled: true, ExitCode: 137}, written, 200)
	want := report.Life{
		Measured: true, CPU: 3 * time.Second, Peak: 300, OOMKills: 1, OOMKilled: true, Exited: true, ExitCode: 137,
		SolverRuns: []report.SolverRun{{Status: "ok", Steps: 10, Took: 5 * time.Millisecond}}, Ceiling: 200,
	}
	if life.Measured != want.Measured || life.CPU != want.CPU || life.Peak != want.Peak || life.OOMKills != want.OOMKills ||
		life.OOMKilled != want.OOMKilled || life.Exited != want.Exited || life.ExitCode != want.ExitCode ||
		life.Ceiling != want.Ceiling || !slices.Equal(life.SolverRuns, want.SolverRuns) {
		t.Errorf("lifeOf() = %+v, want %+v", life, want)
	}
	if unmeasured := lifeOf(nil, state{Running: true}, "", 0); unmeasured.Measured || unmeasured.Exited {
		t.Errorf("lifeOf(no reading, running) = %+v, want an instance unmeasured and stopped by the run", unmeasured)
	}
}

// A container is taken as ended only when docker says so: one docker could
// say nothing of — a daemon slow to answer, or failing a moment — is taken as
// running, so that nothing still up is given up on.
func TestAContainerIsTakenAsEndedOnlyWhenDockerSaysSo(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		state   state
		err     error
		running bool
	}{
		{"running", state{Running: true}, nil, true},
		{"ended", state{Running: false, ExitCode: 1}, nil, false},
		{"nothing to say", state{}, errors.New("instance: docker inspect: context deadline exceeded"), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := running(test.state, test.err); got != test.running {
				t.Errorf("running(%+v, %v) = %v, want %v", test.state, test.err, got, test.running)
			}
		})
	}
}

// An instance that ended by itself keeps the last lines it wrote, those that
// say anything, as many as are told; one the run stopped keeps none.
func TestAnInstanceThatEndedKeepsItsLastLines(t *testing.T) {
	t.Parallel()

	written := "one\n\ntwo\nthree\r\nfour\nfive\nsix\n  \n"
	if life := lifeOf(nil, state{Running: false, ExitCode: 2}, written, 0); life.LastLines != "two\nthree\nfour\nfive\nsix" {
		t.Errorf("lifeOf(ended).LastLines = %q, want the last five lines that say anything", life.LastLines)
	}
	if life := lifeOf(nil, state{Running: true}, written, 0); life.LastLines != "" {
		t.Errorf("lifeOf(stopped by the run).LastLines = %q, want none", life.LastLines)
	}
}

// Every docker command runs in a group of processes of its own, so that an
// interrupt typed at the terminal reaches the tool, and not a command that
// has to finish for what it started to be removed.
func TestEveryDockerCommandRunsInAGroupOfItsOwn(t *testing.T) {
	t.Parallel()

	command := dockerCommand(t.Context(), "run", "--detach", "mathtrail:dev")
	if command.SysProcAttr == nil || !command.SysProcAttr.Setpgid {
		t.Errorf("dockerCommand().SysProcAttr = %+v, want a group of processes of its own", command.SysProcAttr)
	}
}
