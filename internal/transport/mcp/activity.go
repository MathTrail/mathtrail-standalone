package mcpserver

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/country"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
)

// unknownPlace is a country or a region nobody said.
const unknownPlace = "unknown"

// cohortLayout writes the month a profile was made in.
const cohortLayout = "2006-01"

// hostKey holds the chat host a call came from. It is a type of its own, so
// that nothing else a context carries can be taken for it.
type hostKey struct{}

// withHost is how the boundary hands a call the chat host it came from, as the
// word its own line writes it in: the line about a task handed out counts the
// task by it, and is kept where no line of the call is.
func withHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, hostKey{}, host)
}

// hostFrom is the chat host the boundary put in ctx, or unknown for a call
// that came by no boundary.
func hostFrom(ctx context.Context) string {
	if host, given := ctx.Value(hostKey{}).(string); given {
		return host
	}
	return "unknown"
}

// countedFields are what every line that counts a child carries about the
// child: the name the child is counted under this month, the grade the parent
// gave, and the month the profile was made in, which is when the child started.
// Nothing here leads back to a child: the name is a keyed digest that changes
// with the month, and the grade and the month are what many children share.
func (s *Service) countedFields(p *profile.Profile, now time.Time) []zap.Field {
	return []zap.Field{
		zap.String("learner", s.learners.Of(p.StudentID, now)),
		zap.Int("grade", p.Student.Grade),
		zap.String("cohort", p.CreatedAt.Format(cohortLayout)),
	}
}

// acceptedFields are what the line about a task handed out adds to the child's
// counted fields: which chat host it was handed out in, the language it was
// written in, and where the family is — the country and the region the parent
// gave, and the country the parent's browser signed in from. A country or a
// region is a code of the list, unknown when nobody said, and other when the
// file, edited by hand, names one the list does not have.
func (s *Service) acceptedFields(ctx context.Context, p *profile.Profile, account store.Account, task *profile.CurrentTask, now time.Time) []zap.Field {
	return append(s.countedFields(p, now),
		zap.String("host", hostFrom(ctx)),
		zap.String("language", languageLabel(task.Language)),
		zap.String("country", countryLabel(p.Student.Country)),
		zap.String("region", regionLabel(p.Student.Country, p.Student.Region)),
		zap.String("signin_country", countryLabel(account.SignInCountry)),
	)
}

// countryLabel is a country of the list, unknown for none, or other.
func countryLabel(code string) string {
	switch {
	case code == "":
		return unknownPlace
	case country.Known(code):
		return code
	}
	return other
}

// regionLabel is a region of the country, unknown for none, or other — a
// region of another country among them.
func regionLabel(countryCode, region string) string {
	switch {
	case region == "":
		return unknownPlace
	case country.KnownRegion(countryCode, region):
		return region
	}
	return other
}
