package intent

import "testing"

func TestSeverityOf_HealthyLinkIsCalm(t *testing.T) {
	if got := SeverityOf(Snapshot{LinkUp: true}); got != Calm {
		t.Fatalf("severity = %v, want Calm", got)
	}
}

func TestSeverityOf_BigLossIsBad(t *testing.T) {
	if got := SeverityOf(Snapshot{LinkUp: true, Loss: 0.30}); got != Bad {
		t.Fatalf("severity = %v, want Bad", got)
	}
}

func TestSeverityOf_DownLinkIsBadEvenAtZeroPerturbation(t *testing.T) {
	if got := SeverityOf(Snapshot{LinkUp: false}); got != Bad {
		t.Fatalf("severity = %v, want Bad", got)
	}
}

func TestSeverityOf_PromotesToWorstAcrossKnobs(t *testing.T) {
	// Latency warning + loss bad → bad.
	s := Snapshot{LinkUp: true, LatencyMs: 100, Loss: 0.30}
	if got := SeverityOf(s); got != Bad {
		t.Fatalf("severity = %v, want Bad", got)
	}
}

func TestCompose_LossClauseUsesOneInNFraming(t *testing.T) {
	_, _, body := compose(Snapshot{LinkUp: true, Loss: 0.10})
	if !contains(body, "1 in 10 packets") {
		t.Fatalf("body = %q, want one-in-N framing", body)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
