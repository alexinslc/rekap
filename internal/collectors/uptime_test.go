package collectors

import (
	"testing"
	"time"
)

func TestParseSleepWakeEvents(t *testing.T) {
	loc := time.FixedZone("MST", -7*3600)
	start := time.Date(2026, 2, 18, 0, 0, 0, 0, loc)
	end := time.Date(2026, 2, 18, 14, 0, 0, 0, loc)

	tests := []struct {
		name     string
		output   string
		start    time.Time
		end      time.Time
		wantMins int
	}{
		{
			name:     "no sleep events",
			output:   "",
			start:    start,
			end:      end,
			wantMins: 0,
		},
		{
			name: "single sleep/wake pair",
			output: `2026-02-18 01:00:00 -0700  Sleep  Entering Sleep state due to Idle Sleep
2026-02-18 08:00:00 -0700  Wake   Wake from Deep Idle [CDNVA]`,
			start:    start,
			end:      end,
			wantMins: 420, // 7 hours
		},
		{
			name: "multiple sleep/wake cycles",
			output: `2026-02-18 01:00:00 -0700  Sleep  Entering Sleep state
2026-02-18 08:00:00 -0700  Wake   Wake from Deep Idle
2026-02-18 12:00:00 -0700  Sleep  Entering Sleep state
2026-02-18 12:30:00 -0700  Wake   Wake from Deep Idle`,
			start:    start,
			end:      end,
			wantMins: 450, // 7h + 30m
		},
		{
			name: "open sleep interval (machine still sleeping)",
			output: `2026-02-18 13:00:00 -0700  Sleep  Entering Sleep state`,
			start:    start,
			end:      end,
			wantMins: 60, // 1 hour until end
		},
		{
			name: "events before start time are ignored",
			output: `2026-02-17 23:00:00 -0700  Sleep  Entering Sleep state
2026-02-18 06:00:00 -0700  Wake   Wake from Deep Idle`,
			start:    time.Date(2026, 2, 18, 8, 0, 0, 0, loc),
			end:      end,
			wantMins: 0,
		},
		{
			name: "malformed timestamps are skipped",
			output: `not a timestamp Sleep Entering Sleep
2026-02-18 10:00:00 -0700  Sleep  Entering Sleep state
2026-02-18 10:30:00 -0700  Wake   Wake from Deep Idle`,
			start:    start,
			end:      end,
			wantMins: 30,
		},
		{
			name: "duplicate sleep events only count first",
			output: `2026-02-18 10:00:00 -0700  Sleep  Entering Sleep state
2026-02-18 10:05:00 -0700  Sleep  Entering Sleep state
2026-02-18 10:30:00 -0700  Wake   Wake from Deep Idle`,
			start:    start,
			end:      end,
			wantMins: 30,
		},
		{
			name: "wake without prior sleep is ignored",
			output: `2026-02-18 08:00:00 -0700  Wake   Wake from Deep Idle
2026-02-18 10:00:00 -0700  Sleep  Entering Sleep state
2026-02-18 10:30:00 -0700  Wake   Wake from Deep Idle`,
			start:    start,
			end:      end,
			wantMins: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSleepWakeEvents(tt.output, tt.start, tt.end)
			gotMins := int(got.Minutes())
			if gotMins != tt.wantMins {
				t.Errorf("parseSleepWakeEvents() = %d mins, want %d mins", gotMins, tt.wantMins)
			}
		})
	}
}
