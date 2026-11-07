package checks

import (
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// unread holds the solver of a task solved from its picture to the picture: a
// program that left a member with facts unread on either run carries those
// facts in its own code, or proves the answer from a part of what is drawn,
// and so is no witness to the picture. A word task's solver is held to none
// of it, and a solver that never ran to an answer is not judged, since it may
// have stopped before it read what it would have.
//
// The refusal names the members by their paths and never what they hold, and
// it does not ask for a member to be left out of the picture: a task whose
// facts stand on the picture keeps them there.
func (r *reviewer) unread(answer *solver.Agreement, task *Task, asked *profile.Brief) (problems []Problem, unchecked string) {
	switch {
	case !asked.FromPicture():
		return nil, ""
	case task == nil:
		return nil, "what the solver read of the picture was not checked: that needs a task that can be read"
	case !ran(answer):
		return nil, "what the solver read of the picture was not checked: that needs a solver that runs to an answer"
	}

	facts := picture.Facts(task.Picture)
	var left []string
	for _, run := range answer.Runs {
		for _, name := range facts {
			if !slices.Contains(run.Read, name) {
				left = append(left, "task."+pictureField+"."+picture.NamedSafely(name))
			}
		}
	}
	if len(left) == 0 {
		return nil, ""
	}
	slices.Sort(left)
	return []Problem{{Code: CodeSolverUnread, Message: "the solver left " + strings.Join(slices.Compact(left), ", ") +
		" unread: a task solved from the picture takes every fact from its picture, so solve takes it as a second" +
		" argument, solve(options, picture), and works the answer out from what it reads there rather than from" +
		" numbers written into the program"}}, ""
}
