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

// telemetryEnabled reports whether crash reporting should be active.
// Telemetry is disabled in development builds (Version == "develop")
// and when the NUMSCRIPT_NO_TELEMETRY environment variable is set to
// any non-empty value.
func telemetryEnabled() bool {
	return Version != "develop" && os.Getenv("NUMSCRIPT_NO_TELEMETRY") == ""
}

func recoverPanic() {
	if !telemetryEnabled() {
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

func main() {
	if telemetryEnabled() {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              "https://b8b6cfd5dab95e1258d80963c3db73bf@o4504394442539008.ingest.us.sentry.io/4507623538884608",
			AttachStacktrace: true,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize Sentry: %v\n", err)
		}

		defer sentry.Flush(2 * time.Second)
	}

	defer recoverPanic()

	cmd.Execute(cmd.CliOptions{
		Version: Version,
	})
}
