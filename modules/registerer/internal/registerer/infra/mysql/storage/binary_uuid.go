package storage

import (
	"database/sql/driver"
	"fmt"
	"uuid"
)

type binaryUUID uuid.UUID

func (id binaryUUID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id binaryUUID) Value() (driver.Value, error) {
	return id[:], nil
}

func (id *binaryUUID) Scan(value any) error {
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("scan UUID: expected []byte, got %T", value)
	}
	if len(data) != len(id) {
		return fmt.Errorf("scan UUID: expected %d bytes, got %d", len(id), len(data))
	}
	copy(id[:], data)
	return nil
}

func newBinaryUUID(id uuid.UUID) binaryUUID {
	return binaryUUID(id)
}
