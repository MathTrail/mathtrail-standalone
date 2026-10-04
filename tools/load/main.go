// Command load drives the MathTrail service with a scenario of a load run, and
// writes what the run came to as Markdown: every kind of answer the service
// gave and how long each took, what one unit of the run's work cost as the
// platform bills it against its free tier and what the instance spent of it,
// and the facts that fail the run.
//
// Usage:
//
//	load -scenario lesson -image mathtrail:dev [flags]
//	load -scenario lesson -url http://localhost:8080 [flags]
//	load -signin parent -url https://mcp.mathtrail.app
//	load -scenario paces -url https://mcp.mathtrail.app -accounts parent,load [flags]
//
// With an image, the run starts the service itself, in a container of an
// instance's size, and reads what the instance spent and how it ended. With a
// URL, it runs against a service somebody else started, which has to run with
// the development sign-in, as `just run` runs it: the children of a run sign
// in through it, each by a name of its own. A deployed service has no such
// sign-in, so a parent signs accounts in on it once with -signin, in a
// browser, and the children of the runs after sign in with the accounts that
// -accounts names.
//
// It exits 0 when the run found nothing the service should never do, 1 when
// it found something, and 2 when it could not run, or was stopped before its
// end — the report of the part that ran is written all the same.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/account"
	"github.com/MathTrail/mathtrail-standalone/tools/load/instance"
	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/scenario"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// The ways the command ends.
const (
	exitClean  = 0
	exitHard   = 1
	exitCannot = 2
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// run is the command, from its arguments to its exit code: the report goes to
// stdout, and what went wrong before there was one to stderr. While an account
// signs in, the address a browser was sent back to may be pasted on stdin.
func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	asked, err := parse(args, stderr)
	switch {
	case errors.Is(err, errUsage):
		return exitCannot
	case err != nil:
		return cannot(stderr, err)
	}
	if asked.signIn != "" {
		return signIn(ctx, &asked, stdin, stderr)
	}
	if unsigned := asked.loadAccounts(stderr); unsigned != nil {
		return cannot(stderr, unsigned)
	}
	defer asked.tellUnkept(stderr)

	fmt.Fprintf(stderr, "load: %s against %s, about %s\n", asked.options.Scenario, asked.against, asked.options.Lasts())
	runs, failed := scenario.Run(ctx, &asked.options, asked.launch)
	if len(runs) == 0 && failed != nil {
		return cannot(stderr, failed)
	}
	hards, err := report.Write(stdout, runs, asked.instance)
	switch {
	case err != nil:
		return cannot(stderr, err)
	case failed != nil:
		// What ran is reported, and what could not go on is said.
		return cannot(stderr, fmt.Errorf("%w; the report is of the part that ran", failed))
	case slices.ContainsFunc(runs, func(r report.Run) bool { return r.Stopped }):
		// What ran is reported, but a run cut short proves nothing either way.
		return cannot(stderr, errors.New("the run was stopped before its end; the report is of the part that ran"))
	case len(hards) > 0:
		return exitHard
	}
	return exitClean
}

// errUsage is a command line the flags could not read, which they have said
// why of already.
var errUsage = errors.New("load: usage")

// invocation is what the command line asks for: the scenario and its options,
// the services it is run against, and the instance the platform bills for.
type invocation struct {
	options  scenario.Options
	launch   scenario.Launch
	instance report.Instance
	// against is what the run is against, in words.
	against string
	// signIn is the account to sign in on the service at the URL, when that
	// is all the command line asks for, and port the port of this computer
	// the browser comes back to; accounts is where accounts are kept.
	signIn, url, accounts string
	port                  int
	// loaded are the accounts the run signs in with, once their tokens are read.
	loaded []*account.Account
}

