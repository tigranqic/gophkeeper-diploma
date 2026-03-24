package commands

import (
	"context"
	"os"
	"testing"

	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/app"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAppWithDEK(t *testing.T) *app.App {
	t.Helper()
	tmp, err := os.CreateTemp("", "gophkeeper_data_test_*.json")
	require.NoError(t, err)
	t.Cleanup(func() { os.Remove(tmp.Name()) })

	store, err := storage.Load(tmp.Name())
	require.NoError(t, err)

	client := &mockAPIClient{
		syncFn: func(_ context.Context, records []*pb.EncryptedRecord, _ int64) ([]*pb.EncryptedRecord, int64, error) {
			return records, 1, nil
		},
	}

	// Use a valid 32-byte DEK for AES-256.
	dek := make([]byte, 32)
	for i := range dek {
		dek[i] = byte(i + 1)
	}

	return &app.App{
		Client:  client,
		Storage: store,
		DEK:     dek,
		ReadPassword: func(prompt string) ([]byte, error) {
			return []byte("pass"), nil
		},
	}
}

func TestListCmd_Empty(t *testing.T) {
	a := newTestAppWithDEK(t)
	cmd := ListCmd(a)
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestAddLoginCmd_Success(t *testing.T) {
	a := newTestAppWithDEK(t)
	cmd := AddLoginCmd(a)
	cmd.SetArgs([]string{"--login", "user@example.com", "--pass", "secret"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Len(t, a.Storage.Records, 1)
	assert.Equal(t, pb.RecordType_RECORD_TYPE_LOGIN, a.Storage.Records[0].Type)
}

func TestAddTextCmd_Success(t *testing.T) {
	a := newTestAppWithDEK(t)
	cmd := AddTextCmd(a)
	cmd.SetArgs([]string{"--text", "my secret note"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Len(t, a.Storage.Records, 1)
	assert.Equal(t, pb.RecordType_RECORD_TYPE_TEXT, a.Storage.Records[0].Type)
}

func TestAddCardCmd_Success(t *testing.T) {
	a := newTestAppWithDEK(t)
	cmd := AddCardCmd(a)
	cmd.SetArgs([]string{"--number", "4111111111111111", "--expiry", "12/26"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Len(t, a.Storage.Records, 1)
	assert.Equal(t, pb.RecordType_RECORD_TYPE_CARD, a.Storage.Records[0].Type)
}

func TestAddBinaryCmd_Success(t *testing.T) {
	a := newTestAppWithDEK(t)

	tmpFile, err := os.CreateTemp("", "bin_test_*.bin")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte{0x01, 0x02, 0x03})
	tmpFile.Close()

	cmd := AddBinaryCmd(a)
	cmd.SetArgs([]string{"--file", tmpFile.Name()})
	err = cmd.Execute()
	require.NoError(t, err)
	assert.Len(t, a.Storage.Records, 1)
	assert.Equal(t, pb.RecordType_RECORD_TYPE_BINARY, a.Storage.Records[0].Type)
}

func TestAddBinaryCmd_MissingFile(t *testing.T) {
	a := newTestAppWithDEK(t)
	cmd := AddBinaryCmd(a)
	cmd.SetArgs([]string{"--file", "/nonexistent/file.bin"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestAddLoginCmd_NoDEK(t *testing.T) {
	a := newTestAppWithDEK(t)
	a.DEK = nil
	cmd := AddLoginCmd(a)
	cmd.SetArgs([]string{"--login", "user", "--pass", "pass"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_Success(t *testing.T) {
	a := newTestAppWithDEK(t)

	// Add a record first.
	addCmd := AddLoginCmd(a)
	addCmd.SetArgs([]string{"--login", "user", "--pass", "secret"})
	require.NoError(t, addCmd.Execute())
	require.Len(t, a.Storage.Records, 1)
	id := a.Storage.Records[0].Id

	getCmd := GetCmd(a)
	getCmd.SetArgs([]string{id})
	err := getCmd.Execute()
	require.NoError(t, err)
}

func TestGetCmd_NotFound(t *testing.T) {
	a := newTestAppWithDEK(t)
	cmd := GetCmd(a)
	cmd.SetArgs([]string{"nonexistent-id"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestGetCmd_NoDEK(t *testing.T) {
	a := newTestAppWithDEK(t)
	a.DEK = nil
	cmd := GetCmd(a)
	cmd.SetArgs([]string{"some-id"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCmd_Success(t *testing.T) {
	a := newTestAppWithDEK(t)

	addCmd := AddLoginCmd(a)
	addCmd.SetArgs([]string{"--login", "user", "--pass", "secret"})
	require.NoError(t, addCmd.Execute())
	id := a.Storage.Records[0].Id

	delCmd := DeleteCmd(a)
	delCmd.SetArgs([]string{id})
	err := delCmd.Execute()
	require.NoError(t, err)
	assert.True(t, a.Storage.Records[0].IsDeleted)
}

func TestDeleteCmd_NotFound(t *testing.T) {
	a := newTestAppWithDEK(t)
	cmd := DeleteCmd(a)
	cmd.SetArgs([]string{"nonexistent-id"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestSyncCmd_Success(t *testing.T) {
	a := newTestAppWithDEK(t)
	cmd := SyncCmd(a)
	err := cmd.Execute()
	require.NoError(t, err)
}
