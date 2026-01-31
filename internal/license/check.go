package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// SigningKeyHex is the HMAC-SHA256 key used to verify license JWTs.
// Set at build time via:
//
//	-ldflags "-X github.com/firewood-buck-3000/wiz/internal/license.SigningKeyHex=<hex>"
//
// Falls back to a dev-only default if unset.
var SigningKeyHex = "77697a2d6465762d6b65792d6e6f742d666f722d70726f64" // "wiz-dev-key-not-for-prod" in hex

func getSigningKey() []byte {
	b, err := hex.DecodeString(SigningKeyHex)
	if err != nil {
		return []byte("wiz-dev-key-not-for-prod")
	}
	return b
}

var (
	cachedTier   Tier
	cachedOnce   sync.Once
	cachedErr    error
)

// licenseFile holds the JSON structure of ~/.config/wiz/license.json.
type licenseFile struct {
	Key string `json:"key"`
}

// jwtHeader is the JWT header (we only support HS256).
type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// jwtPayload is the JWT payload for a wiz license.
type jwtPayload struct {
	Sub  string `json:"sub"`            // email
	Tier string `json:"tier"`           // free, pro, team, enterprise
	Exp  int64  `json:"exp"`            // expiry (unix)
	Iat  int64  `json:"iat"`            // issued at (unix)
}

// CheckLicense resolves the current license tier. It checks:
// 1. WIZ_LICENSE_KEY environment variable
// 2. ~/.config/wiz/license.json
// 3. Falls back to TierFree
//
// The result is cached for the lifetime of the process.
func CheckLicense() (Tier, error) {
	cachedOnce.Do(func() {
		cachedTier, cachedErr = resolveLicense()
	})
	return cachedTier, cachedErr
}

// ResetCache clears the cached license result (for testing).
func ResetCache() {
	cachedOnce = sync.Once{}
	cachedTier = TierFree
	cachedErr = nil
}

func resolveLicense() (Tier, error) {
	// 1. Environment variable.
	if key := os.Getenv("WIZ_LICENSE_KEY"); key != "" {
		return validateKey(key)
	}

	// 2. License file.
	home, err := os.UserHomeDir()
	if err != nil {
		return TierFree, nil
	}
	path := filepath.Join(home, ".config", "wiz", "license.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return TierFree, nil // no file = free
	}
	var lf licenseFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return TierFree, nil
	}
	if lf.Key == "" {
		return TierFree, nil
	}
	return validateKey(lf.Key)
}

// validateKey decodes and verifies a JWT license key, returning the tier.
func validateKey(token string) (Tier, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return TierFree, nil // invalid token = free, not error
	}

	// Verify signature.
	signingInput := parts[0] + "." + parts[1]
	sig, err := base64URLDecode(parts[2])
	if err != nil {
		return TierFree, nil
	}
	mac := hmac.New(sha256.New, getSigningKey())
	mac.Write([]byte(signingInput))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return TierFree, nil // bad signature = free
	}

	// Decode payload.
	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return TierFree, nil
	}
	var payload jwtPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return TierFree, nil
	}

	// Check expiry.
	if payload.Exp > 0 && time.Now().Unix() > payload.Exp {
		return TierFree, nil // expired = free
	}

	return parseTier(payload.Tier), nil
}

func parseTier(s string) Tier {
	switch strings.ToLower(s) {
	case "pro":
		return TierPro
	case "team":
		return TierTeam
	case "enterprise":
		return TierEnterprise
	default:
		return TierFree
	}
}

func base64URLDecode(s string) ([]byte, error) {
	// Add padding if needed.
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

// GenerateKey creates a signed JWT license key (for tooling/testing).
func GenerateKey(email string, tier Tier, expiry time.Time) string {
	header := base64URLEncode(mustJSON(jwtHeader{Alg: "HS256", Typ: "JWT"}))
	payload := base64URLEncode(mustJSON(jwtPayload{
		Sub:  email,
		Tier: tier.String(),
		Exp:  expiry.Unix(),
		Iat:  time.Now().Unix(),
	}))
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, getSigningKey())
	mac.Write([]byte(signingInput))
	sig := base64URLEncode(mac.Sum(nil))
	return signingInput + "." + sig
}

func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func mustJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
