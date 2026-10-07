package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// The next task is written ahead: once a task is on the card, the model writes
// the one after it while the child works, and the service keeps it, sealed,
// until the child asks for another — then it comes at once, on the card the ask
// draws. What the tools that ask for a task and write one ahead share is here:
// the lesson as it stands, which a task written ahead has to fit; letting go of
// one that no longer does; and handing out the task kept.

// aheadNextText sends the model on to write the next task ahead, once a task
// is on the card.
const aheadNextText = "Now call prepare_task: it gives you the package of the next task, which you write ahead " +
	"while the child works on this one, and which is kept until the child asks for another. Say nothing to the " +
	"child about that task."

// lessonNow is the lesson as it stands for a task written ahead: its language
// — the parent's choice, or else language, the chat's —, the topic the
// lessons are kept to, the skills kept out of the tasks, the instructions
// tasks are written to now, and where a person asked this task to stand.
func (s *Service) lessonNow(p *profile.Profile, language string, asked profile.Place) profile.Lesson {
	return profile.Lesson{
		Language:            p.Student.LessonLanguage(language),
		Topic:               tutor.LessonTopic(p, s.content),
		ExcludedSkills:      p.Student.ExcludedSkills,
		InstructionsVersion: s.content.InstructionsVersion(),
		Asked:               asked,
	}
}

// askedPlace is where a person asked a task to stand, as a choice passed to
// the tool that asks for one says it beside the topic the lessons are kept
// to: naming that topic asks for nothing, since it is set anyway.
func (s *Service) askedPlace(p *profile.Profile, choice tutor.Choice) profile.Place {
	asked := choice.Beside(tutor.LessonTopic(p, s.content))
	return profile.Place{Topic: asked.Topic, GradeLevel: asked.GradeLevel, Difficulty: asked.Difficulty}
}

// letGone is a task written ahead that was let go: why, whether it had been
// written and kept or was still being written, where it stood, and the
// version of the instructions it was written to.
type letGone struct {
	why     profile.Misfit
	written bool
	at      profile.Place
	version string
}

// letGo lets go of the task written ahead — kept, or still being written —
// that no longer fits the lesson, and says what it let go and why. Nothing is
// written here: the caller writes the profile, and then the lines.
func (s *Service) letGo(p *profile.Profile, lesson *profile.Lesson, now time.Time) []letGone {
	var gone []letGone
	if ready := p.ReadyTask; ready != nil {
		if why := ready.Misfit(lesson); why != profile.MisfitNone {
			p.DropReady()
			gone = append(gone, letGone{why: why, written: true,
				at:      profile.Place{Topic: ready.Topic, GradeLevel: ready.GradeLevel, Difficulty: ready.Difficulty},
				version: ready.InstructionsVersion})
		}
	}
	if open := s.aheadOpen(p, now); open != nil {
		if why := open.Misfit(lesson); why != profile.MisfitNone {
			p.CloseRequest()
			gone = append(gone, letGone{why: why,
				at:      profile.Place{Topic: open.Brief.TargetConcept, GradeLevel: open.Brief.GradeLevel, Difficulty: open.Brief.Difficulty},
				version: s.content.InstructionsVersion()})
		}
	}
	return gone
}

// sayLetGo leaves the line of each task written ahead that was let go, once
// the profile that lets it go is written: why, whether it had been written,
// and where it stood. The topic is read from the profile, which a person can
// edit, so a line names it only while the catalog has it, and each is counted
// against the instructions it was written to.
func (s *Service) sayLetGo(ctx context.Context, account store.Account, gone []letGone) {
	for _, task := range gone {
		s.events.writeFor(ctx, account, task.version, eventTaskDropped,
			zap.String("reason", string(task.why)),
			zap.Bool("written", task.written),
			zap.String("topic", s.topicLabel(task.at.Topic)),
			zap.String("level", string(task.at.GradeLevel)),
			zap.Int("difficulty", task.at.Difficulty),
		)
	}
}

// aheadOpen is the request open for the task written ahead, while it is still
// waited for, or nil.
func (s *Service) aheadOpen(p *profile.Profile, now time.Time) *profile.OpenRequest {
	if open := p.OpenRequest; open != nil && open.Ahead && open.Awaited(s.window, now) {
		return open
	}
	return nil
}

