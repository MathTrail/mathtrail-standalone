package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/account"
	"github.com/MathTrail/mathtrail-standalone/tools/load/scenario"
)

// renewBefore is how much longer than a run the tokens of its accounts are to
// last when it starts. Tokens that would run out sooner are renewed before the
// run, never in its middle: a renewal is held to the pace of the address it
// comes from, which a run may be spending on purpose.
const renewBefore = 2 * time.Minute

// The flags of accounts, by name.
const (
	flagSignIn       = "signin"
	flagURL          = "url"
	flagAccounts     = "accounts"
	flagAccountsDir  = "accounts-dir"
	flagCallbackPort = "callback-port"
)

// accountFlags are the flags of accounts a parent signed in for real: the one
// to sign in now, or those the children of a run sign in with, and where they
// are kept.
type accountFlags struct {
	signIn, names, dir *string
	port               *int
}

// declareAccounts declares the flags of accounts.
func declareAccounts(flags *flag.FlagSet) *accountFlags {
	return &accountFlags{
		signIn: flags.String(flagSignIn, "", "sign the account of this name in on the service at -url, in a browser, "+
			"and keep it for the runs after; nothing else is run"),
		names: flags.String(flagAccounts, "", "the accounts the children of a run against the deployed service at -url "+
			"sign in with, in order, as name,name; each signed in before with -signin"),
		dir: flags.String(flagAccountsDir, "", "where accounts are kept, when not in the user's own configuration"),
		port: flags.Int(flagCallbackPort, 0, "the port of this computer the browser comes back to when -signin signs "+
			"an account in; any free one when not given"),
	}
}

// signingIn is the command line that signs an account in on the service at
// the URL given, and does nothing else: no other flag but where the account
// is kept and the port the browser comes back to.
func (a *accountFlags) signingIn(flags *flag.FlagSet, url string) (invocation, error) {
	var others []string
	flags.Visit(func(given *flag.Flag) {
		if !slices.Contains([]string{flagSignIn, flagURL, flagAccountsDir, flagCallbackPort}, given.Name) {
			others = append(others, "-"+given.Name)
		}
	})
	switch {
	case len(others) > 0:
		return invocation{}, fmt.Errorf("-signin signs an account in and runs nothing: %s go with a run",
			strings.Join(others, ", "))
	case url == "":
		return invocation{}, errors.New("-signin signs an account in on the service at -url, which names none")
	case *a.port < 0 || *a.port > 65535:
		return invocation{}, fmt.Errorf("-callback-port %d is no port", *a.port)
	}
	return invocation{signIn: *a.signIn, url: url, accounts: *a.dir, port: *a.port}, nil
}

// into gives the options the accounts the flags name, by name: their tokens
// are read once the command line is known to be one that can run.
func (a *accountFlags) into(flags *flag.FlagSet, options *scenario.Options, url string) error {
	given := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { given[f.Name] = true })
	switch {
	case given[flagCallbackPort]:
		return errors.New("-callback-port is for -signin")
	case given[flagAccountsDir] && !given[flagAccounts]:
		return errors.New("-accounts-dir is where the accounts of -accounts or -signin are kept, and neither is given")
	case !given[flagAccounts]:
		return nil
	case !slices.Contains(scenario.Reads(options.Scenario), flagAccounts):
		return fmt.Errorf("%s reads no -accounts: its children are made up by a development sign-in", options.Scenario)
	case url == "":
		return errors.New("-accounts are signed in on a deployed service, which -url names")
	case given["children"]:
		return errors.New("-accounts are the children of the run, and -children goes with no -accounts")
	}
	for name := range strings.SplitSeq(*a.names, ",") {
		name = strings.TrimSpace(name)
		if name == "" || slices.ContainsFunc(options.Accounts, func(known scenario.Account) bool { return known.Name == name }) {
			return fmt.Errorf("-accounts %q names each account once, as name,name", *a.names)
		}
		options.Accounts = append(options.Accounts, scenario.Account{Name: name})
	}
	return nil
}

// namesOf are the names of the accounts given.
func namesOf(accounts []scenario.Account) []string {
	names := make([]string, len(accounts))
	for i := range accounts {
		names[i] = accounts[i].Name
	}
	return names
}

// accountsDir is where the accounts of the command line are kept.
func (in *invocation) accountsDir() (string, error) {
	if in.accounts != "" {
		return in.accounts, nil
	}
	return account.Dir()
}

// loadAccounts reads the tokens of the accounts the run signs in with, each
// good for the length of the run and a margin.
func (in *invocation) loadAccounts() error {
	if len(in.options.Accounts) == 0 {
		return nil
	}
	dir, err := in.accountsDir()
	if err != nil {
		return err
	}
	for i := range in.options.Accounts {
		loaded, err := account.Load(dir, in.options.Accounts[i].Name, in.url, in.options.Lasts()+renewBefore)
		if err != nil {
			return err
		}
		in.options.Accounts[i].Token = loaded.Token
	}
	return nil
}

// signIn signs the account the command line names in on its service, in the
// parent's browser, and keeps it.
func signIn(ctx context.Context, in *invocation, stdin io.Reader, stderr io.Writer) int {
	dir, err := in.accountsDir()
	if err != nil {
		return cannot(stderr, err)
	}
	browser := account.Browser{
		Show: func(address string) {
			fmt.Fprintf(stderr, "load: open this address in a browser, sign in with the Google account %s stands for "+
				"and allow the load:\n\n%s\n\nload: if the browser cannot come back to this computer, paste here the "+
				"address it ends up at.\n", in.signIn, address)
		},
		Pasted: stdin,
	}
	signedIn, err := account.SignIn(ctx, dir, in.signIn, in.url, browser, in.port)
	if err != nil {
		return cannot(stderr, err)
	}
	fmt.Fprintf(stderr, "load: %s is signed in on %s and kept in %s\n", signedIn.Name(), in.url, dir)
	return exitClean
}
