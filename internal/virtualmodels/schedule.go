package virtualmodels

import (
	"fmt"
	"strings"
	"time"
	_ "time/tzdata"
)

const (
	scheduleTimezoneKey    = "timezone"
	schedulePeakStartKey   = "peak_start"
	schedulePeakEndKey     = "peak_end"
	schedulePeakTargetsKey = "peak_targets"
	scheduleOffTargetsKey  = "off_peak_targets"
)

type scheduleConfig struct {
	timezone     *time.Location
	peakStart    time.Duration
	peakEnd      time.Duration
	peakTargets  []string
	offTargets   []string
}

func validateScheduleConfig(raw map[string]any) error {
	_, err := parseScheduleConfig(raw)
	return err
}

func parseScheduleConfig(raw map[string]any) (scheduleConfig, error) {
	if raw == nil {
		return scheduleConfig{}, fmt.Errorf("schedule strategy_config is required")
	}
	timezone := "UTC"
	if value, ok := raw[scheduleTimezoneKey].(string); ok && strings.TrimSpace(value) != "" {
		timezone = strings.TrimSpace(value)
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return scheduleConfig{}, fmt.Errorf("schedule timezone %q is invalid: %w", timezone, err)
	}
	start, err := parseScheduleTime(raw, schedulePeakStartKey, "09:00")
	if err != nil { return scheduleConfig{}, err }
	end, err := parseScheduleTime(raw, schedulePeakEndKey, "17:00")
	if err != nil { return scheduleConfig{}, err }
	peakTargets, err := scheduleTargetNames(raw[schedulePeakTargetsKey], schedulePeakTargetsKey)
	if err != nil { return scheduleConfig{}, err }
	offTargets, err := scheduleTargetNames(raw[scheduleOffTargetsKey], scheduleOffTargetsKey)
	if err != nil { return scheduleConfig{}, err }
	if len(peakTargets) == 0 && len(offTargets) == 0 {
		return scheduleConfig{}, fmt.Errorf("schedule requires peak_targets or off_peak_targets")
	}
	return scheduleConfig{timezone: location, peakStart: start, peakEnd: end, peakTargets: peakTargets, offTargets: offTargets}, nil
}

func parseScheduleTime(raw map[string]any, key, fallback string) (time.Duration, error) {
	value := fallback
	if candidate, ok := raw[key].(string); ok && strings.TrimSpace(candidate) != "" {
		value = strings.TrimSpace(candidate)
	}
	parsed, err := time.Parse("15:04", value)
	if err != nil {
		return 0, fmt.Errorf("schedule %s must use HH:MM, got %q", key, value)
	}
	return time.Duration(parsed.Hour())*time.Hour + time.Duration(parsed.Minute())*time.Minute, nil
}

func scheduleTargetNames(value any, key string) ([]string, error) {
	if value == nil { return nil, nil }
	var values []string
	switch list := value.(type) {
	case []string:
		values = list
	case []any:
		for _, item := range list {
			name, ok := item.(string)
			if !ok { return nil, fmt.Errorf("schedule %s must contain model names", key) }
			values = append(values, name)
		}
	default:
		return nil, fmt.Errorf("schedule %s must be a list of model names", key)
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" { result = append(result, value) }
	}
	return result, nil
}

func (s *Service) scheduledTarget(entry *redirectEntry, pool []resolvedTarget) (resolvedTarget, bool) {
	cfg, err := parseScheduleConfig(entry.vm.StrategyConfig)
	if err != nil { return resolvedTarget{}, false }
	local := time.Now().In(cfg.timezone)
	minute := time.Duration(local.Hour())*time.Hour + time.Duration(local.Minute())*time.Minute
	inPeak := minute >= cfg.peakStart && minute < cfg.peakEnd
	if cfg.peakStart > cfg.peakEnd { inPeak = minute >= cfg.peakStart || minute < cfg.peakEnd }
	names := cfg.offTargets
	if inPeak { names = cfg.peakTargets }
	for _, name := range names {
		if target, ok := poolTarget(pool, name); ok { return target, true }
	}
	return resolvedTarget{}, false
}
