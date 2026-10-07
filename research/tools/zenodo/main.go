// Command zenodo puts paper A's artifact on Zenodo, a step at a time, each one
// run by a person. draft puts an archive into a draft of the record and says
// where to look at it; the draft can still be thrown away. publish publishes
// that draft, which then cannot be taken back.
//
// Usage:
//
//	zenodo draft -archive <file> -version <name> [-metadata release/zenodo.json] [-record <id>] [-url <base>]
//	zenodo publish -version <name> [-record <id>] [-url <base>]
//
// The access token is read from ZENODO_TOKEN. -version names the paper's
// release the archive comes from. -record is the record's concept id, the
// number every version of it shares; with none, draft starts a record of its
// own. A step does nothing the record already holds: a draft of the same
// version is taken up again rather than made twice, and a version already
// published is left as it is. The exit status is 0 when the step is done, or
// there was nothing to do, and 2 when it could not be done.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const usage = "usage: zenodo draft -archive <file> -version <name> [-metadata <file>] [-record <id>] [-url <base>]\n" +
	"       zenodo publish -version <name> [-record <id>] [-url <base>]"

func main() {
	client := Client{HTTP: &http.Client{Timeout: 10 * time.Minute}, Token: os.Getenv("ZENODO_TOKEN")}
	os.Exit(run(context.Background(), os.Args[1:], client, time.Now, os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, client Client, now func() time.Time, stdout, stderr io.Writer) int {
	if err := step(ctx, args, client, now, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, "zenodo:", err)
		return 2
	}
	return 0
}

func step(ctx context.Context, args []string, client Client, now func() time.Time, stdout, stderr io.Writer) error {
	if len(args) == 0 || (args[0] != "draft" && args[0] != "publish") {
		return errors.New(usage)
	}
	command := args[0]
	flags := flag.NewFlagSet("zenodo "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	base := flags.String("url", "https://zenodo.org", "Zenodo's address, or a sandbox's")
	version := flags.String("version", "", "the name of the paper's release the archive comes from")
	record := flags.String("record", "", "the record's concept id; none starts a record of its own")
	archive := flags.String("archive", "", "the archive the version holds (draft)")
	metadata := flags.String("metadata", "release/zenodo.json", "what the record says of itself (draft)")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	switch {
	case flags.NArg() != 0:
		return errors.New(usage)
	case *version == "":
		return errors.New("-version names no release of the paper")
	case client.Token == "":
		return errors.New("ZENODO_TOKEN holds no token")
	}
	client.URL = strings.TrimSuffix(*base, "/")
	if command == "publish" {
		return Publish(ctx, client, *version, *record, stdout)
	}
	if *archive == "" {
		return errors.New("-archive names no file")
	}
	m, err := ReadMetadata(*metadata)
	if err != nil {
		return err
	}
	return Draft(ctx, client, &DraftRequest{
		Metadata: m, Archive: *archive, Version: *version, Record: *record,
		Today: now().UTC().Format(time.DateOnly),
	}, stdout)
}
