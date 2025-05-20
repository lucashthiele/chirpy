package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)

	encoded := hex.EncodeToString(key)

	return encoded
}
