package collectors

import "testing"

func TestParseLoadStats(t *testing.T) {
	stats, err := parseLoadStats([]byte("1.25 2.50 3.75 4/10 123\n"))
	if err != nil || stats.Load1 != 1.25 || stats.Load5 != 2.50 || stats.Load15 != 3.75 {
		t.Fatalf("unexpected load stats: %#v, %v", stats, err)
	}
}

func TestParseLoadStatsRejectsMalformedInput(t *testing.T) {
	if _, err := parseLoadStats([]byte("1.00 2.00")); err == nil {
		t.Fatal("expected short loadavg input to be rejected")
	}
	if _, err := parseLoadStats([]byte("nope 2.00 3.00")); err == nil {
		t.Fatal("expected non-numeric loadavg input to be rejected")
	}
}
