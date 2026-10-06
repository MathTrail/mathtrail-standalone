// Package analytics holds the SQL of the counts of how the service is used,
// kept for years for grant applications: the schemas of the tables, the script
// that counts the children every night from the raw lines, and the views two
// reports read — exact numbers for the owner, and numbers that show no group of
// fewer than ten children for anyone — and the query whose one row, kept every
// day in a table of its own, is the snapshot of the live numbers of the page
// "Research". Terraform hands the files to BigQuery, but for the queries the
// owner reads the views with: the main numbers for an application.
//
// The Go here is the tests alone. Most run the same files against a BigQuery
// emulator, behind the build tag analytics, since the emulator is a container
// a plain test run does not start. Those that hold the files to one another
// need no emulator, and every test run has them.
package analytics
