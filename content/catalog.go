package content

import (
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// Topic is one entry of the topic catalog: what a task can be about, and the
// grade levels it is taught at. The catalog is closed and the model never
// invents a topic, because two children's histories are comparable only while
// they name the same things.
//
// Everything the catalogs hold is keyed by the level a task is written for
// rather than by a school year: what changes between year 1 and year 2 is not
// which topics exist, and a child of any grade may be set the tasks of any
// level their answers lead them to.
type Topic struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	GradeLevels []rating.GradeLevel `json:"grade_levels"`
	// Slug is the last part of the address of the topic's page on the public
	// site. Once the page is published it never changes: somebody else holds a
	// link to it by then.
	Slug string `json:"slug"`
	// BuildsOn are the topics whose solid mastery is a good base to start this
	// one, in catalog order. A topic that builds on none is a foundation. The
	// links draw the map of topics and suggest which new topic to start; the
	// choice of a task does not read them.
	BuildsOn []string `json:"builds_on"`
	// SitePage reports whether the site has published the topic's page, which
	// is what a link to the page waits for.
	SitePage bool `json:"site_page"`
}

// HasLevel reports whether the topic is offered at this grade level.
func (t *Topic) HasLevel(level rating.GradeLevel) bool {
	return slices.Contains(t.GradeLevels, level)
}

// clone copies a topic together with the levels and the bases inside it, so
// that what a caller is handed shares no memory with the catalog it came from.
// Traps and skills need nothing of the kind: they are strings, and a value copy
// of one is already a copy of all of it.
func (t *Topic) clone() Topic {
	copied := *t
	copied.GradeLevels = slices.Clone(t.GradeLevels)
	copied.BuildsOn = slices.Clone(t.BuildsOn)
	return copied
}

// Trap is one entry of the trap catalog: the mistake behind a wrong option.
// Every wrong option in every task names one, which is what turns a wrong
// answer into a diagnosis instead of a tick in the wrong column.
type Trap struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Skill is one entry of the skill catalog: something a child may not have met
// at school yet. A task uses none of what the profile excludes, neither in the
// wording nor in a trap.
type Skill struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

const catalogsDir = "catalogs"

var (
	// A topic id names an area and a topic inside it, as in logic.ordering. Each
	// of the two starts with letters, and a digit may follow an underscore.
	topicIDPattern = regexp.MustCompile(`^[a-z]+(_[a-z0-9]+)*\.[a-z]+(_[a-z0-9]+)*$`)
	// A trap or skill id is one snake_case name, as in off_by_one.
	entryIDPattern = regexp.MustCompile(`^[a-z]+(_[a-z0-9]+)*$`)
	// A slug is lowercase words and numbers joined by dashes, as in
	// knights-and-liars: nothing in it can lead an address anywhere but the
	// topic's page. It reads like the name of a solver template and is a rule of
	// its own all the same, so that either can change without the other.
	slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
)

// loadTopics reads the topic catalog and refuses one that could not be used:
// an id nothing else could refer to, an entry twice over, a level outside the
// three that exist, a page with no address or with another topic's, and links
// between topics that lead nowhere or round in a circle.
func loadTopics(src fs.FS) ([]Topic, error) {
	file := catalogsDir + "/topics.json"
	var topics []Topic
	if err := decode(src, file, &topics); err != nil {
		return nil, err
	}

	p := catalogProblems(file, len(topics))
	seen := make(map[string]bool, len(topics))
	for i := range topics {
		checkTopic(p, entryName("topic", i, topics[i].ID), &topics[i], seen)
	}
	checkSlugs(p, topics)
	checkBases(p, topics)
	return topics, p.err()
}

// checkTopic records everything wrong with one entry of the topic catalog.
func checkTopic(p *problems, where string, topic *Topic, seen map[string]bool) {
	if !topicIDPattern.MatchString(topic.ID) {
		p.addf("%s: an id is an area and a topic in snake case, separated by a dot", where)
	} else if seen[topic.ID] {
		p.addf("%s: the id is used twice", where)
	}
	seen[topic.ID] = true

	if topic.Name == "" {
		p.addf("%s: the name is empty", where)
	}
	if topic.Description == "" {
		p.addf("%s: the description is empty", where)
	}
	if len(topic.GradeLevels) == 0 {
		p.addf("%s: no grade levels, so the topic can never be chosen", where)
	}

	levels := make(map[rating.GradeLevel]bool, len(topic.GradeLevels))
	for _, level := range topic.GradeLevels {
		switch {
		case !level.Known():
			p.addf("%s: grade level %q is not one of %v", where, level, rating.GradeLevels())
		case levels[level]:
			p.addf("%s: grade level %q is listed twice", where, level)
		}
		levels[level] = true
	}

	if !slugPattern.MatchString(topic.Slug) {
		p.addf("%s: a slug is lowercase words and numbers joined by dashes, as in knights-and-liars", where)
	}
	// A topic that builds on nothing says so, so that a forgotten key cannot
	// quietly set it at the foot of the map.
	if topic.BuildsOn == nil {
		p.addf("%s: builds_on is missing; a topic that builds on no other says []", where)
	}
}

