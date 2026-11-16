package termview

import "testing"

func TestFrameUpdateTouchesOnlyChangedRows(t *testing.T) {
	prev := []string{"header", "alpha", "beta", "footer"}
	if got := frameUpdate(prev, prev); got != "" {
		t.Fatalf("unchanged frame produced output: %q", got)
	}
	next := []string{"header", "ALPHA", "beta", "footer"}
	if got, want := frameUpdate(prev, next), "\x1b[2;1HALPHA\x1b[K"; got != want {
		t.Fatalf("frame update = %q, want %q", got, want)
	}
	shorter := []string{"header", "alpha"}
	if got, want := frameUpdate(prev, shorter), "\x1b[3;1H\x1b[K\x1b[4;1H\x1b[K"; got != want {
		t.Fatalf("shrink update = %q, want %q", got, want)
	}
}