// parse reads the command line.
func parse(args []string, stderr io.Writer) (invocation, error) {
	flags := flag.NewFlagSet("load", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("scenario", "", "the scenario to run: "+strings.Join(scenario.Names(), ", "))
	service := declareService(flags)
	given := declareOptions(flags)
	cpus := flags.Float64("cpus", 1, "the vCPUs an instance has, and is billed for")
	memory := flags.String("memory", "1g", "the memory an instance has, and is billed for, as docker writes it: 1g, 512m")
	accounts := declareAccounts(flags)
	if err := flags.Parse(args); err != nil {
		return invocation{}, errUsage
	}
	if *accounts.signIn != "" {
		return accounts.signingIn(flags, *service.address)
	}

	options, err := scenario.Defaults(*name)
	if err != nil {
		return invocation{}, err
	}
	if wrong := accounts.into(flags, &options, *service.address); wrong != nil {
		return invocation{}, wrong
	}
	if unread := given.change(flags, &options); len(unread) > 0 {
		return invocation{}, fmt.Errorf("%s reads no %s", options.Scenario, strings.Join(unread, ", "))
	}
	if !(*cpus > 0) || math.IsInf(*cpus, 0) {
		return invocation{}, fmt.Errorf("-cpus %v is no share of a processor an instance could have", *cpus)
	}
	gib, err := gibibytes("-memory", *memory)
	if err != nil {
		return invocation{}, err
	}
	if invalid := options.Check(); invalid != nil {
		return invocation{}, invalid
	}
	launch, against, err := service.launch(&options, *cpus, *memory)
	if err != nil {
		return invocation{}, err
	}
	if len(options.Accounts) > 0 {
		against += " as " + strings.Join(namesOf(options.Accounts), ", ")
	}
	return invocation{
		options: options, launch: launch, instance: report.Instance{VCPU: *cpus, GiB: gib}, against: against,
		url: *service.address, accounts: *accounts.dir,
	}, nil
}

// serviceFlags are the flags that name the service a run goes against: one
// somebody else started, at a URL, or one the run starts itself, from an
// image.
type serviceFlags struct {
	address, host, image, ceiling *string
	env                           environment
}

// declareService declares the flags that name the service a run goes against.
func declareService(flags *flag.FlagSet) *serviceFlags {
	s := &serviceFlags{
		address: flags.String("url", "", "where a service somebody else started is reached, such as http://localhost:8080"),
		host:    flags.String("host", "", "the host the service at -url knows itself by, when the URL names it otherwise"),
		image: flags.String("image", "", "the image to start the service from, in a container of the instance's size, "+
			"such as mathtrail:dev"),
		ceiling: flags.String("memory-ceiling", "", "the most memory an instance started from -image may hold "+
			"without failing the run, as docker writes a size: 400m"),
	}
	flags.Var(&s.env, "env", "a variable NAME=VALUE the service started from -image has; given again for another")
	return s
}

// launch is how the service the flags name is launched for the scenario of
// the options, on an instance of the processors and the memory given, and the
// service in words.
func (s *serviceFlags) launch(options *scenario.Options, cpus float64, memory string) (scenario.Launch, string, error) {
	switch {
	case *s.image != "" && *s.address != "":
		return nil, "", errors.New("-image and -url are two services: run against one")
	case *s.image != "":
		return s.fromImage(options, cpus, memory)
	case *s.address != "":
		if scenario.Starts(options.Scenario) {
			return nil, "", fmt.Errorf("%s starts the service itself, from -image, and cannot run against -url", options.Scenario)
		}
		if len(s.env) > 0 || *s.ceiling != "" {
			return nil, "", errors.New("-env and -memory-ceiling are for a service started from -image")
		}
		return scenario.At(session.Target{URL: *s.address, Host: *s.host}), *s.address, nil
	}
	return nil, "", errors.New("-image or -url names no service: which does the run go against?")
}

// fromImage is how a service started from the image the flags name is
// launched, in a container of the processors and the memory given, and the
// service in words.
func (s *serviceFlags) fromImage(options *scenario.Options, cpus float64, memory string) (scenario.Launch, string, error) {
	if *s.host != "" {
		return nil, "", errors.New("-host is for a service at -url: one started from -image knows its own")
	}
	ceiling, err := bytesOf("-memory-ceiling", *s.ceiling)
	if err != nil {
		return nil, "", err
	}
	held, _ := bytesOf("-memory", memory)
	if ceiling >= held && ceiling > 0 {
		// The instance is killed at its memory before it could hold more.
		return nil, "", fmt.Errorf("-memory-ceiling %s is no lower than the instance's -memory %s, and could never be passed",
			*s.ceiling, memory)
	}
	env := deployed(cpus, held, s.env)
	spec := instance.Spec{Image: *s.image, CPUs: cpus, Memory: memory, Env: env, Ceiling: ceiling, Up: options.Timeout}
	if invalid := spec.Check(); invalid != nil {
		return nil, "", invalid
	}
	return scenario.FromImage(&spec), "an instance of " + *s.image + " with " + strings.Join(sized(env), " "), nil
}

// The variables a deployment derives from the size of an instance rather than
// takes as given: a solver slot for every whole processor, and nine tenths of
// the memory as the runtime's soft limit.
const (
	solverSlots = "MATHTRAIL_SOLVER_CONCURRENCY"
	softLimit   = "GOMEMLIMIT"
)

// sizedNames are those variables, by name.
var sizedNames = []string{solverSlots, softLimit}

// deployed is the environment an instance of the processors and the memory
// given starts with: what a deployment derives from that size, then the
// variables given, which take the place of a derived one they name.
func deployed(cpus float64, memory uint64, given environment) []string {
	derived := []struct{ name, value string }{
		{solverSlots, strconv.Itoa(max(1, int(cpus)))},
		{softLimit, strconv.FormatUint(memory*9/10>>20, 10) + "MiB"},
	}
	env := make([]string, 0, len(derived)+len(given))
	for _, variable := range derived {
		if !slices.ContainsFunc(given, func(v string) bool { return strings.HasPrefix(v, variable.name+"=") }) {
			env = append(env, variable.name+"="+variable.value)
		}
	}
	return append(env, given...)
}

// sized is what an environment gives of the variables a deployment derives
// from the size of an instance.
func sized(env []string) []string {
	var found []string
	for _, variable := range env {
		if name, _, _ := strings.Cut(variable, "="); slices.Contains(sizedNames, name) {
			found = append(found, variable)
		}
	}
	return found
}

// environment is the variables given with -env, each NAME=VALUE.
type environment []string

func (e *environment) String() string { return strings.Join(*e, " ") }

func (e *environment) Set(variable string) error {
	*e = append(*e, variable)
	return nil
}

// bytesOf is a size as docker writes one, in bytes, and zero for no size at
// all.
func bytesOf(flagName, size string) (uint64, error) {
	if size == "" {
		return 0, nil
	}
	gib, err := gibibytes(flagName, size)
	if err != nil {
		return 0, err
	}
	return uint64(math.Round(gib * (1 << 30))), nil
}

// optionFlags are the flags that change the options of a scenario, each left
// as the scenario has it unless given.
type optionFlags struct {
	tasks, children         *int
	pace, timeout, duration *time.Duration
	rate                    *float64
	variants                *string
	steps                   *uint64
}

// declareOptions declares the flags that change the options of a scenario.
func declareOptions(flags *flag.FlagSet) *optionFlags {
	return &optionFlags{
		tasks: flags.Int("tasks", 0, "how many tasks a lesson asks for, when not the scenario's own"),
		pace:  flags.Duration("pace", 0, "the pause before every step of a lesson, when not the scenario's own"),
		timeout: flags.Duration("timeout", 0, "how long one call may take, and the service to come back after an "+
			"attack, when not the scenario's own"),
		rate: flags.Float64("rate", 0, "how many calls a second an attack sends: hand-ins, or the calls of the "+
			"greedy child and address of limits, each; when not the scenario's own"),
		duration: flags.Duration("duration", 0, "how long an attack goes on, each variant's in adversarial, when "+
			"not the scenario's own"),
		children: flags.Int("children", 0, "how many children share an attack's hand-ins, or walk lessons beside "+
			"the greedy child of limits, when not the scenario's own"),
		variants: flags.String("variants", "", "the costly solvers adversarial hands in, one after another, of "+
			strings.Join(lesson.Costly(), ", ")+"; when not the scenario's own"),
		steps: flags.Uint64("steps", 0, "the step ceiling of the service, which the costly solvers are sized to, "+
			"when not the scenario's own"),
	}
}

// change changes the options the flags given on the command line name, and
// only those: every other is the scenario's. It is the flags among them the
// scenario does not read, which are refused rather than left be, so that
// nobody takes a run for one it was not.
func (f *optionFlags) change(flags *flag.FlagSet, options *scenario.Options) (unread []string) {
	flags.Visit(func(given *flag.Flag) {
		switch given.Name {
		case "tasks":
			options.Tasks = *f.tasks
		case "pace":
			options.Pace = *f.pace
		case "timeout":
			options.Timeout = *f.timeout
		case "rate":
			options.Rate = *f.rate
		case "duration":
			options.Duration = *f.duration
		case "children":
			options.Children = *f.children
		case "variants":
			options.Variants = strings.Split(*f.variants, ",")
			for i := range options.Variants {
				options.Variants[i] = strings.TrimSpace(options.Variants[i])
			}
		case "steps":
			options.Steps = *f.steps
		default:
			return
		}
		if !slices.Contains(scenario.Reads(options.Scenario), given.Name) {
			unread = append(unread, "-"+given.Name)
		}
	})
	return unread
}

// cannot says why the command could not run, and is the exit code that says
// so.
func cannot(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "load: %v\n", err)
	return exitCannot
}

// gibibytes reads a size as docker writes one — a number and one of the units
// b, k, m and g, each 1024 times the one before — in GiB, and names the flag
// given when the size is none.
func gibibytes(flagName, size string) (float64, error) {
	units := map[byte]float64{'b': 1 << 30, 'k': 1 << 20, 'm': 1 << 10, 'g': 1}
	if size == "" {
		return 0, fmt.Errorf("%s names no size", flagName)
	}
	unit := size[len(size)-1]
	if 'A' <= unit && unit <= 'Z' {
		unit += 'a' - 'A'
	}
	per, known := units[unit]
	if !known {
		return 0, fmt.Errorf("%s %q has no unit of b, k, m or g", flagName, size)
	}
	amount, err := strconv.ParseFloat(size[:len(size)-1], 64)
	// Written the other way round, a size that is not a number would pass.
	if err != nil || !(amount > 0) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("%s %q is not a size", flagName, size)
	}
	return amount / per, nil
}
