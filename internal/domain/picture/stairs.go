package picture

// stairs is a staircase of steps from the floor up, with labels on some steps
// and arcs drawn from one step to another: the moves of whoever walks it. The
// card does not number the steps: counting them is the child's.
type stairs struct {
	noWords
	labels []string
}

// stepRule is what a step a mark or a jump names must be, as a refusal says
// it, and never how many steps there are: the child counts them, and the
// count may be the answer.
const stepRule = "must be a step of the staircase: a whole number from 0, the floor, to its top step"

func readStairs(o *object) Picture {
	var s stairs
	steps, _ := o.whole("steps", minSteps, maxSteps, true)
	top := steps
	if top == 0 {
		top = maxSteps
	}
	marked := map[int]bool{}
	for _, mark := range o.objects("marks", 0, maxStairMarks, "marks", false) {
		if mark == nil {
			continue
		}
		if step, held := mark.bounded("step", 0, top, true, stepRule); held {
			if marked[step] {
				mark.fault(mark.at("step"), "is a step another mark already stands on")
			}
			marked[step] = true
		}
		s.labels = append(s.labels, mark.label("label", true))
		mark.close("a mark of stairs")
	}
	jumped := map[[2]int]bool{}
	for _, jump := range o.objects("jumps", 0, maxStairJumps, "jumps", false) {
		if jump == nil {
			continue
		}
		from, fromHeld := jump.bounded("from", 0, top, true, stepRule)
		to, toHeld := jump.bounded("to", 0, top, true, stepRule)
		switch {
		case !fromHeld || !toHeld:
		case from == to:
			jump.fault(jump.path, "must go from one step to another")
		case jumped[[2]int{from, to}]:
			jump.fault(jump.path, "draws a jump already drawn")
		default:
			jumped[[2]int{from, to}] = true
		}
		jump.close("a jump of stairs")
	}
	o.close("stairs")
	return &s
}

func (*stairs) Kind() Kind { return Stairs }

func (s *stairs) Labels() []string { return present(s.labels...) }

func (s *stairs) Shown() []Shown { return texts(s.labels...) }