// waitedOpen is the request open for a task the child waits for, while it is
// still waited for, or nil.
func (s *Service) waitedOpen(p *profile.Profile, now time.Time) *profile.OpenRequest {
	if open := p.OpenRequest; open != nil && !open.Ahead && open.Awaited(s.window, now) {
		return open
	}
	return nil
}

// waitForAhead makes the task being written ahead the one the child waits
// for: they asked for the next task before it was handed in. The task left on
// the card without an answer is recorded as skipped, the profile written, and
// the lines left: the task skipped, and the request, waited for now. asked
// says a person asked for where the task stands. It is the request, and the
// task skipped, when there was one.
func (s *Service) waitForAhead(ctx context.Context, account store.Account, p *profile.Profile, revision store.Revision,
	asked bool, now time.Time,
) (*profile.OpenRequest, *profile.Answer, error) {
	left, wasSkipped := p.Skip(now)
	request := p.OpenRequest
	request.Await(now)
	request.Asked = asked
	p.Touch(s.version, now)
	if _, err := s.store.Save(ctx, account, p, revision); err != nil {
		return nil, nil, fmt.Errorf("mcp: save the profile: %w", err)
	}

	var skipped *profile.Answer
	if wasSkipped {
		skipped = &left
		s.events.write(ctx, account, eventTaskSkipped, s.skippedFields(left.Topic, left.GradeLevel, left.Difficulty)...)
	}
	s.events.write(ctx, account, eventTaskRequested, requestFields(request, true)...)
	return request, skipped, nil
}

// handOutKept is the hand-out of the kept task taskID to the child, who asked
// for the next one, as a change any read of the profile can be given: the task
// kept becomes the one on the card, and the one left there without an answer
// is skipped. A profile with that task on the card already holds the change,
// and one that no longer keeps it allows it no more. Once the file holds the
// change, the lines of the hand-out are left: the task skipped, and the task
// kept, accepted now, as ready. asked says a person asked for where the task
// stands, which the task kept stood at: the task written ahead after it is
// then one more like it.
func (s *Service) handOutKept(account store.Account, taskID string, asked bool, now time.Time) change {
	return func(p *profile.Profile) (made, error) {
		switch {
		case p.CurrentTask != nil && p.CurrentTask.ID == taskID:
			return made{state: already}, nil
		case p.ReadyTask == nil || p.ReadyTask.ID != taskID:
			return made{state: gone}, nil
		}
		ready, left := *p.ReadyTask, p.InFlight()
		task, err := p.HandOutReady(now)
		if err != nil {
			return made{}, fmt.Errorf("mcp: hand out the task kept: %w", err)
		}
		task.Asked = asked
		p.Touch(s.version, now)
		return made{state: changed, landed: func(ctx context.Context) {
			if left != nil {
				s.events.write(ctx, account, eventTaskSkipped, s.skippedFields(left.Topic, left.GradeLevel, left.Difficulty)...)
			}
			s.events.write(ctx, account, eventTaskAccepted, s.acceptedLine(ctx, p, account, task, &handedOut{
				attempts: ready.Attempts, written: ready.WrittenAt.Sub(ready.OpenedAt.Time), ready: true,
			}, now)...)
		}}, nil
	}
}

// handedOut is how a task handed out came to the child: the attempts and the
// time its writing took, and whether it was written ahead and kept ready.
type handedOut struct {
	attempts int
	written  time.Duration
	ready    bool
}

// acceptedLine is the line of a task handed out: where it stands, how it came
// to the child, whether it came with a drawing — never the drawing — and what
// the child is counted by.
func (s *Service) acceptedLine(ctx context.Context, p *profile.Profile, account store.Account, task *profile.CurrentTask,
	how *handedOut, now time.Time,
) []zap.Field {
	return append([]zap.Field{
		zap.String("topic", task.Topic),
		zap.String("level", string(task.GradeLevel)),
		zap.Int("difficulty", task.Difficulty),
		zap.Int("attempts", how.attempts),
		zap.Int64("seconds_since_request", int64(how.written/time.Second)),
		zap.Bool("drawing", strings.TrimSpace(task.Drawing) != ""),
		zap.Bool("ready", how.ready),
	}, s.acceptedFields(ctx, p, account, task, now)...)
}
