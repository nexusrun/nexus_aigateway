package run

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envDemoMode = "AIGATEWAY_DEMO_MODE"
	// Accepted after the rename so deployments still setting the old name keep
	// working; the canonical name wins when both are present.
	envDemoModeLegacy       = "GOMODEL_DEMO_MODE"
	demoModeWarningInterval = 5 * time.Minute
)

func demoModeFromEnv() (bool, error) {
	name := envDemoMode
	raw := strings.TrimSpace(os.Getenv(envDemoMode))
	if raw == "" {
		name = envDemoModeLegacy
		raw = strings.TrimSpace(os.Getenv(envDemoModeLegacy))
	}
	if raw == "" {
		return false, nil
	}
	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("invalid %s %q: expected a boolean", name, raw)
	}
	return enabled, nil
}

func startDemoModeWarnings(ctx context.Context) {
	logDemoModeWarning()
	go repeatDemoModeWarnings(ctx, demoModeWarningInterval, logDemoModeWarning)
}

func repeatDemoModeWarnings(ctx context.Context, interval time.Duration, warn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			warn()
		}
	}
}

func logDemoModeWarning() {
	slog.Warn("DEMO MODE: this environment is public; do not enter sensitive or personal data; demo data is reset regularly",
		"demo_mode", true,
		"reminder_interval", demoModeWarningInterval,
	)
}
