package prefix

import "testing"

func TestPrefixDateAllocated(t *testing.T) {
	t.Parallel()

	cleared, err := prefixDateAllocated("")
	if err != nil {
		t.Fatalf("clear allocation date: %v", err)
	}
	if !cleared.IsSet() || cleared.Get() != nil {
		t.Fatalf("empty allocation date did not produce an explicit API null: %#v", cleared.Get())
	}

	const value = "2026-01-02T03:04:05Z"
	parsed, err := prefixDateAllocated(value)
	if err != nil {
		t.Fatalf("parse valid date: %v", err)
	}
	if !parsed.IsSet() || parsed.Get() == nil || parsed.Get().Format("2006-01-02T15:04:05Z07:00") != value {
		t.Fatalf("unexpected parsed date: %#v", parsed.Get())
	}
	if _, err := prefixDateAllocated("not-a-date"); err == nil {
		t.Fatal("expected an invalid RFC3339 date to fail")
	}
}
