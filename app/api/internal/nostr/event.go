package nostr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// Event verification errors
var (
	ErrInvalidEventID    = errors.New("invalid event ID")
	ErrInvalidEventSig   = errors.New("invalid event signature")
	ErrEventIDMismatch   = errors.New("event ID does not match content hash")
	ErrSignatureMismatch = errors.New("signature verification failed")
)

// SyncEvent represents a Nostr event received from a relay for syncing.
// Uses int64 for created_at to match the Nostr protocol wire format.
type SyncEvent struct {
	ID        string     `json:"id"`
	Pubkey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

// Serialize returns the canonical JSON serialization used for hashing.
// Format: [0, pubkey, created_at, kind, tags, content]
func (e *SyncEvent) Serialize() ([]byte, error) {
	// Build the canonical array for hashing
	canonical := []interface{}{
		0,           // Version (always 0)
		e.Pubkey,    // Pubkey as hex string
		e.CreatedAt, // Unix timestamp
		e.Kind,      // Event kind
		e.Tags,      // Tags array
		e.Content,   // Content string
	}

	return json.Marshal(canonical)
}

// ComputeID computes the expected event ID from the content.
// The ID is SHA256(Serialize()) encoded as hex.
func (e *SyncEvent) ComputeID() (string, error) {
	serialized, err := e.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize event: %w", err)
	}

	hash := sha256.Sum256(serialized)
	return hex.EncodeToString(hash[:]), nil
}

// VerifyID checks that the event ID matches the SHA256 hash of the serialized content.
func (e *SyncEvent) VerifyID() error {
	if len(e.ID) != 64 {
		return fmt.Errorf("%w: expected 64 hex characters, got %d", ErrInvalidEventID, len(e.ID))
	}

	computedID, err := e.ComputeID()
	if err != nil {
		return err
	}

	if computedID != e.ID {
		return ErrEventIDMismatch
	}

	return nil
}

// VerifySignature verifies the BIP-340 Schnorr signature against the event ID.
func (e *SyncEvent) VerifySignature() error {
	// Decode the pubkey (32 bytes x-only)
	pubkeyBytes, err := hex.DecodeString(e.Pubkey)
	if err != nil {
		return fmt.Errorf("%w: invalid pubkey hex: %v", ErrInvalidEventSig, err)
	}
	if len(pubkeyBytes) != 32 {
		return fmt.Errorf("%w: pubkey must be 32 bytes, got %d", ErrInvalidEventSig, len(pubkeyBytes))
	}

	// Decode the signature (64 bytes)
	sigBytes, err := hex.DecodeString(e.Sig)
	if err != nil {
		return fmt.Errorf("%w: invalid signature hex: %v", ErrInvalidEventSig, err)
	}
	if len(sigBytes) != 64 {
		return fmt.Errorf("%w: signature must be 64 bytes, got %d", ErrInvalidEventSig, len(sigBytes))
	}

	// Decode the event ID (the message that was signed)
	msgBytes, err := hex.DecodeString(e.ID)
	if err != nil {
		return fmt.Errorf("%w: invalid event ID hex: %v", ErrInvalidEventSig, err)
	}
	if len(msgBytes) != 32 {
		return fmt.Errorf("%w: event ID must be 32 bytes, got %d", ErrInvalidEventSig, len(msgBytes))
	}

	// Parse the x-only public key
	pubkey, err := schnorr.ParsePubKey(pubkeyBytes)
	if err != nil {
		return fmt.Errorf("%w: failed to parse pubkey: %v", ErrInvalidEventSig, err)
	}

	// Parse the signature
	sig, err := schnorr.ParseSignature(sigBytes)
	if err != nil {
		return fmt.Errorf("%w: failed to parse signature: %v", ErrInvalidEventSig, err)
	}

	// Verify the signature
	if !sig.Verify(msgBytes, pubkey) {
		return ErrSignatureMismatch
	}

	return nil
}

// Verify performs full event verification: ID hash check and signature verification.
func (e *SyncEvent) Verify() error {
	if err := e.VerifyID(); err != nil {
		return err
	}

	if err := e.VerifySignature(); err != nil {
		return err
	}

	return nil
}

// ParseEventFromRelay parses an event from a relay EVENT message.
// The input should be the event object (third element of ["EVENT", subID, event]).
func ParseEventFromRelay(data []byte) (*SyncEvent, error) {
	var event SyncEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}
	return &event, nil
}

// ToDBEvent converts a SyncEvent to a format suitable for database storage.
// This is a helper that returns the event with properly typed fields.
type DBEvent struct {
	ID        string
	Pubkey    string
	CreatedAt int64
	Kind      int
	Tags      [][]string
	Content   string
	Sig       string
}

// ToDBFormat converts the SyncEvent to DBEvent format.
func (e *SyncEvent) ToDBFormat() *DBEvent {
	return &DBEvent{
		ID:        e.ID,
		Pubkey:    e.Pubkey,
		CreatedAt: e.CreatedAt,
		Kind:      e.Kind,
		Tags:      e.Tags,
		Content:   e.Content,
		Sig:       e.Sig,
	}
}

// XOnlyPubKey returns the pubkey parsed as an x-only secp256k1 public key.
func (e *SyncEvent) XOnlyPubKey() (*btcec.PublicKey, error) {
	pubkeyBytes, err := hex.DecodeString(e.Pubkey)
	if err != nil {
		return nil, err
	}
	return schnorr.ParsePubKey(pubkeyBytes)
}
