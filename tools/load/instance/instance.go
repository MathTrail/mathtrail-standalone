// Package instance runs the service in a container of an instance's size — as
// many processors and as much memory as the platform gives one instance, and
// no swap, since it gives none — and watches what the instance spends through
// its cgroup, until it stops the container and reads what became of it: how it
// ended, and the lines it wrote. It drives the docker command, which has to
// reach a daemon on this machine.
package instance

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// Spec is the instance a run's service is given.
type Spec struct {
	// Image is the image of the service.
	Image string
	// CPUs and Memory are the processors and the memory the instance may use,
	// the memory as docker writes a size, such as 512m.
	CPUs   float64
	Memory string
	// Env are the variables, each NAME=VALUE, the service's environment has on
	// top of those every instance of a load run has.
	Env []string
	// Ceiling is the most memory, in bytes, the instance may hold without
	// failing the run; zero is none.
	Ceiling uint64
	// Up is how long the service has to answer its probe once started.
	Up time.Duration
}

// Check refuses a spec no instance could be started from.
func (s *Spec) Check() error {
	switch {
	case s.Image == "":
		return errors.New("instance: no image to start")
	case !(s.CPUs > 0):
		return fmt.Errorf("instance: %v processors is no share of one", s.CPUs)
	case s.Memory == "":
		return errors.New("instance: no memory to start with")
	case s.Up <= 0:
		return fmt.Errorf("instance: a service has no time at all to come up: %v", s.Up)
	}
	for _, variable := range s.Env {
		name, _, found := strings.Cut(variable, "=")
		switch {
		case !found || name == "":
			return fmt.Errorf("instance: %q is no NAME=VALUE", variable)
		case slices.Contains(reserved, name):
			return fmt.Errorf("instance: %s is every instance's own, and not the spec's to change", name)
		}
	}
	return nil
}

// The pace of the watching.
const (
	// sampleEvery is how often the cgroup is read while the instance runs.
	sampleEvery = 250 * time.Millisecond
	// probeEvery is how often a service coming up is asked whether it is.
	probeEvery = 20 * time.Millisecond
	// aliveEvery is how often, while it comes up, its container is asked
	// whether it still runs: one that ended will not come up however long it
	// is waited for.
	aliveEvery = time.Second
	// inspectWithin is how long docker has to say how a container stands.
	inspectWithin = 5 * time.Second
	// stopWithin is how long a container has to be stopped, read and removed:
	// the service's own way out, the platform's ten seconds, and the reading.
	stopWithin = 30 * time.Second
	// stopAfter is how long docker lets the service stop by itself before it
	// kills it, as the platform does.
	stopAfter = "10"
)

// Container is the service started in a container of its own.
type Container struct {
	// Target is where the service is reached, and the host it knows itself by.
	Target session.Target
	// Asked is when the container was asked for, Began when its process began,
	// and Healthy when the service first answered its probe, zero when it never
	// did.
	Asked, Began, Healthy time.Time

	id      string
	ceiling uint64
	// cgroup is the directory of the instance's cgroup, empty when it cannot
	// be read from here.
	cgroup string

	mu      sync.Mutex
	samples []sample
	// quit tells the sampler to stop, and quitted says it has.
	quit, quitted chan struct{}
}

// Start starts a container of the spec, and is the container once its service
// answered its probe or its time to do so ran out. The error is for a
// container that could not be started at all. A service that started and
// never answered — or ended as it started — is a container all the same,
// with Healthy left zero, so that how it ended and what it wrote can still be
// read: it is the caller's to stop either way.
func Start(ctx context.Context, spec *Spec) (*Container, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("instance: make a sealing key: %w", err)
	}
	// An image not on this machine is pulled first, in the run's own time,
	// which a run asked to stop ends: a pull cut short leaves nothing behind.
	if _, err := docker(ctx, nil, "image", "inspect", "--format", "{{.Id}}", spec.Image); err != nil {
		if _, err := docker(ctx, nil, "pull", spec.Image); err != nil {
			return nil, fmt.Errorf("instance: pull %s: %w", spec.Image, err)
		}
	}
	c := &Container{Asked: time.Now(), ceiling: spec.Ceiling, quit: make(chan struct{}), quitted: make(chan struct{})}
	// The container is asked for in a time of its own rather than the run's:
	// a run asked to stop in the middle of it would end the command after the
	// daemon had made the container, and leave it with nobody to remove it.
	asking, cancel := context.WithTimeout(context.WithoutCancel(ctx), stopWithin)
	defer cancel()
	id, err := docker(asking, []string{sealKey + "=" + base64.StdEncoding.EncodeToString(key)}, runArgs(spec)...)
	if err != nil {
		return nil, fmt.Errorf("instance: start a container of %s: %w", spec.Image, err)
	}
	c.id = strings.TrimSpace(id)
	if err := ctx.Err(); err != nil {
		c.remove()
		return nil, fmt.Errorf("instance: start a container of %s: %w", spec.Image, err)
	}
	if err := c.locate(ctx, spec); err != nil {
		err = c.lastWords(err)
		c.remove()
		return nil, err
	}
	go c.sample()
	if c.Target.URL != "" {
		c.Healthy = c.comeUp(ctx, spec.Up)
	}
	return c, nil
}

