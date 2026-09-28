package drivestore

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
)

// How a call ended is one word of a closed list, told by the error's kind and
// never by its text, and only the calls that failed — Drive's failures, or the
// service's — are marked as failures.
func TestEveryCallEndsInAWordAndOnlyFailuresFail(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		err     error
		outcome string
		failure bool
	}{
		{nil, outcomeOK, false},
		{fmt.Errorf("drive: download: %w", drive.ErrNotFound), outcomeNotFound, false},
		{fmt.Errorf("drive: download: %w", drive.ErrUnauthorized), outcomeRevoked, false},
		{fmt.Errorf("drive: update: %w", drive.ErrStorageFull), outcomeStorageFull, false},
		{fmt.Errorf("drive: download: %w", drive.ErrTooLarge), outcomeTooLarge, false},
		{fmt.Errorf("drive: list: %w", context.Canceled), outcomeCanceled, false},
		{fmt.Errorf("drive: list: %w", drive.ErrRateLimited), outcomeRateLimited, true},
		{fmt.Errorf("drive: list: %w: %w", drive.ErrUnavailable, context.DeadlineExceeded), outcomeTimeout, true},
		{fmt.Errorf("drive: list: %w", drive.ErrUnavailable), outcomeUnavailable, true},
		{errors.New("drive: list: refused with status 400, reason badRequest"), outcomeFailed, true},
	} {
		if got := outcomeOf(tc.err); got != tc.outcome {
			t.Errorf("outcomeOf(%v) = %q, want %q", tc.err, got, tc.outcome)
		}
		if got := isFailure(outcomeOf(tc.err)); got != tc.failure {
			t.Errorf("isFailure(%q) = %t, want %t", outcomeOf(tc.err), got, tc.failure)
		}
	}
}
