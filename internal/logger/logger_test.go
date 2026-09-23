package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// The JSON encoder exists to be read by Cloud Logging without any further
// configuration, and that means three key names and one severity scale. A test
// that only checked "it logs something" would not notice the day one of them
// is renamed back to zap's default.
func TestJSONEncoderUsesCloudRunKeys(t *testing.T) {
	t.Parallel()

	encoder := zapcore.NewJSONEncoder(cloudRunEncoderConfig())
	entry := zapcore.Entry{
		Level:   zapcore.WarnLevel,
		Time:    time.Date(2026, 9, 21, 10, 30, 0, 0, time.UTC),
		Message: "tool_call",
	}

	buf, err := encoder.EncodeEntry(entry, []zapcore.Field{{
		Key:    "tool",
		Type:   zapcore.StringType,
		String: "get_profile",
	}})
	if err != nil {
		t.Fatalf("EncodeEntry() error = %v, want nil", err)
	}

	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("the line is not JSON: %v; line = %s", err, buf.String())
	}

	for key, want := range map[string]string{
		"severity": "WARNING",
		"message":  "tool_call",
		"time":     "2026-09-21T10:30:00Z",
		"tool":     "get_profile",
	} {
		got, ok := line[key].(string)
		if !ok {
			t.Errorf("line[%q] is missing; line = %s", key, buf.String())
			continue
		}
		if got != want {
			t.Errorf("line[%q] = %q, want %q", key, got, want)
		}
	}

	for _, absent := range []string{"level", "msg", "ts"} {
		if _, found := line[absent]; found {
			t.Errorf("line still has zap's default key %q; line = %s", absent, buf.String())
		}
	}
}

func TestSeverityNames(t *testing.T) {
	t.Parallel()

	cases := map[zapcore.Level]string{
		zapcore.DebugLevel: "DEBUG",
		zapcore.InfoLevel:  "INFO",
		zapcore.WarnLevel:  "WARNING",
		zapcore.ErrorLevel: "ERROR",
		zapcore.PanicLevel: "CRITICAL",
		zapcore.FatalLevel: "EMERGENCY",
	}

	for level, want := range cases {
		t.Run(want, func(t *testing.T) {
			t.Parallel()
			encoder := zapcore.NewJSONEncoder(cloudRunEncoderConfig())
			buf, err := encoder.EncodeEntry(zapcore.Entry{Level: level, Message: "x"}, nil)
			if err != nil {
				t.Fatalf("EncodeEntry() error = %v, want nil", err)
			}
			var line map[string]any
			if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
				t.Fatalf("the line is not JSON: %v", err)
			}
			if got := line["severity"]; got != want {
				t.Errorf("severity = %v, want %q", got, want)
			}
		})
	}
}

// A level nobody recognises must not stop the service: it falls back to info.
func TestNewFallsBackToInfo(t *testing.T) {
	t.Parallel()

	log := New("loud", "json")
	if log == nil {
		t.Fatal("New() = nil, want a logger")
	}
	if !log.Core().Enabled(zapcore.InfoLevel) {
		t.Error("info is disabled, want it enabled after an unknown level")
	}
	if log.Core().Enabled(zapcore.DebugLevel) {
		t.Error("debug is enabled, want info as the fallback level")
	}
}

// Two settings of the configuration are decisions, not defaults, and both are
// easy to lose to a future refactor that starts from zap's own constructors.
func TestConfigurationDecisions(t *testing.T) {
	t.Parallel()

	for _, format := range []string{"json", "console"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()

			cfg := newConfig("info", format)

			if cfg.Sampling != nil {
				t.Error("sampling is on, want it off: the counts are taken from these lines")
			}
			if len(cfg.OutputPaths) != 1 || cfg.OutputPaths[0] != "stdout" {
				t.Errorf("OutputPaths = %v, want [stdout]", cfg.OutputPaths)
			}
			if len(cfg.ErrorOutputPaths) != 1 || cfg.ErrorOutputPaths[0] != "stderr" {
				t.Errorf("ErrorOutputPaths = %v, want [stderr]", cfg.ErrorOutputPaths)
			}
		})
	}
}

// A request carries its own path into a log line, and the code scan is right
// that this is user input reaching a log. What it cannot see is the thing that
// makes it harmless: the encoder.
//
// The attack a scanner has in mind is a caller ending the line they are in and
// starting one of their own — a path of "/x\n{severity: ERROR, …}" becoming a
// second entry that nobody wrote. A JSON encoder escapes the newline, so the
// forgery stays inside the string it arrived in, and the fields it tried to
// forge keep the values this service gave them.
//
// This is the test that holds the encoder to that. Swap it for one that does
// not escape and the alert stops being theoretical.
func TestAFieldCannotForgeASecondLine(t *testing.T) {
	t.Parallel()

	var written bytes.Buffer
	log := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(cloudRunEncoderConfig()),
		zapcore.AddSync(&written),
		zapcore.DebugLevel,
	))

	forged := "/x\n{\"severity\":\"ERROR\",\"message\":\"a line nobody wrote\"}"
	log.Info("http_request", zap.String("path", forged))

	if extra := strings.Count(strings.TrimRight(written.String(), "\n"), "\n"); extra != 0 {
		t.Errorf("the log holds %d lines, want exactly one", extra+1)
	}

	var entry map[string]any
	if err := json.Unmarshal(written.Bytes(), &entry); err != nil {
		t.Fatalf("what was written is not one JSON object: %v", err)
	}
	if entry["path"] != forged {
		t.Errorf("path = %q, want it kept whole and escaped", entry["path"])
	}
	if entry["message"] != "http_request" {
		t.Errorf("message = %q, want the forgery unable to replace it", entry["message"])
	}
	if entry["severity"] != "INFO" {
		t.Errorf("severity = %q, want the forgery unable to raise it", entry["severity"])
	}
}
