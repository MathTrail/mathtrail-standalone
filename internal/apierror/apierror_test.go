package apierror_test

import (
	"encoding/json"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
)

func TestResponseJSONKeys(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(apierror.Response{Code: apierror.CodeNotFound, Message: "no such endpoint"})
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}

	var fields map[string]string
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}
	if fields["code"] != apierror.CodeNotFound {
		t.Errorf("code = %q, want %q", fields["code"], apierror.CodeNotFound)
	}
	if fields["message"] != "no such endpoint" {
		t.Errorf("message = %q, want %q", fields["message"], "no such endpoint")
	}
	if len(fields) != 2 {
		t.Errorf("the response has %d fields, want exactly code and message", len(fields))
	}
}

// A code is what a client branches on, so what it says on the wire is part of
// the contract: a code renamed in the source would break every client that
// knew the old one while the tests that compare with the constant stayed green.
func TestTheCodesSayWhatTheySaid(t *testing.T) {
	t.Parallel()

	for code, want := range map[string]string{
		apierror.CodeInternal:         "INTERNAL_ERROR",
		apierror.CodeNotFound:         "NOT_FOUND",
		apierror.CodeMethodNotAllowed: "METHOD_NOT_ALLOWED",
	} {
		t.Run(want, func(t *testing.T) {
			t.Parallel()
			if code != want {
				t.Errorf("code = %q, want %q", code, want)
			}
		})
	}
}
