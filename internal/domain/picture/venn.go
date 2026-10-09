package picture

// venn is two groups that overlap, inside a box of everyone the question
// counts: each group's label and size, the count in both, in neither, and of
// everyone, every count a label.
type venn struct {
	labels []string
}

func readVenn(o *object) Picture {
	var v venn
	for _, set := range o.objects("sets", vennSets, vennSets, "sets", true) {
		if set == nil {
			continue
		}
		label := set.label("label", true)
		count := set.label("count", false)
		set.close("a set")
		v.labels = append(v.labels, label, count)
	}
	for _, name := range []string{"both", "neither", "total"} {
		label := o.label(name, false)
		v.labels = append(v.labels, label)
	}
	o.close("two groups")
	return v
}

func (venn) Kind() Kind { return Venn }

func (v venn) Labels() []string { return present(v.labels...) }

// Shown are every label and count the groups carry.
func (v venn) Shown() []Shown { return texts(v.labels...) }
