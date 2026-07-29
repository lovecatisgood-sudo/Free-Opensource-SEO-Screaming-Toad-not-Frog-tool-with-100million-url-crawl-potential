package contracts

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// ID is an opaque, stable identifier exposed across application boundaries.
type ID string

// NewID creates a 128-bit random identifier with a type prefix. Prefixes make
// logs and exported records easier to inspect without encoding authority.
func NewID(prefix string) (ID, error) {
	if prefix == "" {
		return "", fmt.Errorf("id prefix is required")
	}
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return ID(prefix + "_" + hex.EncodeToString(b)), nil
}