// lastWords is the error of a container that could not be reached, with the
// last lines it wrote: a service that ended as it started says why there.
func (c *Container) lastWords(err error) error {
	ctx, cancel := context.WithTimeout(context.Background(), stopWithin)
	defer cancel()
	written, logErr := logOf(ctx, c.id, "--tail", strconv.Itoa(lastLineCount))
	if logErr != nil {
		return err
	}
	return withLastWords(err, written)
}

// withLastWords is an error with the lines given after it, when there are any.
func withLastWords(err error, written string) error {
	written = strings.TrimSpace(written)
	if written == "" {
		return err
	}
	return fmt.Errorf("%w; the container's last lines:\n%s", err, written)
}

// Alive says whether the container's service still runs. It is false only
// when docker says the container ended — one that ended will not answer
// however long it is waited for — and a docker that cannot say takes it as
// running.
func (c *Container) Alive() bool {
	ctx, cancel := context.WithTimeout(context.Background(), inspectWithin)
	defer cancel()
	s, err := c.state(ctx)
	return running(s, err)
}

// running says whether a container runs, by what docker said of it: one it
// could say nothing of is taken as running.
func running(s state, err error) bool { return err != nil || s.Running }

// locate finds when the container's process began, where its service is
// reached, and its cgroup. A container that ended as it started publishes no
// port: it is a service that never came up, reached nowhere, and its life
// tells how it ended. The cgroup is looked for and not required: a daemon
// whose processes this machine does not see, or cgroups of the first version,
// leave the instance unmeasured rather than unrun.
func (c *Container) locate(ctx context.Context, spec *Spec) error {
	inspected, err := docker(ctx, nil, "inspect", "--format", "{{.State.Pid}} {{.State.StartedAt}}", c.id)
	if err != nil {
		return err
	}
	pid, began, err := process(inspected)
	if err != nil {
		return err
	}
	c.Began = began

	published, err := docker(ctx, nil, "port", c.id, servicePort+"/tcp")
	if err != nil {
		if s, stateErr := c.state(ctx); stateErr == nil && !s.Running {
			return nil
		}
		return err
	}
	port, err := portOf(published)
	if err != nil {
		return err
	}
	c.Target = session.Target{URL: "http://127.0.0.1:" + port, Host: publicHost(spec.Env)}
	if listed, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid)); err == nil {
		c.cgroup = cgroupDir(string(listed), c.id)
	}
	return nil
}

// cgroupDir is the directory of the cgroup the listing of a process names,
// when that is the container's, and nothing otherwise. A process of another
// machine's daemon would be some other process here: its cgroup names the
// container, or it is not the container's.
func cgroupDir(listed, id string) string {
	path, err := cgroupOf(listed)
	if err != nil || !strings.Contains(path, id) {
		return ""
	}
	return filepath.Join(cgroupRoot, path)
}

// comeUp asks the service's probe until it answers, for as long as given or
// until its container ends, and is when it first answered, zero when it never
// did.
func (c *Container) comeUp(ctx context.Context, within time.Duration) time.Time {
	probe := &http.Client{Timeout: time.Second}
	until := time.Now().Add(within)
	lastAlive := time.Now()
	for time.Now().Before(until) && ctx.Err() == nil {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Target.URL+"/health", http.NoBody)
		if err != nil {
			return time.Time{}
		}
		if response, err := probe.Do(request); err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return time.Now()
			}
		}
		if time.Since(lastAlive) >= aliveEvery {
			if !c.Alive() {
				return time.Time{}
			}
			lastAlive = time.Now()
		}
		time.Sleep(probeEvery)
	}
	return time.Time{}
}

