// Package nostr provides Nostr protocol utilities including bech32 encoding,
// pubkey validation, and NIP-05 resolution.
package nostr

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// Bech32 character set for encoding
const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// charsetMap is the reverse lookup for decoding
var charsetMap = func() map[rune]int {
	m := make(map[rune]int)
	for i, c := range charset {
		m[c] = i
	}
	return m
}()

// Bech32 errors
var (
	ErrInvalidBech32      = errors.New("invalid bech32 string")
	ErrInvalidChecksum    = errors.New("invalid bech32 checksum")
	ErrInvalidHRP         = errors.New("invalid human-readable part")
	ErrInvalidDataPart    = errors.New("invalid data part")
	ErrInvalidPubkey      = errors.New("invalid pubkey: must be 64 hex characters or valid npub")
	ErrInvalidNpub        = errors.New("invalid npub format")
	ErrInvalidHexPubkey   = errors.New("invalid hex pubkey: must be 64 characters")
)

// polymod calculates the BCH checksum
func polymod(values []int) int {
	gen := []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := 1
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if (top>>i)&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}

// hrpExpand expands the human-readable part for checksum computation
func hrpExpand(hrp string) []int {
	result := make([]int, len(hrp)*2+1)
	for i, c := range hrp {
		result[i] = int(c) >> 5
	}
	result[len(hrp)] = 0
	for i, c := range hrp {
		result[len(hrp)+1+i] = int(c) & 31
	}
	return result
}

// verifyChecksum verifies the bech32 checksum
func verifyChecksum(hrp string, data []int) bool {
	values := append(hrpExpand(hrp), data...)
	return polymod(values) == 1
}

// createChecksum creates the bech32 checksum
func createChecksum(hrp string, data []int) []int {
	values := append(hrpExpand(hrp), data...)
	values = append(values, []int{0, 0, 0, 0, 0, 0}...)
	mod := polymod(values) ^ 1
	result := make([]int, 6)
	for i := 0; i < 6; i++ {
		result[i] = (mod >> (5 * (5 - i))) & 31
	}
	return result
}

// convertBits converts between bit widths
func convertBits(data []byte, fromBits, toBits int, pad bool) ([]int, error) {
	acc := 0
	bits := 0
	result := []int{}
	maxv := (1 << toBits) - 1

	for _, value := range data {
		acc = (acc << fromBits) | int(value)
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, (acc>>bits)&maxv)
		}
	}

	if pad {
		if bits > 0 {
			result = append(result, (acc<<(toBits-bits))&maxv)
		}
	} else if bits >= fromBits || ((acc<<(toBits-bits))&maxv) != 0 {
		return nil, ErrInvalidDataPart
	}

	return result, nil
}

// DecodeBech32 decodes a bech32 string into its HRP and data parts
func DecodeBech32(bech string) (string, []byte, error) {
	bech = strings.ToLower(bech)

	// Find separator
	pos := strings.LastIndex(bech, "1")
	if pos < 1 || pos+7 > len(bech) {
		return "", nil, ErrInvalidBech32
	}

	hrp := bech[:pos]
	dataPart := bech[pos+1:]

	// Validate HRP
	for _, c := range hrp {
		if c < 33 || c > 126 {
			return "", nil, ErrInvalidHRP
		}
	}

	// Decode data part
	data := make([]int, len(dataPart))
	for i, c := range dataPart {
		idx, ok := charsetMap[c]
		if !ok {
			return "", nil, ErrInvalidDataPart
		}
		data[i] = idx
	}

	// Verify checksum
	if !verifyChecksum(hrp, data) {
		return "", nil, ErrInvalidChecksum
	}

	// Remove checksum from data
	data = data[:len(data)-6]

	// Convert from 5-bit to 8-bit
	result := make([]byte, 0)
	acc := 0
	bits := 0
	for _, value := range data {
		acc = (acc << 5) | value
		bits += 5
		for bits >= 8 {
			bits -= 8
			result = append(result, byte((acc>>bits)&0xff))
		}
	}

	return hrp, result, nil
}

// EncodeBech32 encodes data with the given HRP into a bech32 string
func EncodeBech32(hrp string, data []byte) (string, error) {
	// Convert 8-bit data to 5-bit
	conv, err := convertBits(data, 8, 5, true)
	if err != nil {
		return "", err
	}

	// Create checksum
	checksum := createChecksum(hrp, conv)

	// Build result
	var result strings.Builder
	result.WriteString(hrp)
	result.WriteRune('1')
	for _, d := range conv {
		result.WriteByte(charset[d])
	}
	for _, d := range checksum {
		result.WriteByte(charset[d])
	}

	return result.String(), nil
}

// DecodeNpub decodes an npub bech32 string to a hex pubkey
func DecodeNpub(npub string) (string, error) {
	hrp, data, err := DecodeBech32(npub)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidNpub, err)
	}

	if hrp != "npub" {
		return "", fmt.Errorf("%w: expected 'npub' prefix, got '%s'", ErrInvalidNpub, hrp)
	}

	if len(data) != 32 {
		return "", fmt.Errorf("%w: expected 32 bytes, got %d", ErrInvalidNpub, len(data))
	}

	return hex.EncodeToString(data), nil
}

// EncodeNpub encodes a hex pubkey to an npub bech32 string
func EncodeNpub(hexPubkey string) (string, error) {
	// Validate hex string
	if len(hexPubkey) != 64 {
		return "", ErrInvalidHexPubkey
	}

	data, err := hex.DecodeString(hexPubkey)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidHexPubkey, err)
	}

	return EncodeBech32("npub", data)
}

// ValidatePubkey validates a pubkey input (hex or npub) and returns both formats
// Returns (hexPubkey, npub, error)
func ValidatePubkey(input string) (string, string, error) {
	input = strings.TrimSpace(input)

	// Try as npub first
	if strings.HasPrefix(strings.ToLower(input), "npub") {
		hexPubkey, err := DecodeNpub(input)
		if err != nil {
			return "", "", err
		}
		// Re-encode to get canonical npub
		npub, _ := EncodeNpub(hexPubkey)
		return hexPubkey, npub, nil
	}

	// Try as hex pubkey
	if len(input) == 64 {
		// Validate it's valid hex
		_, err := hex.DecodeString(input)
		if err != nil {
			return "", "", ErrInvalidPubkey
		}

		// Encode to npub
		npub, err := EncodeNpub(input)
		if err != nil {
			return "", "", err
		}
		return strings.ToLower(input), npub, nil
	}

	return "", "", ErrInvalidPubkey
}

// IsValidHexPubkey checks if a string is a valid 64-character hex pubkey
func IsValidHexPubkey(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// IsValidNpub checks if a string is a valid npub
func IsValidNpub(s string) bool {
	_, err := DecodeNpub(s)
	return err == nil
}

// IsNIP05Identifier checks if a string looks like a NIP-05 identifier (user@domain)
func IsNIP05Identifier(s string) bool {
	parts := strings.Split(s, "@")
	if len(parts) != 2 {
		return false
	}
	name, domain := parts[0], parts[1]
	// Basic validation
	return len(name) > 0 && len(domain) > 2 && strings.Contains(domain, ".")
}
