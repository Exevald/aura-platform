package password

import (
	"auth/internal/auth/app"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      = 19 * 1024
	iterations  = 2
	parallelism = 1
	saltLength  = 16
	keyLength   = 32
)

var ErrInvalidHash = errors.New("invalid password hash")

type Argon2ID struct{}

func NewArgon2ID() app.PasswordHasher {
	return &Argon2ID{}
}

func (*Argon2ID) Hash(value string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(value), salt, iterations, memory, parallelism, keyLength)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (*Argon2ID) Compare(encoded, value string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return ErrInvalidHash
	}
	var version int
	var parsedMemory uint32
	var parsedIterations uint32
	var parsedParallelism uint8
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return ErrInvalidHash
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &parsedMemory, &parsedIterations, &parsedParallelism); err != nil {
		return ErrInvalidHash
	}
	if parsedMemory == 0 || parsedIterations == 0 || parsedParallelism == 0 {
		return ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) == 0 {
		return ErrInvalidHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) == 0 {
		return ErrInvalidHash
	}
	if len(want) > math.MaxUint32 {
		return ErrInvalidHash
	}
	//nolint:gosec // The explicit MaxUint32 check above makes the conversion safe.
	parsedKeyLength := uint32(len(want))
	got := argon2.IDKey([]byte(value), salt, parsedIterations, parsedMemory, parsedParallelism, parsedKeyLength)
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return errors.New("password does not match")
	}
	return nil
}
