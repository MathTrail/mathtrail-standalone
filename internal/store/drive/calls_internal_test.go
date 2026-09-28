package drivestore

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/store"
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
		{fmt.Errorf("drive: download revision: %w", drive.ErrNotKept), outcomeNotKept, false},
		{fmt.Errorf("drivestore: %w", store.ErrAccessExpired), outcomeExpired, true},
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

// A call is made again when Drive asked for a pause, whatever the call, and
// when Drive failed on its side on a call that changes nothing — or nothing
// more when made twice — but never a creation or an upload Drive failed on,
// which may have landed; and never a call that ran out of its time, was given
// up on, or was refused for any other reason.
func TestOnlyACallThatCannotHaveLandedIsMadeAgain(t *testing.T) {
	t.Parallel()

	paused := fmt.Errorf("drive: update: %w", drive.ErrRateLimited)
	failed := fmt.Errorf("drive: update: %w: status 503", drive.ErrUnavailable)
	timedOut := fmt.Errorf("drive: list: %w: %w", drive.ErrUnavailable, context.DeadlineExceeded)
	for _, tc := range []struct {
		op    string
		err   error
		again bool
	}{
		{opUpdate, paused, true},
		{opCreate, paused, true},
		{opDownload, paused, true},
		{opList, failed, true},
		{opGet, failed, true},
		{opDownload, failed, true},
		{opRevisions, failed, true},
		{opRevision, failed, true},
		{opKeep, failed, true},
		{opDelete, failed, true},
		{opUpdate, failed, false},
		{opCreate, failed, false},
		{opList, timedOut, false},
		{opList, fmt.Errorf("drive: list: %w", context.Canceled), false},
		{opDownload, fmt.Errorf("drive: download: %w", drive.ErrNotFound), false},
		{opUpdate, fmt.Errorf("drive: update: %w", drive.ErrStorageFull), false},
		{opList, fmt.Errorf("drive: list: %w", drive.ErrUnauthorized), false},
	} {
		if got := worthRepeating(tc.op, tc.err); got != tc.again {
			t.Errorf("worthRepeating(%s, %v) = %t, want %t", tc.op, tc.err, got, tc.again)
		}
	}
}
