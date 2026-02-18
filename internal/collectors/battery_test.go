package collectors

import (
	"testing"
	"time"
)

func TestParsePmsetLogOutput(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	tests := []struct {
		name          string
		lines         string
		wantStartPct  int
		wantPlugCount int
	}{
		{
			name:          "no data from today",
			lines:         "2020-01-01 10:00:00 -0700 Assertions Summary- Using AC(Charge: 50)\n",
			wantStartPct:  -1,
			wantPlugCount: 0,
		},
		{
			name: "single AC entry today",
			lines: today + " 00:30:00 -0700 Assertions Summary- Using AC(Charge: 80)\n" +
				today + " 08:00:00 -0700 Assertions Summary- Using AC(Charge: 95)\n",
			wantStartPct:  80,
			wantPlugCount: 0,
		},
		{
			name: "one plug event (Batt to AC)",
			lines: today + " 06:00:00 -0700 Sleep Using Batt (Charge:100%)\n" +
				today + " 08:00:00 -0700 Assertions Summary- Using Batt(Charge: 90)\n" +
				today + " 10:00:00 -0700 Assertions Summary- Using AC(Charge: 85)\n",
			wantStartPct:  100,
			wantPlugCount: 1,
		},
		{
			name: "multiple plug events",
			lines: today + " 06:00:00 -0700 Using AC(Charge: 100)\n" +
				today + " 07:00:00 -0700 Using Batt(Charge: 95)\n" +
				today + " 08:00:00 -0700 Using AC(Charge: 90)\n" +
				today + " 09:00:00 -0700 Using Batt(Charge: 85)\n" +
				today + " 10:00:00 -0700 Using AC(Charge: 80)\n",
			wantStartPct:  100,
			wantPlugCount: 2,
		},
		{
			name:          "empty input",
			lines:         "",
			wantStartPct:  -1,
			wantPlugCount: 0,
		},
		{
			name: "AC to AC is not a plug event",
			lines: today + " 06:00:00 -0700 Using AC(Charge: 80)\n" +
				today + " 07:00:00 -0700 Using AC(Charge: 85)\n",
			wantStartPct:  80,
			wantPlugCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startPct, plugCount := parsePmsetLogOutput(tt.lines)
			if startPct != tt.wantStartPct {
				t.Errorf("startPct = %d, want %d", startPct, tt.wantStartPct)
			}
			if plugCount != tt.wantPlugCount {
				t.Errorf("plugCount = %d, want %d", plugCount, tt.wantPlugCount)
			}
		})
	}
}
