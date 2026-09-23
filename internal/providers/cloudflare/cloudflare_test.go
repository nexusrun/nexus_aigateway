package cloudflare

import "testing"

func TestRegistration(t *testing.T) {
	if Registration.Type != "cloudflare" {
		t.Fatalf("Registration.Type = %q, want cloudflare", Registration.Type)
	}
	if Registration.New == nil {
		t.Fatal("Registration.New is nil")
	}
	if !Registration.Discovery.RequireBaseURL {
		t.Fatal("Cloudflare registration must require an account-scoped base URL")
	}
}
