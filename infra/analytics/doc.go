// Package analytics holds the SQL of the counts of how the service is used,
// kept for years for grant applications: the schemas of the tables, the script
// that counts the children every night from the raw lines, and the views two
// reports read — exact numbers for the owner, and numbers that show no group of
// fewer than ten children for anyone. Terraform hands the files to BigQuery.
//
// The Go here is the tests alone. They run the same files against a BigQuery
// emulator, behind the build tag analytics, since the emulator is a container
// a plain test run does not start.
package analytics
