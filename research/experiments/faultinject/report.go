package main

import (
	"cmp"
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// codeOrder is the order the service reports its checks in, which every table
// follows.
var codeOrder = []checks.Code{
	checks.CodeBadStructure, checks.CodeDistractorExplanations, checks.CodeDrawingFormat,
	checks.CodeDrawingMismatch, checks.CodeReadability, checks.CodeSolverError,
	checks.CodeSolverDisagrees, checks.CodeSelfCheckBlocking, checks.CodeNearDuplicate,
}

// readPerOperator is how many cases of each out-of-scope operator are read.
const readPerOperator = 10

// writeAll writes every file of the results into a directory. The cases to
// read go first, so that a person can read a sample that has changed even when
// the verdicts kept from an earlier one no longer fit it.
func writeAll(dir string, found *results, ops []operator) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("faultinject: make the results directory: %w", err)
	}
	ids, sample := readingSample(found.Cases)
	if err := writeReading(filepath.Join(dir, "reading.md"), ids, sample); err != nil {
		return fmt.Errorf("faultinject: write reading.md: %w", err)
	}
	read, err := readVerdicts(dir, sample)
	if err != nil {
		return err
	}
	writers := []struct {
		name  string
		write func(path string) error
	}{
		{"sanity.csv", func(path string) error { return writeCSV(path, sanityTable(found.Sanity)) }},
		{"cases.csv", func(path string) error { return writeCSV(path, caseTable(found.Cases)) }},
		{"operators.csv", func(path string) error { return writeCSV(path, operatorTable(found.Cases, ops, read)) }},
		{"classes.csv", func(path string) error { return writeCSV(path, classTable(found.Cases)) }},
		{"checks.csv", func(path string) error { return writeCSV(path, checkTable(found.Cases)) }},
		{"numbers.txt", func(path string) error { return writeNumbers(path, found, ops, read) }},
	}
	for _, w := range writers {
		if err := w.write(filepath.Join(dir, w.name)); err != nil {
			return fmt.Errorf("faultinject: write %s: %w", w.name, err)
		}
	}
	return nil
}

