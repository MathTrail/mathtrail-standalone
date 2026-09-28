package drive

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
)

// maxRefusal is the most of a refusal that is read: a status, and the reasons
// Drive gives for it.
const maxRefusal = 64 << 10

// maxReason is the longest word of Drive's list of reasons a refusal names.
const maxReason = 64

// refusalOf says which refusal an answer of Drive's is: by its status, and for
// a 403 by the reason Drive gives, because Drive answers 403 for a pause it
// asks for, a Drive that is full, a token that grants nothing of Drive, a file
// the service may no longer reach and a revision it gives no content of. A
// refusal that is none of those is told by its status and the first reason
// Drive gave. The sentence Drive writes beside a reason is never read: it
// names the file.
func refusalOf(resp *http.Response) error {
	var answer struct {
		Error struct {
			Errors []struct {
				Reason string `json:"reason"`
			} `json:"errors"`
		} `json:"error"`
	}
	// An answer that is not Drive's error gives no reason, and its status
	// decides alone.
	_ = json.NewDecoder(io.LimitReader(resp.Body, maxRefusal)).Decode(&answer)
	reasons := make([]string, 0, len(answer.Error.Errors))
	for _, given := range answer.Error.Errors {
		reasons = append(reasons, given.Reason)
	}

	status := resp.StatusCode
	switch {
	case status == http.StatusUnauthorized:
		return ErrUnauthorized
	case status == http.StatusNotFound:
		return ErrNotFound
	case status == http.StatusTooManyRequests:
		return ErrRateLimited
	case status >= http.StatusInternalServerError:
		return fmt.Errorf("%w: status %d", ErrUnavailable, status)
	case status == http.StatusForbidden && slices.ContainsFunc(reasons, isPause):
		return ErrRateLimited
	case status == http.StatusForbidden && slices.Contains(reasons, "storageQuotaExceeded"):
		return ErrStorageFull
	case status == http.StatusForbidden && slices.Contains(reasons, "insufficientPermissions"):
		// The token was taken, and grants nothing of Drive: to the parent's
		// files it is as good as no token.
		return ErrUnauthorized
	case status == http.StatusForbidden && slices.ContainsFunc(reasons, isOutOfReach):
		return ErrNotFound
	case status == http.StatusForbidden && slices.ContainsFunc(reasons, isNotKept):
		return ErrNotKept
	}
	return fmt.Errorf("refused with status %d, reason %s", status, firstReason(reasons))
}

// isOutOfReach reports whether a reason is Drive saying that the service may
// no longer reach a file: to the service, a file that is not there.
func isOutOfReach(reason string) bool {
	return reason == "appNotAuthorizedToFile" || reason == "insufficientFilePermissions"
}

// isNotKept reports whether a reason is Drive refusing the content of a
// revision that is not kept forever. Drive's guide to its errors names the
// reason in one spelling and shows it in another, and both are taken.
func isNotKept(reason string) bool {
	return reason == "downloadRestrictedForRevision" || reason == "download_restricted_for_revision"
}

// isPause reports whether a reason is Drive asking for a pause: too many calls
// for this parent or for the service, or the service's calls for the day
// spent, which is a pause until the next.
func isPause(reason string) bool {
	return reason == "userRateLimitExceeded" || reason == "rateLimitExceeded" || reason == "dailyLimitExceeded"
}

// firstReason is the first reason Drive gave, when it has the shape of a word
// of Drive's list — letters, and the underscores some are written with — and
// "none" when no reason has.
func firstReason(reasons []string) string {
	for _, reason := range reasons {
		if isWord(reason) {
			return reason
		}
	}
	return "none"
}

// isWord reports whether a reason is made of letters and underscores alone,
// as every reason Drive documents is.
func isWord(reason string) bool {
	if reason == "" || len(reason) > maxReason {
		return false
	}
	for _, r := range reason {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '_' {
			return false
		}
	}
	return true
}
