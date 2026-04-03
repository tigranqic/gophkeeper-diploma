// Package models defines the plaintext payload structures for each type of
// secret stored in GophKeeper. Payloads are JSON-marshalled and then
// AES-256-GCM encrypted before being sent to the server.
package models

import "encoding/json"

// RecordType is a numeric identifier for the kind of secret stored.
type RecordType int32

const (
	// TypeLogin identifies a login/password pair record.
	TypeLogin RecordType = 1
	// TypeText identifies an arbitrary text record.
	TypeText RecordType = 2
	// TypeBinary identifies an arbitrary binary data record.
	TypeBinary RecordType = 3
	// TypeCard identifies a bank card record.
	TypeCard RecordType = 4
)

// LoginPayload holds a login/password pair and optional metadata.
type LoginPayload struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Metadata string `json:"metadata"`
}

// TextPayload holds arbitrary text data and optional metadata.
type TextPayload struct {
	Text     string `json:"text"`
	Metadata string `json:"metadata"`
}

// BinaryPayload holds arbitrary binary data and optional metadata.
type BinaryPayload struct {
	Data     []byte `json:"data"`
	Metadata string `json:"metadata"`
}

// CardPayload holds bank card details and optional metadata.
type CardPayload struct {
	Number   string `json:"number"`
	Expiry   string `json:"expiry"`
	CVC      string `json:"cvc"`
	Holder   string `json:"holder"`
	Metadata string `json:"metadata"`
}

// MarshalPayload JSON-encodes any payload struct to bytes.
func MarshalPayload(payload interface{}) ([]byte, error) {
	return json.Marshal(payload)
}

// UnmarshalPayload JSON-decodes bytes into the given payload pointer.
func UnmarshalPayload(data []byte, payload interface{}) error {
	return json.Unmarshal(data, payload)
}
