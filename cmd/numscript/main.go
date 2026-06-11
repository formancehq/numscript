package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/formancehq/numscript/internal/cmd"

	"github.com/getsentry/sentry-go"
)

// This has to be set dynamically via the following flag:
// -ldflags "-X main.Version=0.0.1"
var Version string = "develop"

// sentryDsn is the Sentry endpoint used for crash reporting in release
// builds. It is a variable so tests can point it elsewhere.
var sentryDsn = "https://b8b6cfd5dab95e1258d80963c3db73bf@o4504394442539008.ingest.us.sentry.io/4507623538884608"

// telemetryEnabled reports whether crash reporting should be active for
// the given build version. Telemetry is disabled in development builds
// (version == "develop") and when the NUMSCRIPT_NO_TELEMETRY environment
// variable is set to any non-empty value.
func telemetryEnabled(version string) bool {
	return version != "develop" && os.Getenv("NUMSCRIPT_NO_TELEMETRY") == ""
}

func recoverPanic() {
	if !telemetryEnabled(Version) {
		return
	}

	r := recover()
	if r == nil {
		return
	}

	errMsg := fmt.Sprintf("[uncaught panic]@%s: %s\n%s\n", Version, r, string(debug.Stack()))
	fmt.Fprint(os.Stderr, errMsg)
	sentry.CaptureMessage(errMsg)
	sentry.Flush(2 * time.Second)
}

// setupCrashReporting initializes Sentry when telemetry is enabled for
// the given build version, and returns a flush function the caller must
// defer. It returns a no-op when telemetry is disabled.
func setupCrashReporting(version string) func() {
	if !telemetryEnabled(version) {
		return func() {}
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		AttachStacktrace: true,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Sentry: %v\n", err)
	}

	return func() { sentry.Flush(2 * time.Second) }
}

func main() {
	flush := setupCrashReporting(Version)
	defer flush()
	defer recoverPanic()

	cmd.Execute(cmd.CliOptions{
		Version: Version,
	})
}
