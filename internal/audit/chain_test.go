package audit

import (
	"testing"
	"time"
)

func TestChainVerificationDetectsTampering(t *testing.T) {
	first, err := NewEvent("case", "created", "actor", "req-1", 1, time.Unix(1, 0), map[string]any{"b": 2, "a": 1}, "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewEvent("case", "updated", "actor", "req-2", 2, time.Unix(2, 0), map[string]any{"ok": true}, first.EventHash)
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify([]Event{first, second}); err != nil {
		t.Fatal(err)
	}
	second.Payload = jsonRaw([]byte(`{"ok":false}`))
	if err := Verify([]Event{first, second}); err == nil {
		t.Fatal("tampered payload should fail")
	}
}
