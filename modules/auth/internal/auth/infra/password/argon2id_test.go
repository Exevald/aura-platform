package password

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArgon2IDHashAndCompare(t *testing.T) {
	hasher := NewArgon2ID()
	hash, err := hasher.Hash("correct horse battery staple")
	require.NoError(t, err)
	require.NotContains(t, hash, "correct horse battery staple")
	require.NoError(t, hasher.Compare(hash, "correct horse battery staple"))
	require.Error(t, hasher.Compare(hash, "wrong password"))
	require.ErrorIs(t, hasher.Compare("invalid", "password"), ErrInvalidHash)
}