// checkSlugs refuses two topics whose pages would share one address.
func checkSlugs(p *problems, topics []Topic) {
	owner := make(map[string]string, len(topics))
	for i, topic := range topics {
		if topic.Slug == "" {
			continue
		}
		if first, taken := owner[topic.Slug]; taken {
			p.addf("%s: the slug %q is already %q's", entryName("topic", i, topic.ID), topic.Slug, first)
			continue
		}
		owner[topic.Slug] = topic.ID
	}
}

// checkBases refuses links between topics that lead nowhere or round in a
// circle: a base the catalog does not have, a topic that builds on itself or
// names one base twice, and topics that build on one another in a circle,
// which would leave the map of topics with no first step. A topic may build on
// one listed after it, so the links are read once every topic is.
func checkBases(p *problems, topics []Topic) {
	byID := index(topics, func(t Topic) string { return t.ID })
	place := make(map[string]int, len(topics))
	for i := range topics {
		place[topics[i].ID] = i
	}
	for i, topic := range topics {
		where := entryName("topic", i, topic.ID)
		named := make(map[string]bool, len(topic.BuildsOn))
		last := ""
		for _, base := range topic.BuildsOn {
			_, known := byID[base]
			switch {
			case base == topic.ID:
				p.addf("%s: a topic cannot build on itself", where)
			case named[base]:
				p.addf("%s: it builds on %q twice", where, base)
			case !known:
				p.addf("%s: it builds on %q, which is not in the catalog", where, base)
			case last != "" && place[base] < place[last]:
				p.addf("%s: it names %q after %q, and a topic's bases follow the catalog's order", where, base, last)
			}
			named[base] = true
			if known && base != topic.ID {
				last = base
			}
		}
	}
	for _, circle := range circles(topics, byID) {
		p.addf("topics build on one another in a circle: %s", strings.Join(circle, " → "))
	}
}

// circles finds the circles of topics that build on one another, each as the
// chain of ids that closes it, read as "builds on": one for every link that
// leads back into the chain being followed, so at least one through every knot
// of topics that build on one another. A topic that builds on itself, names a
// base twice or builds on a topic the catalog does not have is reported on its
// own, and the link is skipped here.
func circles(topics []Topic, byID map[string]Topic) [][]string {
	const (
		unseen = iota
		onPath
		done
	)
	state := make(map[string]int, len(topics))
	var path []string
	var found [][]string
	var visit func(id string)
	visit = func(id string) {
		state[id] = onPath
		path = append(path, id)
		named := make(map[string]bool, len(byID[id].BuildsOn))
		for _, base := range byID[id].BuildsOn {
			if _, known := byID[base]; !known || base == id || named[base] {
				continue
			}
			named[base] = true
			switch state[base] {
			case onPath:
				found = append(found, append(slices.Clone(path[slices.Index(path, base):]), base))
			case unseen:
				visit(base)
			}
		}
		path = path[:len(path)-1]
		state[id] = done
	}
	for _, topic := range topics {
		if state[topic.ID] == unseen {
			visit(topic.ID)
		}
	}
	return found
}

// loadTraps reads the trap catalog.
func loadTraps(src fs.FS) ([]Trap, error) {
	file := catalogsDir + "/traps.json"
	var traps []Trap
	if err := decode(src, file, &traps); err != nil {
		return nil, err
	}

	p := catalogProblems(file, len(traps))
	seen := make(map[string]bool, len(traps))
	for i, trap := range traps {
		checkEntry(p, entryName("trap", i, trap.ID), trap.ID, trap.Description, seen)
	}
	return traps, p.err()
}

// loadSkills reads the skill catalog.
func loadSkills(src fs.FS) ([]Skill, error) {
	file := catalogsDir + "/skills.json"
	var skills []Skill
	if err := decode(src, file, &skills); err != nil {
		return nil, err
	}

	p := catalogProblems(file, len(skills))
	seen := make(map[string]bool, len(skills))
	for i, skill := range skills {
		checkEntry(p, entryName("skill", i, skill.ID), skill.ID, skill.Description, seen)
	}
	return skills, p.err()
}

// catalogProblems starts the fault list of one catalog with the fault every
// catalog can have before a single entry is read: a file that parses and holds
// nothing at all.
func catalogProblems(file string, entries int) *problems {
	p := &problems{file: file}
	if entries == 0 {
		p.addf("the catalog is empty")
	}
	return p
}

// checkEntry checks what the trap and skill catalogs have in common: a
// snake_case id that appears once, and a description a person can read. The
// description is what the model is shown, so an empty one silently weakens
// every task built on that entry.
func checkEntry(p *problems, where, id, description string, seen map[string]bool) {
	if !entryIDPattern.MatchString(id) {
		p.addf("%s: an id is one name in snake case", where)
	} else if seen[id] {
		p.addf("%s: the id is used twice", where)
	}
	seen[id] = true

	if description == "" {
		p.addf("%s: the description is empty", where)
	}
}

// entryName names an entry in a fault the way its reader will look for it: by
// id where there is one, and by position where the id is what is broken.
func entryName(kind string, i int, id string) string {
	if id == "" {
		return fmt.Sprintf("%s %d", kind, i+1)
	}
	return fmt.Sprintf("%s %q", kind, id)
}
