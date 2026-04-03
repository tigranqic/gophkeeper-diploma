package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"os"
	"testing"
)

func TestLocalStore_UpsertAndGet(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "gophkeeper_test_*.json")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	store, err := Load(tmpFile.Name())
	require.NoError(t, err)

	rec := &pb.EncryptedRecord{Id: "1", UpdatedAt: 100}
	store.UpsertRecord(rec)

	found := store.GetRecord("1")
	assert.NotNil(t, found)
	assert.Equal(t, int64(100), found.UpdatedAt)

	// Update with newer
	recNew := &pb.EncryptedRecord{Id: "1", UpdatedAt: 200}
	store.UpsertRecord(recNew)
	assert.Equal(t, int64(200), store.GetRecord("1").UpdatedAt)

	// Update with older (should not change)
	recOld := &pb.EncryptedRecord{Id: "1", UpdatedAt: 150}
	store.UpsertRecord(recOld)
	assert.Equal(t, int64(200), store.GetRecord("1").UpdatedAt)
}

func TestLocalStore_SaveAndLoad(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "gophkeeper_test_*.json")
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	store, err := Load(tmpFile.Name())
	require.NoError(t, err)

	store.Token = "test-token"
	store.Salt = []byte("test-salt")
	store.UpsertRecord(&pb.EncryptedRecord{Id: "1", UpdatedAt: 100})

	err = store.Save()
	require.NoError(t, err)

	// Load back
	store2, err := Load(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, "test-token", store2.Token)
	assert.Equal(t, []byte("test-salt"), store2.Salt)
	assert.Len(t, store2.Records, 1)
}