// sample reads the cgroup as the instance runs, until told to stop. A cgroup
// that cannot be read is not read again: it went with its container.
func (c *Container) sample() {
	defer close(c.quitted)
	if c.cgroup == "" {
		return
	}
	ticker := time.NewTicker(sampleEvery)
	defer ticker.Stop()
	for c.keep() {
		select {
		case <-c.quit:
			return
		case <-ticker.C:
		}
	}
}

// keep reads the cgroup once more and keeps the reading, and says whether it
// could.
func (c *Container) keep() bool {
	s, err := read(c.cgroup)
	if err != nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.samples = append(c.samples, s)
	return true
}

// Spent is what the instance spent between two moments, as its cgroup counted
// it, and false when nothing counted it then.
func (c *Container) Spent(from, to time.Time) (report.Spent, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return spentBetween(c.samples, from, to)
}

// Peak is the most memory the instance has held so far, read from its cgroup
// now, and false when it cannot be read.
func (c *Container) Peak() (uint64, bool) {
	if c.cgroup == "" {
		return 0, false
	}
	peak, err := peakOf(c.cgroup)
	return peak, err == nil
}

// Stop stops the container and removes it, and is the life of its instance.
// What can only be read while the container is there is read before it goes,
// in this order: the last count of its cgroup, which goes with it, and its
// state, which says whether it ended before it was stopped. Its log is read
// after it has stopped, so that its last lines are in it. It is removed
// whatever else fails, and a run asked to stop still stops it.
func (c *Container) Stop(ctx context.Context) (report.Life, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), stopWithin)
	defer cancel()
	defer c.remove()

	close(c.quit)
	<-c.quitted
	if c.cgroup != "" {
		c.keep()
	}
	s, err := c.state(ctx)
	if err != nil {
		return report.Life{}, err
	}
	if s.Running {
		if _, err = docker(ctx, nil, "stop", "--timeout", stopAfter, c.id); err != nil {
			return report.Life{}, err
		}
	}
	written, err := logOf(ctx, c.id)
	if err != nil {
		return report.Life{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return lifeOf(c.samples, s, written, c.ceiling), nil
}

// state is what docker says of the container as it stands.
func (c *Container) state(ctx context.Context) (state, error) {
	inspected, err := docker(ctx, nil, "inspect", "--format", "{{json .State}}", c.id)
	if err != nil {
		return state{}, err
	}
	return stateOf(inspected)
}

// remove removes the container, whatever state it is in, with a time of its
// own: it is the last thing a run does to it, and a run may already be over.
func (c *Container) remove() {
	ctx, cancel := context.WithTimeout(context.Background(), stopWithin)
	defer cancel()
	_, _ = docker(ctx, nil, "rm", "--force", c.id)
}

// lifeOf is the life of an instance: what its cgroup counted over it, how it
// stood before it was stopped, and what its log told.
func lifeOf(samples []sample, s state, written string, ceiling uint64) report.Life {
	life := report.Life{
		Measured: len(samples) > 0, OOMKilled: s.OOMKilled, Exited: !s.Running, ExitCode: s.ExitCode, Ceiling: ceiling,
	}
	for _, reading := range samples {
		life.CPU = max(life.CPU, reading.cpu)
		life.Peak = max(life.Peak, reading.peak)
		life.OOMKills = max(life.OOMKills, reading.oomKills)
	}
	life.Panics, life.SolverRuns = readLog(written)
	if life.Exited {
		life.LastLines = lastOf(written, lastLineCount)
	}
	return life
}

// lastLineCount is how many of its last lines an instance that ended by
// itself is told by.
const lastLineCount = 5

// lastOf is the last lines of the text that say anything, as many as given.
func lastOf(written string, count int) string {
	var lines []string
	for line := range strings.Lines(written) {
		if line = strings.TrimRight(line, "\r\n"); strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines[max(len(lines)-count, 0):], "\n")
}
