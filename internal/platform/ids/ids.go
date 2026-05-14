package ids

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

func New(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "id"
	}

	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return prefix + "-" + time.Now().UTC().Format("20060102150405.000000000")
	}

	return prefix + "-" + time.Now().UTC().Format("20060102150405") + "-" + hex.EncodeToString(randomBytes)
}
