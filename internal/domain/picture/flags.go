package picture

import "fmt"

// flags is flags in groups — the flags an enumeration makes, gathered by what
// they share — each flag two or three stripes from top to bottom, a stripe a
// colour of the picture or ?, and each group named under it by a colour or a
// label.
type flags struct {
	colors colors
	groups []flagGroup
}

// flagGroup is one group of flags: the label that names it, if a label does,
// and how many flags it holds.
type flagGroup struct {
	label string
	count int
}

func readFlags(o *object) Picture {
	f := flags{colors: readColors(o)}
	used := map[Paint]bool{}
	total := 0
	for _, group := range o.objects("groups", minFlagGroups, maxFlagGroups, "groups of flags", true) {
		read := readFlagGroup(group, f.colors, used)
		total += read.count
		f.groups = append(f.groups, read)
	}
	if total > maxFlags {
		o.fault(o.at("groups"), "must hold at most %d flags in all", maxFlags)
	}
	f.colors.unused(o, used)
	o.close("flags")
	return f
}

// readFlagGroup reads one group of flags and what names it under them: a
// label or a colour, never both.
func readFlagGroup(group *object, painted colors, used map[Paint]bool) flagGroup {
	if group == nil {
		return flagGroup{}
	}
	read := flagGroup{label: group.label("label", false)}
	if color, held := group.member("color"); held {
		if read.label != "" {
			group.fault(group.path, "is named by a colour or a label, never both")
		}
		painted.paint(group.reading, group.at("color"), color, false, used)
	}
	stripesOf := group.list("flags", minFlags, maxFlagsInGroup, "flags", true)
	read.count = len(stripesOf)
	for i, flag := range stripesOf {
		path := fmt.Sprintf("%s.%d", group.at("flags"), i)
		stripes, isList := flag.([]any)
		if !isList || len(stripes) < minStripes || len(stripes) > maxStripes {
			group.fault(path, "must be a flag: a list of %d to %d stripes, from top to bottom", minStripes, maxStripes)
			continue
		}
		for j, stripe := range stripes {
			painted.paint(group.reading, fmt.Sprintf("%s.%d", path, j), stripe, true, used)
		}
	}
	group.close("a group of flags")
	return read
}

func (flags) Kind() Kind { return Flags }

func (f flags) Labels() []string {
	labels := make([]string, 0, len(f.groups))
	for _, group := range f.groups {
		labels = append(labels, group.label)
	}
	return present(labels...)
}

// Shown are the labels of the groups, how many flags each group holds and
// how many there are in all: a child counts the flags off the picture.
func (f flags) Shown() []Shown {
	var shown []Shown
	total := 0
	for _, group := range f.groups {
		shown = append(shown, texts(group.label)...)
		shown = append(shown, number(group.count))
		total += group.count
	}
	return append(shown, number(total))
}

// Words are the words of the colours the flags are painted with.
func (f flags) Words() []string { return f.colors.words() }
