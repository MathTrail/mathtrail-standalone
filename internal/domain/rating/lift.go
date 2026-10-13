package rating

import "math"

// ProbeLift is the least a right answer to a probe raises the level in its
// topic. A probe is a task set in a topic put on pause because even its
// easiest task stood out of reach; a right answer to it is the sign the pause
// waited for, and the topic comes back by at least this much.
const ProbeLift = 0.1

// nudges bounds the steps of the least float64 that ReachedAt and Lifted take
// past a level worked out backwards: one or two always do, and the bound only
// keeps a loop over floats from running on a number that no step moves.
const nudges = 16

// ReachedAt is the lowest level at which a task at this point is in reach, as
// InReach reads it: the corridor's floor above the point's difficulty, raised
// by the least step a float64 takes until InReach agrees, since the floor
// worked out backwards can land a hair below it. A point that stands nowhere
// gives no number.
func ReachedAt(point Point) float64 {
	level := point.Beta() + logit(CorridorLow)
	for range nudges {
		if math.IsNaN(level) || InReach(level, point) {
			return level
		}
		level = math.Nextafter(level, math.Inf(1))
	}
	return level
}

// Lifted is the state a right answer to a probe leaves: the state the answer's
// step left, with the topic's correction raised where the step fell short, so
// that the level in the topic stands at least ProbeLift above before, where it
// stood ahead of the answer, and at least where its easiest point is in reach.
// The overall level and the counts stay as the step left them, since the lift
// is the topic's alone; a step that went further is left as it is, and so is
// one whose numbers are no numbers.
func Lifted(stepped State, before float64, easiest Point) State {
	target := max(before+ProbeLift, ReachedAt(easiest))
	level := stepped.Level()
	if math.IsNaN(level) || math.IsNaN(target) || level >= target {
		return stepped
	}
	lifted, aim := stepped, target
	for range nudges {
		lifted.Delta = aim - lifted.Theta
		if lifted.Level() >= target {
			break
		}
		aim = math.Nextafter(aim, math.Inf(1))
	}
	return lifted
}