func writeCSV(path string, rows [][]string) error {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	out := csv.NewWriter(file)
	if err := out.WriteAll(rows); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func sanityTable(rows []sanityRow) [][]string {
	table := [][]string{{"host", "level", "topic", "accepted", "codes", "unchecked"}}
	for i := range rows {
		row := &rows[i]
		table = append(table, []string{
			row.Host.ID, string(row.Host.Level), row.Host.Topic, strconv.FormatBool(row.eligible()),
			codesText(row.Verdict.Codes), strconv.Itoa(len(row.Verdict.Unchecked)),
		})
	}
	return table
}

func caseTable(rows []caseRow) [][]string {
	table := [][]string{{
		"operator", "class", "kind", "host", "level", "topic", "donor", "valid", "refused", "expected",
		"codes", "primary", "unchecked", "similarity", "sketch",
	}}
	for i := range rows {
		row := &rows[i]
		table = append(table, []string{
			row.Operator.ID, row.Operator.Class, string(row.Operator.Kind), row.Host.ID, string(row.Host.Level),
			row.Host.Topic, row.Mutant.Donor, strconv.FormatBool(row.Valid), strconv.FormatBool(row.Verdict.Refused()),
			strconv.FormatBool(row.expected()), codesText(row.Verdict.Codes), string(row.Verdict.Primary),
			strconv.Itoa(len(row.Verdict.Unchecked)), measure(row.Similarity), measure(row.Sketch),
		})
	}
	return table
}

// share is how many cases of a group came out one way, of how many: the valid
// cases of an operator or a class, or the cases of one host.
type share struct {
	hits, total int
}

func (s share) percent() string {
	if s.total == 0 {
		return ""
	}
	return strconv.FormatFloat(100*float64(s.hits)/float64(s.total), 'f', 1, 64)
}

// exact is the share with its Clopper–Pearson interval, as percentages.
func (s share) exact() []string {
	if s.total == 0 {
		return []string{"", "", ""}
	}
	low, high := clopperPearson(s.hits, s.total)
	return []string{s.percent(), percent(low), percent(high)}
}

func percent(p float64) string { return strconv.FormatFloat(100*p, 'f', 1, 64) }

func measure(value float64) string {
	if value < 0 {
		return ""
	}
	return strconv.FormatFloat(value, 'f', 3, 64)
}

func codesText(codes []checks.Code) string {
	names := make([]string, 0, len(codes))
	for _, code := range codes {
		names = append(names, string(code))
	}
	return strings.Join(names, ";")
}

// counted is how many valid cases of a group a condition holds for.
func counted(rows []caseRow, belongs, holds func(row *caseRow) bool) share {
	var s share
	for i := range rows {
		row := &rows[i]
		if !row.Valid || !belongs(row) {
			continue
		}
		s.total++
		if holds(row) {
			s.hits++
		}
	}
	return s
}

// operatorTable is every operator with the shares of its valid cases refused,
// by any check and by the one expected, and how many of its cases a person
// read and found not to be the defect it is named after.
func operatorTable(rows []caseRow, ops []operator, read map[string]share) [][]string {
	table := [][]string{{
		"operator", "class", "kind", "expected_codes", "applied", "valid",
		"refused_percent", "refused_low", "refused_high", "expected_percent", "expected_low", "expected_high",
		"read", "equivalent",
	}}
	for i := range ops {
		op := &ops[i]
		belongs := func(row *caseRow) bool { return row.Operator.ID == op.ID }
		applied := 0
		for j := range rows {
			if belongs(&rows[j]) {
				applied++
			}
		}
		refused := counted(rows, belongs, func(row *caseRow) bool { return row.Verdict.Refused() })
		expected := counted(rows, belongs, (*caseRow).expected)
		line := []string{op.ID, op.Class, string(op.Kind), codesText(op.Expected), strconv.Itoa(applied), strconv.Itoa(refused.total)}
		table = append(table, slices.Concat(line, refused.exact(), expected.exact(), readCounts(read, op.ID)))
	}
	return table
}

// classesOf are the classes of the cases, in order.
func classesOf(rows []caseRow) []string {
	var classes []string
	for i := range rows {
		if !slices.Contains(classes, rows[i].Operator.Class) {
			classes = append(classes, rows[i].Operator.Class)
		}
	}
	slices.Sort(classes)
	return classes
}

// classTable is every class with the shares of its valid cases refused, by
// any check and by the one expected, each with the interval of a bootstrap
// over its hosts and, beside it, the exact interval of the pooled cases: the
// bootstrap has no width when every host has the same share, as at 0 and 100 %.
func classTable(rows []caseRow) [][]string {
	table := [][]string{{
		"class", "valid", "refused_percent", "refused_low", "refused_high", "refused_exact_low", "refused_exact_high",
		"expected_percent", "expected_low", "expected_high", "expected_exact_low", "expected_exact_high", "unchecked_percent",
	}}
	for _, class := range classesOf(rows) {
		belongs := func(row *caseRow) bool { return row.Operator.Class == class }
		refused := counted(rows, belongs, func(row *caseRow) bool { return row.Verdict.Refused() })
		expected := counted(rows, belongs, (*caseRow).expected)
		unchecked := counted(rows, belongs, func(row *caseRow) bool { return len(row.Verdict.Unchecked) > 0 })
		line := []string{class, strconv.Itoa(refused.total), refused.percent()}
		line = append(line, pooledInterval(rows, class, func(row *caseRow) bool { return row.Verdict.Refused() })...)
		line = append(line, refused.exact()[1:]...)
		line = append(line, expected.percent())
		line = append(line, pooledInterval(rows, class, (*caseRow).expected)...)
		line = append(line, expected.exact()[1:]...)
		table = append(table, append(line, unchecked.percent()))
	}
	return table
}

// pooledInterval is a class's interval by the bootstrap over its hosts.
func pooledInterval(rows []caseRow, class string, holds func(row *caseRow) bool) []string {
	perHost := map[string]*share{}
	var hosts []string
	for i := range rows {
		row := &rows[i]
		if !row.Valid || row.Operator.Class != class {
			continue
		}
		t, seen := perHost[row.Host.ID]
		if !seen {
			t = &share{}
			perHost[row.Host.ID] = t
			hosts = append(hosts, row.Host.ID)
		}
		t.total++
		if holds(row) {
			t.hits++
		}
	}
	tallies := make([]share, 0, len(hosts))
	for _, id := range hosts {
		tallies = append(tallies, *perHost[id])
	}
	low, high := bootstrapByHost(tallies, seeded(class, "bootstrap"))
	return []string{percent(low), percent(high)}
}

// checkTable is, for every class and check, the share of valid cases the
// check refused, the share it refused first, and the share it refused alone.
func checkTable(rows []caseRow) [][]string {
	table := [][]string{{"class", "check", "fired_percent", "first_percent", "only_percent"}}
	for _, class := range classesOf(rows) {
		belongs := func(row *caseRow) bool { return row.Operator.Class == class }
		for _, code := range codeOrder {
			fired := counted(rows, belongs, func(row *caseRow) bool { return row.Verdict.Has(code) })
			first := counted(rows, belongs, func(row *caseRow) bool { return row.Verdict.Primary == code })
			only := counted(rows, belongs, func(row *caseRow) bool {
				return len(row.Verdict.Codes) == 1 && row.Verdict.Codes[0] == code
			})
			table = append(table, []string{class, string(code), fired.percent(), first.percent(), only.percent()})
		}
	}
	return table
}

// writeNumbers writes the numbers the paper cites, as key=value lines.
func writeNumbers(path string, found *results, ops []operator, read map[string]share) error {
	lines := []string{"# Numbers of the experiment with injected defects, computed by faultinject."}
	lines = append(lines, hostNumbers(found)...)
	lines = append(lines, "operators="+strconv.Itoa(len(ops)))
	lines = append(lines, mechanismNumbers(found.Cases, ops)...)
	lines = append(lines, classNumbers(found.Cases)...)
	for i := range ops {
		lines = append(lines, operatorNumbers(found.Cases, &ops[i], read)...)
	}
	return writeText(path, strings.Join(lines, "\n")+"\n")
}

// writeText writes a text file of the results, made as the CSV files are.
func writeText(path, text string) error {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	if _, err := file.WriteString(text); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// hostNumbers are how many reference tasks there are at each level and in
// all, how many of them the checks accept as they are, and how many cases and
// classes the run made.
func hostNumbers(found *results) []string {
	lines := []string{"hosts=" + strconv.Itoa(len(found.Sanity))}
	accepted := 0
	for _, level := range rating.GradeLevels() {
		all, eligible := 0, 0
		for i := range found.Sanity {
			if found.Sanity[i].Host.Level != level {
				continue
			}
			all++
			if found.Sanity[i].eligible() {
				eligible++
			}
		}
		suffix := strings.ReplaceAll(string(level), "-", "")
		lines = append(lines, "hosts_"+suffix+"="+strconv.Itoa(all), "eligible_"+suffix+"="+strconv.Itoa(eligible))
		accepted += eligible
	}
	valid := counted(found.Cases, func(*caseRow) bool { return true }, func(*caseRow) bool { return true })
	return append(lines,
		"eligible="+strconv.Itoa(accepted),
		"cases="+strconv.Itoa(valid.total),
		"classes="+strconv.Itoa(len(classesOf(found.Cases))),
	)
}

// drawingClasses are the classes only a task with a drawing takes.
var drawingClasses = []string{"D16", "D17", "D18"}

// mechanismNumbers sum up the operators that break a rule a check states: how
// many there are, how many made no case, their cases, the share refused by the
// check expected and how many it missed, the lowest start of an operator's exact interval, apart for
// the operators only the tasks with a drawing take, and how many operators one
// check alone refused in every case it refused. An operator with no case has no
// interval and is left out of the lowest; the count of them says so.
func mechanismNumbers(rows []caseRow, ops []operator) []string {
	count, empty, alone, lowest, lowestDrawing := 0, 0, 0, 1.0, 1.0
	fewestDrawing, mostDrawing := math.MaxInt, 0
	for i := range ops {
		op := &ops[i]
		if op.Kind != mechanism {
			continue
		}
		count++
		expected := counted(rows, func(row *caseRow) bool { return row.Operator.ID == op.ID }, (*caseRow).expected)
		if expected.total == 0 {
			empty++
			continue
		}
		if refusedByOneCheck(rows, op) {
			alone++
		}
		low, _ := clopperPearson(expected.hits, expected.total)
		if !slices.Contains(drawingClasses, op.Class) {
			lowest = min(lowest, low)
			continue
		}
		lowestDrawing = min(lowestDrawing, low)
		fewestDrawing, mostDrawing = min(fewestDrawing, expected.total), max(mostDrawing, expected.total)
	}
	if mostDrawing == 0 {
		fewestDrawing = 0
	}
	expected := counted(rows, func(row *caseRow) bool { return row.Operator.Kind == mechanism }, (*caseRow).expected)
	return []string{
		"mechanism_operators=" + strconv.Itoa(count),
		"mechanism_operators_without_cases=" + strconv.Itoa(empty),
		"mechanism_operators_one_check=" + strconv.Itoa(alone),
		"mechanism_operators_more_checks=" + strconv.Itoa(count-empty-alone),
		"mechanism_cases=" + strconv.Itoa(expected.total),
		"mechanism_expected_percent=" + expected.percent(),
		"mechanism_missed=" + strconv.Itoa(expected.total-expected.hits),
		"mechanism_lowest_low=" + percent(lowest),
		"drawing_lowest_low=" + percent(lowestDrawing),
		"drawing_hosts_fewest=" + strconv.Itoa(fewestDrawing),
		"drawing_hosts_most=" + strconv.Itoa(mostDrawing),
	}
}

// classNumbers are, for every class, its valid cases, the share refused by any
// check with its exact interval, and the share refused by the check expected.
func classNumbers(rows []caseRow) []string {
	var lines []string
	for _, class := range classesOf(rows) {
		key := strings.ToLower(class)
		belongs := func(row *caseRow) bool { return row.Operator.Class == class }
		refused := counted(rows, belongs, func(row *caseRow) bool { return row.Verdict.Refused() })
		expected := counted(rows, belongs, (*caseRow).expected)
		values := refused.exact()
		lines = append(lines,
			key+"_cases="+strconv.Itoa(refused.total),
			key+"_refused_percent="+values[0],
			key+"_refused_low="+values[1],
			key+"_refused_high="+values[2],
			key+"_expected_percent="+expected.percent(),
		)
	}
	return lines
}

// refusedByOneCheck says whether one and the same check, alone, refused every
// valid case of the operator, so that switching that check off would let all
// of its cases through.
func refusedByOneCheck(rows []caseRow, op *operator) bool {
	var only checks.Code
	for i := range rows {
		if rows[i].Operator.ID != op.ID || !rows[i].Valid {
			continue
		}
		codes := rows[i].Verdict.Codes
		if len(codes) != 1 || (only != "" && codes[0] != only) {
			return false
		}
		only = codes[0]
	}
	return only != ""
}

// numberKey is an id written as a key of the numbers file.
func numberKey(id string) string {
	return strings.ToLower(strings.NewReplacer("-", "_", ".", "_").Replace(id))
}

// operatorNumbers are what the paper cites of one operator: its valid cases
// and the share refused, by the check expected or, out of scope, by any; for
// a discovery operator the exact interval, where its misses fell, and how many
// cases any check refused, with the share's exact interval; for an
// out-of-scope operator what its reading found.
func operatorNumbers(rows []caseRow, op *operator, read map[string]share) []string {
	key := numberKey(op.ID)
	belongs := func(row *caseRow) bool { return row.Operator.ID == op.ID }
	if op.Kind == outOfScope {
		refused := counted(rows, belongs, func(row *caseRow) bool { return row.Verdict.Refused() })
		values := refused.exact()
		lines := []string{
			key + "_cases=" + strconv.Itoa(refused.total), key + "_refused_percent=" + values[0],
			key + "_refused_low=" + values[1], key + "_refused_high=" + values[2],
		}
		if verdicts, found := read[op.ID]; found {
			lines = append(lines, key+"_read="+strconv.Itoa(verdicts.total), key+"_equivalent="+strconv.Itoa(verdicts.hits))
		}
		return lines
	}
	expected := counted(rows, belongs, (*caseRow).expected)
	lines := []string{key + "_cases=" + strconv.Itoa(expected.total), key + "_expected_percent=" + expected.percent()}
	if op.Kind != discovery {
		return lines
	}
	values := expected.exact()
	refused := counted(rows, belongs, func(row *caseRow) bool { return row.Verdict.Refused() })
	byAny := refused.exact()
	lines = append(lines,
		key+"_expected_low="+values[1],
		key+"_expected_high="+values[2],
		key+"_missed="+strconv.Itoa(expected.total-expected.hits),
		key+"_refused="+strconv.Itoa(refused.hits),
		key+"_refused_percent="+byAny[0],
		key+"_refused_low="+byAny[1],
		key+"_refused_high="+byAny[2],
	)
	return append(lines, missesByTopic(rows, op, key)...)
}

// missesByTopic are, for every topic where an operator's defect got through,
// the operator's cases in that topic and how many got through.
func missesByTopic(rows []caseRow, op *operator, key string) []string {
	var topics []string
	for i := range rows {
		row := &rows[i]
		if row.Operator.ID == op.ID && row.Valid && !row.expected() && !slices.Contains(topics, row.Host.Topic) {
			topics = append(topics, row.Host.Topic)
		}
	}
	slices.Sort(topics)
	var lines []string
	for _, topic := range topics {
		inTopic := counted(rows, func(row *caseRow) bool { return row.Operator.ID == op.ID && row.Host.Topic == topic }, (*caseRow).expected)
		prefix := key + "_" + numberKey(topic)
		lines = append(lines, prefix+"_cases="+strconv.Itoa(inTopic.total), prefix+"_missed="+strconv.Itoa(inTopic.total-inTopic.hits))
	}
	return lines
}

// readVerdicts are the verdicts a person recorded in reading.csv on the cases
// they read, per operator: how many were read, and how many of those were not
// the defect their operator is named after. Every verdict must be on a case of
// the sample this run draws, and once: a verdict on another case was given on
// a sample that has since changed. A directory with no such file has none yet.
func readVerdicts(dir string, sample map[string][]*caseRow) (map[string]share, error) {
	file, err := os.Open(filepath.Clean(filepath.Join(dir, "reading.csv")))
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]share{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("faultinject: open the verdicts: %w", err)
	}
	defer func() { _ = file.Close() }()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("faultinject: read the verdicts: %w", err)
	}
	verdicts := map[string]share{}
	seen := map[string]bool{}
	for _, row := range rows[min(1, len(rows)):] {
		if err := sampled(row, sample, seen); err != nil {
			return nil, fmt.Errorf("faultinject: read the verdicts: %w", err)
		}
		verdict := verdicts[row[0]]
		verdict.total++
		if row[2] == "equivalent" {
			verdict.hits++
		}
		verdicts[row[0]] = verdict
	}
	return verdicts, nil
}

