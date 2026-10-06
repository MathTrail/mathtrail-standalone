package analytics_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// read is a file of the repository, by its path from this package's folder.
func read(t *testing.T, file string) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.FromSlash(file))
	if err != nil {
		t.Fatalf("read %s: %v", file, err)
	}
	return string(raw)
}

// blockOf is the block of HCL that opens with header and closes with the first
// brace at the start of a line after it.
func blockOf(t *testing.T, hcl, header string) string {
	t.Helper()

	start := strings.Index(hcl, header)
	if start < 0 {
		t.Fatalf("no block opens with %s", header)
	}
	end := strings.Index(hcl[start:], "\n}\n")
	if end < 0 {
		t.Fatalf("the block %s never closes", header)
	}
	return hcl[start : start+end]
}

// The snapshot is kept where it is read: the scheduled query Terraform makes
// runs the snapshot's SQL and replaces with its row the table impact_site.live;
// the recipe reads that table, and refuses it once the query has stopped
// writing it; and the identity that reads it is lent to the one environment the
// workflow reads it in. The files are held to one another with no emulator, so
// every run of the tests holds them.
func TestTheSnapshotIsKeptWhereItIsRead(t *testing.T) {
	t.Parallel()

	snapshot := read(t, "../terraform/analytics/snapshot.tf")
	query := blockOf(t, snapshot, `resource "google_bigquery_data_transfer_config" "snapshot" {`)
	dataset := blockOf(t, snapshot, `resource "google_bigquery_dataset" "site" {`)
	justfile := read(t, "../../justfile")
	for _, want := range []struct{ what, in, pattern string }{
		{"the scheduled query runs the snapshot's SQL", query, `query\s*=\s*templatefile\("\$\{local\.analytics\}/site/live\.sql", local\.names\)`},
		{"the scheduled query writes into the snapshot's dataset", query, `destination_dataset_id\s*=\s*google_bigquery_dataset\.site\.dataset_id`},
		{"the scheduled query writes the table live", query, `destination_table_name_template\s*=\s*"live"`},
		{"the scheduled query replaces the table whole", query, `write_disposition\s*=\s*"WRITE_TRUNCATE"`},
		{"the snapshot's dataset is impact_site", dataset, `dataset_id\s*=\s*"impact_site"`},
		{"the recipe reads impact_site.live", justfile, `table="\$project:impact_site\.live"[\s\S]*?head --max_rows=1 "\$table"`},
		{"the recipe refuses a table not written for two days", justfile, `show "\$table"[\s\S]*?lastModifiedTime[\s\S]*?> 2 \* 24 \* 60 \* 60`},
		{"the reader is lent to the environment live-numbers", read(t, "../terraform/analytics.tf"), `:environment:live-numbers"`},
		{"the workflow reads in the environment live-numbers", read(t, "../../.github/workflows/live.yml"), `(?m)^\s+environment: live-numbers$`},
	} {
		t.Run(want.what, func(t *testing.T) {
			t.Parallel()

			if !regexp.MustCompile(want.pattern).MatchString(want.in) {
				t.Errorf("got no match for %s, want one", want.pattern)
			}
		})
	}
}

// The workflow borrows the reader by the address infra/ci.env gives it, which
// has to be the one Terraform makes of the deployment's project and service.
func TestTheWorkflowBorrowsTheReaderTerraformMakes(t *testing.T) {
	t.Parallel()

	reader := blockOf(t, read(t, "../terraform/analytics/snapshot.tf"), `resource "google_service_account" "snapshot_reader" {`)
	if account := `account_id\s*=\s*"\$\{var\.service_name\}-live"`; !regexp.MustCompile(account).MatchString(reader) {
		t.Fatalf("got no match for %s in the reader's block, want one", account)
	}
	tfvars := read(t, "../terraform/prod.auto.tfvars")
	service, set := assigned(tfvars, "service_name")
	if !set {
		service, _ = assigned(blockOf(t, read(t, "../terraform/variables.tf"), `variable "service_name" {`), "default")
	}
	project, _ := assigned(tfvars, "project_id")
	want := service + "-live@" + project + ".iam.gserviceaccount.com"

	got := ""
	if line := regexp.MustCompile(`(?m)^GCP_LIVE_SERVICE_ACCOUNT=(\S*)$`).FindStringSubmatch(read(t, "../ci.env")); line != nil {
		got = line[1]
	}
	if got != want {
		t.Errorf("infra/ci.env names the reader %q, want %q", got, want)
	}
}

// assigned is the value a line name = "value" of text gives name, and whether
// a line gives one.
func assigned(text, name string) (string, bool) {
	match := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(name) + `\s*=\s*"([^"]*)"`).FindStringSubmatch(text)
	if match == nil {
		return "", false
	}
	return match[1], true
}
