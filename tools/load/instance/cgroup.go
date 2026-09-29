package instance

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
)

// cgroupRoot is where this machine mounts the unified hierarchy of cgroups.
const cgroupRoot = "/sys/fs/cgroup"

// errVersion1 is a process whose cgroups are of the first version, whose
// counters are laid out otherwise and are not read here.
var errVersion1 = errors.New("instance: the cgroups are of the first version")

// cgroupOf is the path of a process's cgroup in the unified hierarchy, read
// from its /proc/<pid>/cgroup: the line of hierarchy 0 with no controllers
// named, as it stands for docker's own driver and for systemd's alike. A
// process that names a controller on any line is under the first version.
func cgroupOf(listed string) (string, error) {
	var path string
	for line := range strings.Lines(listed) {
		hierarchy, rest, found := strings.Cut(strings.TrimSpace(line), ":")
		if !found {
			continue
		}
		controllers, where, found := strings.Cut(rest, ":")
		if !found {
			continue
		}
		if hierarchy != "0" || controllers != "" {
			return "", errVersion1
		}
		path = where
	}
	if path == "" {
		return "", fmt.Errorf("instance: no cgroup of the unified hierarchy in %q", listed)
	}
	return path, nil
}

// sample is one reading of an instance's cgroup.
type sample struct {
	at time.Time
	// cpu is the processor time the instance used from its start.
	cpu time.Duration
	// current and peak are the memory it holds now and the most it ever
	// held, in bytes.
	current, peak uint64
	// oomKills is how many processes the kernel killed in it for want of
	// memory.
	oomKills int
}

// read reads the counters of the cgroup in the directory given.
func read(dir string) (sample, error) {
	s := sample{at: time.Now()}
	usage, err := field(filepath.Join(dir, "cpu.stat"), "usage_usec")
	if err != nil {
		return sample{}, err
	}
	s.cpu = time.Duration(min(usage, math.MaxInt64/uint64(time.Microsecond))) * time.Microsecond
	if s.current, err = number(filepath.Join(dir, "memory.current")); err != nil {
		return sample{}, err
	}
	if s.peak, err = peakOf(dir); err != nil {
		return sample{}, err
	}
	kills, err := field(filepath.Join(dir, "memory.events"), "oom_kill")
	if err != nil {
		return sample{}, err
	}
	s.oomKills = int(min(kills, math.MaxInt32))
	return s, nil
}

// peakOf is the most memory the cgroup in the directory given has held: its
// peak, or, where the kernel keeps none, what it holds now — the most of
// every reading is then the most there is to tell.
func peakOf(dir string) (uint64, error) {
	peak, err := number(filepath.Join(dir, "memory.peak"))
	if errors.Is(err, fs.ErrNotExist) {
		return number(filepath.Join(dir, "memory.current"))
	}
	return peak, err
}

// number reads a file of one number.
func number(path string) (uint64, error) {
	written, err := os.ReadFile(path) //nolint:gosec // a file of the cgroup the tool's own container runs in
	if err != nil {
		return 0, fmt.Errorf("instance: read %s: %w", path, err)
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(written)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("instance: read %s: %w", path, err)
	}
	return value, nil
}

// field reads the number of one key in a file of lines of a key and a number.
func field(path, key string) (uint64, error) {
	file, err := os.Open(path) //nolint:gosec // a file of the cgroup the tool's own container runs in
	if err != nil {
		return 0, fmt.Errorf("instance: read %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()
	lines := bufio.NewScanner(file)
	for lines.Scan() {
		name, value, found := strings.Cut(lines.Text(), " ")
		if found && name == key {
			number, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
			if err != nil {
				return 0, fmt.Errorf("instance: read %s of %s: %w", key, path, err)
			}
			return number, nil
		}
	}
	if err := lines.Err(); err != nil {
		return 0, fmt.Errorf("instance: read %s: %w", path, err)
	}
	return 0, fmt.Errorf("instance: %s has no %s", path, key)
}

// spentBetween is what the samples say the instance spent between two
// moments: the processor time it used from the last reading at or before the
// first moment — or the first reading of all — to the last at or before the
// second, and the most memory the readings between the two moments saw, a
// reading before them being what an earlier run left. It is false when the
// readings cannot tell: when not two of them stand that way round the
// stretch, as for one shorter than the pace they are taken at.
func spentBetween(samples []sample, from, to time.Time) (report.Spent, bool) {
	first, last := -1, -1
	for i := range samples {
		if samples[i].at.After(to) {
			break
		}
		if first < 0 || !samples[i].at.After(from) {
			first = i
		}
		last = i
	}
	if first < 0 || first == last {
		return report.Spent{}, false
	}
	spent := report.Spent{CPU: samples[last].cpu - samples[first].cpu}
	for _, s := range samples[first : last+1] {
		if !s.at.Before(from) {
			spent.Memory = max(spent.Memory, s.current)
		}
	}
	return spent, true
}
