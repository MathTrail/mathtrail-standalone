package analytics_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// placeholder is a name the SQL is written with, as templatefile writes one.
var placeholder = regexp.MustCompile(`\$\{(\w+)\}`)

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

// The daily report's view is made over the export only once the deployment
// says the export has made its table, in the dataset of the owner's exact
// numbers, from the file its tests run; and the export's dataset is kept where
// the counts are, so one query reads both.
func TestTheDailyReportWaitsForTheExportsTable(t *testing.T) {
	t.Parallel()

	report := read(t, "../terraform/analytics/daily_report.tf")
	view := blockOf(t, report, `resource "google_bigquery_table" "daily_report" {`)
	dataset := blockOf(t, report, `resource "google_bigquery_dataset" "billing" {`)
	for _, want := range []struct{ what, in, pattern string }{
		{"the view waits for the export's table", view, `count\s*=\s*var\.billing_export \? 1 : 0`},
		{"the view is among the owner's exact numbers", view, `dataset_id\s*=\s*google_bigquery_dataset\.private\.dataset_id`},
		{"the view is daily_report", view, `table_id\s*=\s*"daily_report"`},
		{"the view runs the report's SQL", view, `query\s*=\s*templatefile\("\$\{local\.analytics\}/billing/daily_report\.sql"`},
		{"the export's dataset is billing", dataset, `dataset_id\s*=\s*"billing"`},
		{"the export's dataset is where the counts are", dataset, `location\s*=\s*var\.region`},
		{"the deployment says when the export has its table", read(t, "../terraform/analytics.tf"), `billing_export\s*=\s*var\.billing_export`},
	} {
		t.Run(want.what, func(t *testing.T) {
			t.Parallel()

			if !regexp.MustCompile(want.pattern).MatchString(want.in) {
				t.Errorf("got no match for %s, want one", want.pattern)
			}
		})
	}
}

// Who may write the export's dataset is the console's to say once the export
// is turned on: no block of the configuration names a right on it or puts
// anything in it, so no apply takes the export's own right away.
func TestTheExportsDatasetIsTheConsolesToGrant(t *testing.T) {
	t.Parallel()

	dataset := blockOf(t, read(t, "../terraform/analytics/daily_report.tf"), `resource "google_bigquery_dataset" "billing" {`)
	if access := regexp.MustCompile(`(?m)^\s*access\s*[{=].*$`).FindString(dataset); access != "" {
		t.Errorf("got %q in the dataset billing, want no access named: the console keeps the export's", strings.TrimSpace(access))
	}
	files, err := filepath.Glob("../terraform/analytics/*.tf")
	if err != nil {
		t.Fatalf("list the module: %v", err)
	}
	naming := regexp.MustCompile(`dataset_id\s*=\s*(google_bigquery_dataset\.billing\.dataset_id|"billing")`)
	got := 0
	for _, file := range files {
		got += len(naming.FindAllString(read(t, file), -1))
	}
	if got != 1 {
		t.Errorf("got %d blocks naming the dataset billing, want 1, the dataset itself", got)
	}
}

// The export's table is named after the billing account, which this
// repository does not hold: the report reads the standard export through a
// wildcard and is given no name but the datasets' and the project's.
func TestTheDailyReportNamesNoBillingAccount(t *testing.T) {
	t.Parallel()

	sql := read(t, "billing/daily_report.sql")
	given := map[string]bool{"impact": true, "billing": true, "project_id": true}
	for _, found := range placeholder.FindAllStringSubmatch(sql, -1) {
		if !given[found[1]] {
			t.Errorf("got ${%s} in the report, want only ${impact}, ${billing} and ${project_id}", found[1])
		}
	}
	if wildcard := "`${billing}.gcp_billing_export_v1_*`"; !strings.Contains(sql, wildcard) {
		t.Errorf("got no %s in the report, want the standard export read through it", wildcard)
	}
}
