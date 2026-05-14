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
		prefix = "ID"
	}

	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return prefix + "-" + time.Now().UTC().Format("20060102150405.000000000")
	}

	return prefix + "-" + time.Now().UTC().Format("20060102150405") + "-" + hex.EncodeToString(randomBytes)
}

func NewPersonID() string {
	return New("PER")
}

func NewOrganizationID() string {
	return New("ORG")
}

func NewOrganizationCapabilityID() string {
	return New("CAP")
}

func NewMembershipID() string {
	return New("MEM")
}

func NewMembershipRoleID() string {
	return New("ROLE")
}

func NewCredentialID() string {
	return New("CRD")
}

func NewEventID() string {
	return New("EVT")
}

func NewCorrelationID() string {
	return New("corr")
}
