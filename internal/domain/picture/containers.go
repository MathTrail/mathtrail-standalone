package picture

// containers is jugs, buckets and the like to pour between: how much each can
// hold and how much it holds now, both written on it by the card.
type containers struct {
	noWords
	items []container
}

// container is one container: its capacity, what it holds, and its label, if
// it has one; a number that could not be read is 0 and not shown.
type container struct {
	capacity, amount int
	label            string
	amountHeld       bool
}

func readContainers(o *object) Picture {
	var c containers
	for _, item := range o.objects("items", minContainers, maxContainers, "containers", true) {
		if item != nil {
			c.items = append(c.items, readContainer(item))
		}
	}
	o.close("containers")
	return c
}

// readContainer reads one container: what it holds is never more than it can
// hold, which is judged only where the capacity could be read.
func readContainer(o *object) container {
	var read container
	read.capacity, _ = o.whole("capacity", minCapacity, maxCapacity, true)
	most := read.capacity
	if most == 0 {
		most = maxCapacity
	}
	read.amount, read.amountHeld = o.bounded("amount", 0, most, true,
		"must be a whole number from 0 up to the container's capacity")
	read.label = o.label("label", false)
	o.close("a container")
	return read
}

func (containers) Kind() Kind { return Containers }

func (c containers) Labels() []string {
	var labels []string
	for _, item := range c.items {
		labels = append(labels, item.label)
	}
	return present(labels...)
}

// Shown are each container's capacity, what it holds, and its label.
func (c containers) Shown() []Shown {
	var shown []Shown
	for _, item := range c.items {
		if item.capacity > 0 {
			shown = append(shown, number(item.capacity))
		}
		if item.amountHeld {
			shown = append(shown, number(item.amount))
		}
		shown = append(shown, texts(item.label)...)
	}
	return shown
}
