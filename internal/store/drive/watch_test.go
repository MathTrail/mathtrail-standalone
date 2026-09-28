package drivestore_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/drive/drivetest"
)

// Every call to Drive is a span of its own, inside the span of whatever asked
// for it, and a line; both say which call it was and how it ended, and only a
// failure marks the span as failed.
func TestEveryCallIsASpanAndALine(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	f.spans.Reset()
	f.logs.TakeAll()
	cold := f.instance(t, 5*time.Second)
	failEveryTry(f.fake, drivetest.Download, http.StatusServiceUnavailable, "backendError")

	ctx, caller := f.traces.Tracer("drivestore_test").Start(t.Context(), "tools/call get_profile")
	_, _, err := cold.Load(ctx, mia)
	caller.End()
	if !errors.Is(err, drive.ErrUnavailable) {
		t.Fatalf("Load() error = %v, want %v", err, drive.ErrUnavailable)
	}
	if waited := f.waits.taken(); len(waited) != tries-1 {
		t.Errorf("the download that failed waited %v before its tries, want %d pauses", waited, tries-1)
	}

	var calls []sdktrace.ReadOnlySpan
	for _, span := range f.spans.Ended() {
		if span.Name() != "tools/call get_profile" {
			calls = append(calls, span)
		}
	}
	want := []struct {
		name, outcome string
		failed        bool
		level         zapcore.Level
		retries       int64
	}{
		{"drive list", "ok", false, zapcore.InfoLevel, 0},
		{"drive download", "unavailable", true, zapcore.WarnLevel, tries - 1},
	}
	if len(calls) != len(want) {
		t.Fatalf("spans of the calls = %d, want %d", len(calls), len(want))
	}
	lines := f.logs.FilterMessage("drive_call").All()
	if len(lines) != len(want) {
		t.Fatalf("drive_call lines = %d, want %d", len(lines), len(want))
	}
	for i, call := range want {
		span := calls[i]
		switch {
		case span.Name() != call.name:
			t.Errorf("span %d = %q, want %q", i, span.Name(), call.name)
		case span.Parent().SpanID() != caller.SpanContext().SpanID():
			t.Errorf("span %q is not inside the span of the call that asked for it", span.Name())
		case span.SpanKind() != trace.SpanKindClient:
			t.Errorf("span %q is of kind %v, want client", span.Name(), span.SpanKind())
		case attributeOf(span, "mathtrail.drive.outcome") != call.outcome:
			t.Errorf("span %q outcome = %q, want %q", span.Name(), attributeOf(span, "mathtrail.drive.outcome"), call.outcome)
		case (span.Status().Code == codes.Error) != call.failed:
			t.Errorf("span %q status = %v, want failed %t", span.Name(), span.Status(), call.failed)
		}

		line := &lines[i]
		fields := line.ContextMap()
		op := strings.TrimPrefix(call.name, "drive ")
		switch {
		case line.Level != call.level:
			t.Errorf("line of %s is at %v, want %v", op, line.Level, call.level)
		case fields["op"] != op || fields["outcome"] != call.outcome:
			t.Errorf("line = op %v, outcome %v, want %s, %s", fields["op"], fields["outcome"], op, call.outcome)
		case fields["user"] != mia.ID:
			t.Errorf("line of %s names user %v, want %q", op, fields["user"], mia.ID)
		case fields["retries"] != call.retries:
			t.Errorf("line of %s says %v retries, want %d", op, fields["retries"], call.retries)
		case fields["logging.googleapis.com/spanId"] != span.SpanContext().SpanID().String():
			t.Errorf("line of %s names span %v, want the span of its call", op, fields["logging.googleapis.com/spanId"])
		}
		if _, timed := fields["duration_ms"]; !timed {
			t.Errorf("line of %s has no duration_ms", op)
		}
	}
}

// A file that is not there is an answer about the parent's Drive, not a
// failure: the span of the call that met it is not marked as failed.
func TestAFileNotThereIsNoFailure(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia"))
	f.fake.Delete(miaToken, f.profileFile(t, miaToken).ID)
	f.spans.Reset()

	if _, _, err := f.storage.Load(t.Context(), mia); err == nil {
		t.Fatal("Load() of a deleted profile error = nil, want one")
	}
	for _, span := range f.spans.Ended() {
		if span.Status().Code == codes.Error {
			t.Errorf("span %q (%s) is marked as failed, want an answer", span.Name(), attributeOf(span, "mathtrail.drive.outcome"))
		}
	}
}

// The guard. Nothing of a file reaches a span or a line: not which file it is,
// not what it holds, not the token that reached it — whether the calls went
// well or not.
func TestNothingOfTheFileReachesASpanOrALine(t *testing.T) {
	t.Parallel()

	f := newFixture(t, 5*time.Second)
	f.create(t, mia, child("Mia Secretname"))
	p, read, err := f.storage.Load(t.Context(), mia)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if _, err := f.storage.Save(t.Context(), mia, moved(p, "Mia Secretname the brave"), read); err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}
	if _, err := f.storage.Export(t.Context(), mia); err != nil {
		t.Fatalf("Export() error = %v, want nil", err)
	}
	secrets := []string{miaToken, "Secretname", "File not found"}
	for _, file := range f.fake.Files(miaToken) {
		secrets = append(secrets, file.ID)
	}
	f.fake.Fail(drivetest.Download, http.StatusForbidden, "appNotAuthorizedToFile")
	_, _, _ = f.storage.Load(t.Context(), mia)
	f.fake.Delete(miaToken, f.profileFile(t, miaToken).ID)
	_, _, _ = f.storage.Load(t.Context(), mia)

	for _, span := range f.spans.Ended() {
		wantNoneOf(t, "span "+span.Name(), spanTexts(span), secrets)
	}
	lines := f.logs.All()
	for i := range lines {
		wantNoneOf(t, "line "+lines[i].Message, lineTexts(t, &lines[i]), secrets)
	}
}

// attributeOf is the value of a span's attribute, as text.
func attributeOf(span sdktrace.ReadOnlySpan, key attribute.Key) string {
	for _, attr := range span.Attributes() {
		if attr.Key == key {
			return attr.Value.String()
		}
	}
	return ""
}

// spanTexts is every text a span carries: its name, its status, its
// attributes and its events.
func spanTexts(span sdktrace.ReadOnlySpan) []string {
	texts := []string{span.Name(), span.Status().Description}
	for _, attr := range span.Attributes() {
		texts = append(texts, attr.Value.String())
	}
	for _, event := range span.Events() {
		texts = append(texts, event.Name)
		for _, attr := range event.Attributes {
			texts = append(texts, attr.Value.String())
		}
	}
	return texts
}

// lineTexts is every text a line carries: its message and its fields, as the
// encoder would write them.
func lineTexts(t *testing.T, line *observer.LoggedEntry) []string {
	t.Helper()

	encoded, err := json.Marshal(line.ContextMap())
	if err != nil {
		t.Fatalf("a line does not encode: %v", err)
	}
	return []string{line.Message, string(encoded)}
}

// wantNoneOf holds texts to carrying none of the secrets.
func wantNoneOf(t *testing.T, where string, texts, secrets []string) {
	t.Helper()

	for _, text := range texts {
		if i := slices.IndexFunc(secrets, func(secret string) bool { return strings.Contains(text, secret) }); i >= 0 {
			t.Errorf("%s carries %q in %q", where, secrets[i], text)
		}
	}
}
