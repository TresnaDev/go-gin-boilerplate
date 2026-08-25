package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// Generate creates a cryptographically random opaque refresh token.
// raw is returned to the client; hash is what gets stored in the database
// so a leaked DB dump can't be replayed as a valid refresh token.
func Generate() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	hash = Hash(raw)
	return raw, hash, nil
}

func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
