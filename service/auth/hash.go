package auth

import (
	"crypto/rand"
	"encoding/base64"
	"slices"

	"encore.dev/beta/errs"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/argon2"
)

type Argon2Config struct {
	Memory     uint32 `json:"memory"`
	Time       uint32 `json:"iterations"`
	Threads    uint8  `json:"threads"`
	SaltLength uint32 `json:"salt_length"`
	KeyLength  uint32 `json:"key_length"`
}

type Hash struct {
	Argon2Config
	Salt string `json:"salt"` // base64 encoded
	Hash string `json:"hash"` // base64 encoded
}

// NewHash creates a new Argon2 hash from the given secret.
func NewHash(secret string) (string, error) {
	config := Argon2Config{
		Memory:     19 * 1024, // 19 MiB
		Time:       2,
		Threads:    1,
		SaltLength: 16,
		KeyLength:  32,
	}

	salt := make([]byte, config.SaltLength)
	rand.Read(salt)

	hash := Hash{
		Argon2Config: config,

		Salt: base64.StdEncoding.EncodeToString(salt),
		Hash: base64.StdEncoding.EncodeToString(
			argon2.IDKey(
				[]byte(secret),
				salt,
				config.Time,
				config.Memory,
				config.Threads,
				config.KeyLength)),
	}

	out, err := msgpack.Marshal(&hash)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(out), nil
}

// FromHash decodes a base64-encoded hash string into a Hash struct.
func FromHash(data string) (*Hash, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	var hash Hash
	if err := msgpack.Unmarshal(b, &hash); err != nil {
		return nil, err
	}

	return &hash, nil
}

// VerifyHash checks if the provided password matches the stored hash.
func VerifyHash(r string, h *Hash) (*bool, error) {
	st, err := base64.StdEncoding.DecodeString(h.Salt)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}
	pd, err := base64.StdEncoding.DecodeString(h.Hash)
	if err != nil {
		return nil, &errs.Error{
			Code: errs.Internal,
		}
	}

	pi := argon2.IDKey([]byte(r), st, h.Time, h.Memory, h.Threads, h.KeyLength)

	result := slices.Equal(pd, pi)

	return &result, nil
}
