package main

import (
	"os"
	"testing"
)

func TestTelemetryDisabledInDevelopBuilds(t *testing.T) {
	t.Setenv("NUMSCRIPT_NO_TELEMETRY", "")

	if telemetryEnabled("develop") {
		t.Error("expected telemetry to be disabled when version is \"develop\"")
	}
}

func TestTelemetryDisabledWhenOptOutVarIsSet(t *testing.T) {
	for _, value := range []string{"1", "true", "anything"} {
		t.Setenv("NUMSCRIPT_NO_TELEMETRY", value)

		if telemetryEnabled("0.0.1") {
			t.Errorf("expected telemetry to be disabled when NUMSCRIPT_NO_TELEMETRY=%q", value)
		}
	}
}

func TestTelemetryEnabledInReleaseBuildsByDefault(t *testing.T) {
	// t.Setenv with an empty value ensures the variable is treated as
	// unset by telemetryEnabled, and restores any pre-existing value
	// from the test environment afterwards.
	t.Setenv("NUMSCRIPT_NO_TELEMETRY", "")

	if !telemetryEnabled("0.0.1") {
		t.Error("expected telemetry to be enabled in release builds when NUMSCRIPT_NO_TELEMETRY is not set")
	}
}

// setVersion overrides the build-time Version global for the duration of
// a test and restores the previous value afterwards.
func setVersion(t *testing.T, version string) {
	t.Helper()
	previous := Version
	Version = version
	t.Cleanup(func() { Version = previous })
}

// setSentryDsn overrides the Sentry DSN for the duration of a test so
// that no real endpoint is ever configured.
func setSentryDsn(t *testing.T, dsn string) {
	t.Helper()
	previous := sentryDsn
	sentryDsn = dsn
	t.Cleanup(func() { sentryDsn = previous })
}

// silenceStderr redirects os.Stderr to the null device for the duration
// of a test, to keep expected error output out of the test logs.
func silenceStderr(t *testing.T) {
	t.Helper()
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("failed to open %s: %v", os.DevNull, err)
	}
	previous := os.Stderr
	os.Stderr = devNull
	t.Cleanup(func() {
		os.Stderr = previous
		_ = devNull.Close()
	})
}

func TestSetupCrashReportingIsNoopWhenTelemetryDisabled(t *testing.T) {
	t.Setenv("NUMSCRIPT_NO_TELEMETRY", "1")

	flush := setupCrashReporting("0.0.1")
	if flush == nil {
		t.Fatal("expected a non-nil flush function")
	}
	flush()
}

func TestSetupCrashReportingInitializesSentryWhenEnabled(t *testing.T) {
	t.Setenv("NUMSCRIPT_NO_TELEMETRY", "")
	// An empty DSN makes sentry.Init succeed with a disabled client, so
	// nothing is ever sent over the network.
	setSentryDsn(t, "")

	flush := setupCrashReporting("0.0.1")
	if flush == nil {
		t.Fatal("expected a non-nil flush function")
	}
	flush()
}

func TestSetupCrashReportingReportsInitErrors(t *testing.T) {
	t.Setenv("NUMSCRIPT_NO_TELEMETRY", "")
	setSentryDsn(t, "not-a-valid-dsn")
	silenceStderr(t)

	// Must not panic; the error is reported on stderr and execution
	// continues without crash reporting.
	flush := setupCrashReporting("0.0.1")
	flush()
}

func TestRecoverPanicIsNoopWhenTelemetryDisabled(t *testing.T) {
	setVersion(t, "0.0.1")
	t.Setenv("NUMSCRIPT_NO_TELEMETRY", "1")

	defer func() {
		if recover() == nil {
			t.Error("expected the panic to propagate when telemetry is disabled")
		}
	}()

	defer recoverPanic()
	panic("boom")
}

func TestRecoverPanicIsNoopWithoutPanic(t *testing.T) {
	setVersion(t, "0.0.1")
	t.Setenv("NUMSCRIPT_NO_TELEMETRY", "")

	// Must not capture anything or crash when no panic occurred.
	recoverPanic()
}

func TestRecoverPanicCapturesPanicWhenTelemetryEnabled(t *testing.T) {
	setVersion(t, "0.0.1")
	t.Setenv("NUMSCRIPT_NO_TELEMETRY", "")
	// Bind a disabled Sentry client (empty DSN) to the global hub so
	// CaptureMessage is guaranteed to be a no-op.
	setSentryDsn(t, "")
	setupCrashReporting("0.0.1")()
	silenceStderr(t)

	func() {
		defer recoverPanic()
		panic("boom")
	}()
	// Reaching this point means recoverPanic swallowed the panic.
}