// sampled says what is wrong with one verdict: a row too short, a verdict
// that is neither real nor equivalent, a case outside the sample, or a case
// given twice.
func sampled(row []string, sample map[string][]*caseRow, seen map[string]bool) error {
	if len(row) < 3 {
		return fmt.Errorf("a row of %d fields, want an operator, a host and a verdict", len(row))
	}
	if row[2] != "real" && row[2] != "equivalent" {
		return fmt.Errorf("the verdict %q is neither real nor equivalent", row[2])
	}
	key := row[0] + " on " + row[1]
	if seen[key] {
		return fmt.Errorf("%s has two verdicts", key)
	}
	seen[key] = true
	if !slices.ContainsFunc(sample[row[0]], func(c *caseRow) bool { return c.Host.ID == row[1] }) {
		return fmt.Errorf("%s is not a case of this run's sample", key)
	}
	return nil
}

// readCounts are how many cases of an operator were read, and how many of
// them were found not to be its defect; empty when none was read.
func readCounts(read map[string]share, op string) []string {
	verdicts, found := read[op]
	if !found {
		return []string{"", ""}
	}
	return []string{strconv.Itoa(verdicts.total), strconv.Itoa(verdicts.hits)}
}

// readingSample is, for every out-of-scope operator in order, the cases drawn
// for a person to read, by a seed of the operator's own.
func readingSample(rows []caseRow) (ids []string, sample map[string][]*caseRow) {
	ids = outOfScopeIDs(rows)
	sample = make(map[string][]*caseRow, len(ids))
	for _, op := range ids {
		var cases []*caseRow
		for i := range rows {
			if rows[i].Operator.ID == op && rows[i].Valid {
				cases = append(cases, &rows[i])
			}
		}
		rng := seeded(op, "reading")
		rng.Shuffle(len(cases), func(i, j int) { cases[i], cases[j] = cases[j], cases[i] })
		sample[op] = cases[:min(readPerOperator, len(cases))]
	}
	return ids, sample
}

