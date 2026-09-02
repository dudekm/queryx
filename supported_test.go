package queryx

import (
	"sort"
	"testing"
)

func TestClient_SupportedTypes_Defaults(t *testing.T) {
	client := NewClientWithDefaults()

	types := client.SupportedTypes()
	if len(types) == 0 {
		t.Fatal("expected default client to support at least one type")
	}

	// Result must be sorted (stable output for API consumers).
	if !sort.StringsAreSorted(types) {
		t.Errorf("SupportedTypes() is not sorted: %v", types)
	}

	// A few well-known types must be present.
	for _, want := range []string{"minecraft", "cs2", "rust", "fivem"} {
		if !contains(types, want) {
			t.Errorf("expected %q in supported types, got %v", want, types)
		}
	}
}

func TestClient_SupportedTypes_EmptyWithoutDefaults(t *testing.T) {
	client := NewClient()

	if got := client.SupportedTypes(); len(got) != 0 {
		t.Errorf("expected no supported types on bare client, got %v", got)
	}
}

func TestClient_Supports(t *testing.T) {
	client := NewClientWithDefaults()

	if !client.Supports(ServerMinecraft) {
		t.Error("expected client to support minecraft")
	}
	if client.Supports(ServerType("definitely-not-a-real-game")) {
		t.Error("expected client to not support an unknown type")
	}
}

func contains(haystack []string, needle string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}
