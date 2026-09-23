package huggingface

import "testing"

func TestRegistration(t *testing.T) {
	if Registration.Type != "huggingface" {
		t.Fatalf("Registration.Type = %q, want huggingface", Registration.Type)
	}
	if Registration.New == nil {
		t.Fatal("Registration.New is nil")
	}
	if got, want := Registration.Discovery.DefaultBaseURL, defaultBaseURL; got != want {
		t.Fatalf("default base URL = %q, want %q", got, want)
	}
}
