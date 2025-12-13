package nostr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NIP-05 errors
var (
	ErrInvalidNIP05Format = errors.New("invalid NIP-05 identifier format")
	ErrNIP05FetchFailed   = errors.New("failed to fetch NIP-05 data")
	ErrNIP05NotFound      = errors.New("name not found in NIP-05 response")
	ErrNIP05InvalidPubkey = errors.New("NIP-05 response contains invalid pubkey")
)

// NIP05Result contains the result of a NIP-05 resolution
type NIP05Result struct {
	Name   string   `json:"name"`   // The resolved name part
	Domain string   `json:"domain"` // The domain queried
	Pubkey string   `json:"pubkey"` // Hex pubkey
	Npub   string   `json:"npub"`   // Bech32 npub
	Relays []string `json:"relays"` // Optional relay hints
}

// nip05Response represents the JSON structure from /.well-known/nostr.json
type nip05Response struct {
	Names  map[string]string   `json:"names"`
	Relays map[string][]string `json:"relays,omitempty"`
}

// ParseNIP05 parses a NIP-05 identifier into name and domain parts
// Accepts formats: "name@domain.com" or "_@domain.com"
func ParseNIP05(identifier string) (name, domain string, err error) {
	identifier = strings.TrimSpace(identifier)
	identifier = strings.ToLower(identifier)

	parts := strings.Split(identifier, "@")
	if len(parts) != 2 {
		return "", "", ErrInvalidNIP05Format
	}

	name = parts[0]
	domain = parts[1]

	// Name can be empty (will use "_" as per NIP-05 spec)
	if name == "" {
		name = "_"
	}

	// Basic domain validation
	if len(domain) < 3 || !strings.Contains(domain, ".") {
		return "", "", ErrInvalidNIP05Format
	}

	return name, domain, nil
}

// ResolveNIP05 resolves a NIP-05 identifier to a pubkey
// The identifier should be in the format "name@domain.com"
func ResolveNIP05(ctx context.Context, identifier string) (*NIP05Result, error) {
	name, domain, err := ParseNIP05(identifier)
	if err != nil {
		return nil, err
	}

	return ResolveNIP05Parts(ctx, name, domain)
}

// ResolveNIP05Parts resolves a NIP-05 using separate name and domain
func ResolveNIP05Parts(ctx context.Context, name, domain string) (*NIP05Result, error) {
	// Build the URL
	nip05URL := fmt.Sprintf("https://%s/.well-known/nostr.json?name=%s",
		domain, url.QueryEscape(name))

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nip05URL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNIP05FetchFailed, err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Roostr/1.0")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNIP05FetchFailed, err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d", ErrNIP05FetchFailed, resp.StatusCode)
	}

	// Parse response
	var nip05Resp nip05Response
	if err := json.NewDecoder(resp.Body).Decode(&nip05Resp); err != nil {
		return nil, fmt.Errorf("%w: invalid JSON response", ErrNIP05FetchFailed)
	}

	// Look up the name (case-insensitive)
	var hexPubkey string
	for n, pk := range nip05Resp.Names {
		if strings.EqualFold(n, name) {
			hexPubkey = pk
			break
		}
	}

	if hexPubkey == "" {
		return nil, ErrNIP05NotFound
	}

	// Validate and normalize the pubkey
	hexPubkey = strings.ToLower(hexPubkey)
	if !IsValidHexPubkey(hexPubkey) {
		return nil, ErrNIP05InvalidPubkey
	}

	// Convert to npub
	npub, err := EncodeNpub(hexPubkey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNIP05InvalidPubkey, err)
	}

	// Get relays if available
	var relays []string
	if nip05Resp.Relays != nil {
		relays = nip05Resp.Relays[hexPubkey]
	}

	return &NIP05Result{
		Name:   name,
		Domain: domain,
		Pubkey: hexPubkey,
		Npub:   npub,
		Relays: relays,
	}, nil
}

// ResolveIdentity attempts to resolve an identity string to a pubkey.
// It tries in order: npub, hex pubkey, NIP-05 identifier.
// Returns (hexPubkey, npub, source, nip05Name, error)
// source is one of: "npub", "hex", "nip05"
func ResolveIdentity(ctx context.Context, input string) (hexPubkey, npub, source, nip05Name string, err error) {
	input = strings.TrimSpace(input)

	// Try as npub first
	if strings.HasPrefix(strings.ToLower(input), "npub") {
		hexPubkey, npub, err = ValidatePubkey(input)
		if err != nil {
			return "", "", "", "", err
		}
		return hexPubkey, npub, "npub", "", nil
	}

	// Try as hex pubkey
	if len(input) == 64 && IsValidHexPubkey(input) {
		hexPubkey, npub, err = ValidatePubkey(input)
		if err != nil {
			return "", "", "", "", err
		}
		return hexPubkey, npub, "hex", "", nil
	}

	// Try as NIP-05 identifier
	if IsNIP05Identifier(input) {
		result, err := ResolveNIP05(ctx, input)
		if err != nil {
			return "", "", "", "", err
		}
		return result.Pubkey, result.Npub, "nip05", result.Name, nil
	}

	return "", "", "", "", ErrInvalidPubkey
}
