package virtualmodels

import "testing"

func TestParseScheduleConfig(t *testing.T) {
	cfg, err := parseScheduleConfig(map[string]any{
		"timezone":         "America/Phoenix",
		"peak_start":       "18:00",
		"peak_end":         "02:00",
		"peak_targets":     []any{"fast/model"},
		"off_peak_targets": []any{"cheap/model"},
	})
	if err != nil {
		t.Fatalf("parseScheduleConfig() error = %v", err)
	}
	if cfg.timezone.String() != "America/Phoenix" || len(cfg.peakTargets) != 1 || cfg.peakTargets[0] != "fast/model" {
		t.Fatalf("unexpected schedule config: %+v", cfg)
	}
	if cfg.peakStart != 18*60*60*1e9 || cfg.peakEnd != 2*60*60*1e9 {
		t.Fatalf("unexpected schedule window: %v-%v", cfg.peakStart, cfg.peakEnd)
	}
}

func TestParseScheduleConfigRejectsInvalidValues(t *testing.T) {
	cases := []map[string]any{
		{"timezone": "Not/AZone", "peak_targets": []any{"fast/model"}},
		{"peak_start": "9am", "peak_targets": []any{"fast/model"}},
		{"peak_targets": "fast/model"},
		{"peak_targets": []any{}, "off_peak_targets": []any{}},
	}
	for _, raw := range cases {
		if _, err := parseScheduleConfig(raw); err == nil {
			t.Fatalf("parseScheduleConfig(%v) error = nil", raw)
		}
	}
}