// writeReading lists, for every out-of-scope operator, the cases drawn for a
// person to read: is each the defect it is named after?
func writeReading(path string, ids []string, sample map[string][]*caseRow) error {
	var out strings.Builder
	out.WriteString("# Cases of the out-of-scope operators to read\n\n")
	out.WriteString("For each case: is it the defect its operator is named after? Record the verdict in reading.csv.\n")
	for _, op := range ids {
		fmt.Fprintf(&out, "\n## %s\n", op)
		for _, row := range sample[op] {
			writeReadingCase(&out, row)
		}
	}
	return writeText(path, out.String())
}

func writeReadingCase(out *strings.Builder, row *caseRow) {
	task := &row.Mutant.Sub.Task
	fmt.Fprintf(out, "\n### %s on %s\n\n", row.Operator.ID, row.Host.ID)
	fmt.Fprintf(out, "- Before: %s\n", row.Host.Base.Task.Question)
	if row.Mutant.Source != "" && row.Mutant.Source != task.Question {
		fmt.Fprintf(out, "- Made from: %s\n", row.Mutant.Source)
	}
	fmt.Fprintf(out, "- After: %s\n", task.Question)
	fmt.Fprintf(out, "- Key: %s, %q\n", task.CorrectAnswer, task.Options[task.CorrectAnswer])
	if task.Hint != row.Host.Base.Task.Hint {
		fmt.Fprintf(out, "- Hint after: %s\n", task.Hint)
	}
	for _, letter := range []string{"A", "B", "C", "D", "E"} {
		before, after := row.Host.Base.Task.Distractors[letter].Trap, task.Distractors[letter].Trap
		if before != after {
			fmt.Fprintf(out, "- Trap of %s: %s → %s\n", letter, before, after)
		}
	}
	fmt.Fprintf(out, "- Refused by: %s\n", cmp.Or(codesText(row.Verdict.Codes), "nothing"))
}

// outOfScopeIDs are the out-of-scope operators that made cases, in order.
func outOfScopeIDs(rows []caseRow) []string {
	var ids []string
	for i := range rows {
		if rows[i].Operator.Kind == outOfScope && !slices.Contains(ids, rows[i].Operator.ID) {
			ids = append(ids, rows[i].Operator.ID)
		}
	}
	return ids
}
