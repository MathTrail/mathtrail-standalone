package picture

// balance is a pan balance and what stands on each of its pans.
type balance struct {
	labels []string
}

func readBalance(o *object) Picture {
	var b balance
	b.labels = append(b.labels, o.labels("left", 0, maxOnAPan, true)...)
	b.labels = append(b.labels, o.labels("right", 0, maxOnAPan, true)...)
	o.close("a balance")
	return b
}

func (balance) Kind() Kind { return Balance }

func (b balance) Labels() []string { return present(b.labels...) }

// Shown are the labels of what stands on the pans.
func (b balance) Shown() []Shown { return texts(b.labels...) }
