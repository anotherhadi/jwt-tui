package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"strings"
)

// Decode splits a JWT and returns pretty-printed header and payload JSON.
func Decode(token string) (header, payload string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("expected 3 parts, got %d", len(parts))
	}

	hdrBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", fmt.Errorf("header: %w", err)
	}
	plBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("payload: %w", err)
	}

	var hdrObj any
	if err := json.Unmarshal(hdrBytes, &hdrObj); err != nil {
		return "", "", fmt.Errorf("header JSON: %w", err)
	}
	var plObj any
	if err := json.Unmarshal(plBytes, &plObj); err != nil {
		return "", "", fmt.Errorf("payload JSON: %w", err)
	}

	hdrPretty, _ := json.MarshalIndent(hdrObj, "", "  ")
	plPretty, _ := json.MarshalIndent(plObj, "", "  ")

	return string(hdrPretty), string(plPretty), nil
}

// Encode builds and signs a JWT from raw JSON header and payload strings.
func Encode(header, payload, secret string) (string, error) {
	var hdrObj map[string]any
	if err := json.Unmarshal([]byte(header), &hdrObj); err != nil {
		return "", fmt.Errorf("header JSON: %w", err)
	}
	var plObj any
	if err := json.Unmarshal([]byte(payload), &plObj); err != nil {
		return "", fmt.Errorf("payload JSON: %w", err)
	}

	hdrCompact, _ := json.Marshal(hdrObj)
	plCompact, _ := json.Marshal(plObj)

	hdrB64 := base64.RawURLEncoding.EncodeToString(hdrCompact)
	plB64 := base64.RawURLEncoding.EncodeToString(plCompact)
	signingInput := hdrB64 + "." + plB64

	alg, _ := hdrObj["alg"].(string)

	h, err := hashForAlg(alg)
	if err != nil {
		return signingInput + ".", fmt.Errorf("%w", err)
	}
	if h == nil {
		return signingInput + ".", nil
	}

	mac := hmac.New(h, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := mac.Sum(nil)

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// Verify checks whether the JWT signature is valid for the given secret.
// Returns (false, nil) for an invalid signature, (true, nil) for valid.
func Verify(token, secret string) (bool, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false, fmt.Errorf("expected 3 parts, got %d", len(parts))
	}

	hdrBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false, fmt.Errorf("header encoding: %w", err)
	}
	var hdrObj map[string]any
	if err := json.Unmarshal(hdrBytes, &hdrObj); err != nil {
		return false, fmt.Errorf("header JSON: %w", err)
	}

	alg, _ := hdrObj["alg"].(string)

	h, err := hashForAlg(alg)
	if err != nil {
		return false, err
	}
	if h == nil {
		return parts[2] == "", nil
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(h, []byte(secret))
	mac.Write([]byte(signingInput))
	expected := mac.Sum(nil)

	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false, fmt.Errorf("signature encoding: %w", err)
	}

	return hmac.Equal(actual, expected), nil
}

// Algorithm returns the "alg" claim from the JWT header, or "" if unreadable.
func Algorithm(token string) string {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) < 1 {
		return ""
	}
	hdrBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return ""
	}
	var hdrObj map[string]any
	if err := json.Unmarshal(hdrBytes, &hdrObj); err != nil {
		return ""
	}
	alg, _ := hdrObj["alg"].(string)
	return alg
}

func hashForAlg(alg string) (func() hash.Hash, error) {
	switch strings.ToUpper(alg) {
	case "HS256":
		return sha256.New, nil
	case "HS384":
		return sha512.New384, nil
	case "HS512":
		return sha512.New, nil
	case "NONE", "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", alg)
	}
}
