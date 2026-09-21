package apierror_test

import (
	"encoding/json"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/apierror"
)

func TestResponseJSONKeys(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(apierror.Response{Code: apierror.CodeBadRequest, Message: "field is required"})
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}

	var fields map[string]string
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}
	if fields["code"] != apierror.CodeBadRequest {
		t.Errorf("code = %q, want %q", fields["code"], apierror.CodeBadRequest)
	}
	if fields["message"] != "field is required" {
		t.Errorf("message = %q, want %q", fields["message"], "field is required")
	}
	if len(fields) != 2 {
		t.Errorf("the response has %d fields, want exactly code and message", len(fields))
	}
}
