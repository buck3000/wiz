package license

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTierString(t *testing.T) {
	tests := []struct {
		tier Tier
		want string
	}{
		{TierFree, "Free"},
		{TierPro, "Pro"},
		{TierTeam, "Team"},
		{TierEnterprise, "Enterprise"},
	}
	for _, tt := range tests {
		if got := tt.tier.String(); got != tt.want {
			t.Errorf("Tier(%d).String() = %q, want %q", int(tt.tier), got, tt.want)
		}
	}
}

func TestLimitsForTier(t *testing.T) {
	free := LimitsForTier(TierFree)
	if free.MaxContexts != 10 {
		t.Errorf("Free MaxContexts = %d, want 10", free.MaxContexts)
	}
	if free.OrchestraDeps {
		t.Error("Free should not have OrchestraDeps")
	}

	pro := LimitsForTier(TierPro)
	if pro.MaxContexts != 0 {
		t.Errorf("Pro MaxContexts = %d, want 0 (unlimited)", pro.MaxContexts)
	}
	if !pro.OrchestraDeps {
		t.Error("Pro should have OrchestraDeps")
	}
	if !pro.CostTracking {
		t.Error("Pro should have CostTracking")
	}

	team := LimitsForTier(TierTeam)
	if !team.AIReview || !team.TeamRegistry || !team.CIMode || !team.AuditLog {
		t.Error("Team should have all team features")
	}
}

func TestCheckContextLimit(t *testing.T) {
	// Free tier, under limit.
	if err := CheckContextLimit(TierFree, 5); err != nil {
		t.Errorf("unexpected error at 5/10: %v", err)
	}

	// Free tier, at limit.
	err := CheckContextLimit(TierFree, 10)
	if err == nil {
		t.Fatal("expected error at 10/10")
	}
	limErr, ok := err.(*ContextLimitErr)
	if !ok {
		t.Fatalf("expected *ContextLimitErr, got %T", err)
	}
	if limErr.Current != 10 || limErr.Max != 10 {
		t.Errorf("got %d/%d, want 10/10", limErr.Current, limErr.Max)
	}

	// Pro tier, unlimited.
	if err := CheckContextLimit(TierPro, 1000); err != nil {
		t.Errorf("Pro should be unlimited, got: %v", err)
	}
}

func TestValidateKey(t *testing.T) {
	// Valid Pro key.
	key := GenerateKey("test@example.com", TierPro, time.Now().Add(24*time.Hour))
	tier, err := validateKey(key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierPro {
		t.Errorf("got tier %v, want Pro", tier)
	}

	// Expired key falls back to Free.
	expired := GenerateKey("test@example.com", TierPro, time.Now().Add(-1*time.Hour))
	tier, err = validateKey(expired)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierFree {
		t.Errorf("expired key got tier %v, want Free", tier)
	}

	// Invalid token falls back to Free.
	tier, err = validateKey("not-a-jwt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierFree {
		t.Errorf("invalid token got tier %v, want Free", tier)
	}

	// Tampered signature falls back to Free.
	tier, err = validateKey(key + "tampered")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierFree {
		t.Errorf("tampered key got tier %v, want Free", tier)
	}
}

func TestCheckLicenseEnvVar(t *testing.T) {
	ResetCache()
	defer ResetCache()

	key := GenerateKey("env@example.com", TierTeam, time.Now().Add(24*time.Hour))
	t.Setenv("WIZ_LICENSE_KEY", key)

	tier, err := CheckLicense()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierTeam {
		t.Errorf("got tier %v, want Team", tier)
	}
}

func TestCheckLicenseFile(t *testing.T) {
	ResetCache()
	defer ResetCache()

	// Unset env var so file is checked.
	t.Setenv("WIZ_LICENSE_KEY", "")

	// Write a temporary license file.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	dir := filepath.Join(tmpHome, ".config", "wiz")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	key := GenerateKey("file@example.com", TierPro, time.Now().Add(24*time.Hour))
	data := []byte(`{"key":"` + key + `"}`)
	if err := os.WriteFile(filepath.Join(dir, "license.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	tier, err := CheckLicense()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierPro {
		t.Errorf("got tier %v, want Pro", tier)
	}
}

func TestCheckLicenseNoKey(t *testing.T) {
	ResetCache()
	defer ResetCache()

	t.Setenv("WIZ_LICENSE_KEY", "")
	// Point HOME to a temp dir with no license file.
	t.Setenv("HOME", t.TempDir())

	tier, err := CheckLicense()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != TierFree {
		t.Errorf("got tier %v, want Free", tier)
	}
}
