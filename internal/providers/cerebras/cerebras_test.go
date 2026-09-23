package cerebras

import "testing"

func TestRegistration(t *testing.T) {
	if Registration.Type != "cerebras" {
		t.Fatalf("Registration.Type = %q, want cerebras", Registration.Type)
	}
	if Registration.New == nil {
		t.Fatal("Registration.New is nil")
	}
	if got, want := Registration.Discovery.DefaultBaseURL, defaultBaseURL; got != want {
		t.Fatalf("default base URL = %q, want %q", got, want)
	}
}
