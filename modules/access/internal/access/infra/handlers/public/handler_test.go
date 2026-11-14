package public

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"access/internal/access/domain/model"
)

func TestModuleRecipientsAlwaysIncludeOwnerOnce(t *testing.T) {
	ownerID, memberID := uuid.New(), uuid.New()

	recipients := moduleRecipients(ownerID, []uuid.UUID{memberID, ownerID, memberID})

	require.Equal(t, []uuid.UUID{ownerID, memberID}, recipients)
}

func TestGroupAccessesDeduplicatesModules(t *testing.T) {
	groupID, moduleID := uuid.New(), uuid.New()
	values := []model.ModuleShare{
		{GroupID: groupID, ModuleID: moduleID, ShareID: uuid.New()},
		{GroupID: groupID, ModuleID: moduleID, ShareID: uuid.New()},
	}

	require.Equal(t, []uuid.UUID{moduleID}, groupAccesses(values)[groupID])
}

func TestModuleRecipientsAllowOwnerOnlyGroup(t *testing.T) {
	ownerID := uuid.New()

	require.Equal(t, []uuid.UUID{ownerID}, moduleRecipients(ownerID, nil))
}
