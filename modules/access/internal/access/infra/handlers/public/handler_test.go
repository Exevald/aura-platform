package public

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestModuleRecipientsAlwaysIncludeOwnerOnce(t *testing.T) {
	ownerID, memberID := uuid.New(), uuid.New()

	recipients := moduleRecipients(ownerID, []uuid.UUID{memberID, ownerID, memberID})

	require.Equal(t, []uuid.UUID{ownerID, memberID}, recipients)
}

func TestModuleRecipientsAllowOwnerOnlyGroup(t *testing.T) {
	ownerID := uuid.New()

	require.Equal(t, []uuid.UUID{ownerID}, moduleRecipients(ownerID, nil))
}
