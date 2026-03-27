package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/app"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/models"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/pkg/crypto"
)

// AddCmd returns the parent 'add' command with data-type sub-commands.
func AddCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new data",
	}

	cmd.AddCommand(AddLoginCmd(a))
	cmd.AddCommand(AddTextCmd(a))
	cmd.AddCommand(AddBinaryCmd(a))
	cmd.AddCommand(AddCardCmd(a))
	return cmd
}

// AddLoginCmd returns the command for storing a login/password pair.
func AddLoginCmd(a *app.App) *cobra.Command {
	var login, pass, meta string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Add login/password pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.EnsureDEK(); err != nil {
				return err
			}

			payload := models.LoginPayload{
				Login:    login,
				Password: pass,
				Metadata: meta,
			}

			data, err := models.MarshalPayload(payload)
			if err != nil {
				return err
			}

			encrypted, err := crypto.Encrypt(data, a.DEK)
			if err != nil {
				return err
			}

			rec := &pb.EncryptedRecord{
				Id:        uuid.New().String(),
				Type:      pb.RecordType_RECORD_TYPE_LOGIN,
				Blob:      encrypted,
				Revision:  0,
				UpdatedAt: time.Now().UnixMilli(),
			}

			a.Storage.UpsertRecord(rec)
			return a.Storage.Save()
		},
	}
	cmd.Flags().StringVar(&login, "login", "", "login")
	cmd.Flags().StringVar(&pass, "pass", "", "password")
	cmd.Flags().StringVar(&meta, "meta", "", "metadata")
	if err := cmd.MarkFlagRequired("login"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("pass"); err != nil {
		panic(err)
	}
	return cmd
}

// AddTextCmd returns the command for storing arbitrary text data.
func AddTextCmd(a *app.App) *cobra.Command {
	var text, meta string
	cmd := &cobra.Command{
		Use:   "text",
		Short: "Add arbitrary text data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.EnsureDEK(); err != nil {
				return err
			}

			payload := models.TextPayload{
				Text:     text,
				Metadata: meta,
			}

			data, err := models.MarshalPayload(payload)
			if err != nil {
				return err
			}

			encrypted, err := crypto.Encrypt(data, a.DEK)
			if err != nil {
				return err
			}

			rec := &pb.EncryptedRecord{
				Id:        uuid.New().String(),
				Type:      pb.RecordType_RECORD_TYPE_TEXT,
				Blob:      encrypted,
				Revision:  0,
				UpdatedAt: time.Now().UnixMilli(),
			}

			a.Storage.UpsertRecord(rec)
			return a.Storage.Save()
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "text content")
	cmd.Flags().StringVar(&meta, "meta", "", "metadata")
	if err := cmd.MarkFlagRequired("text"); err != nil {
		panic(err)
	}
	return cmd
}

// AddBinaryCmd returns the command for storing arbitrary binary data encoded as hex.
func AddBinaryCmd(a *app.App) *cobra.Command {
	var filePath, meta string
	cmd := &cobra.Command{
		Use:   "binary",
		Short: "Add binary data from a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.EnsureDEK(); err != nil {
				return err
			}

			fileData, err := readFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			payload := models.BinaryPayload{
				Data:     fileData,
				Metadata: meta,
			}

			data, err := models.MarshalPayload(payload)
			if err != nil {
				return err
			}

			encrypted, err := crypto.Encrypt(data, a.DEK)
			if err != nil {
				return err
			}

			rec := &pb.EncryptedRecord{
				Id:        uuid.New().String(),
				Type:      pb.RecordType_RECORD_TYPE_BINARY,
				Blob:      encrypted,
				Revision:  0,
				UpdatedAt: time.Now().UnixMilli(),
			}

			a.Storage.UpsertRecord(rec)
			return a.Storage.Save()
		},
	}
	cmd.Flags().StringVar(&filePath, "file", "", "path to file")
	cmd.Flags().StringVar(&meta, "meta", "", "metadata")
	if err := cmd.MarkFlagRequired("file"); err != nil {
		panic(err)
	}
	return cmd
}

// AddCardCmd returns the command for storing bank card details.
func AddCardCmd(a *app.App) *cobra.Command {
	var number, expiry, cvc, holder, meta string
	cmd := &cobra.Command{
		Use:   "card",
		Short: "Add bank card data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.EnsureDEK(); err != nil {
				return err
			}

			payload := models.CardPayload{
				Number:   number,
				Expiry:   expiry,
				CVC:      cvc,
				Holder:   holder,
				Metadata: meta,
			}

			data, err := models.MarshalPayload(payload)
			if err != nil {
				return err
			}

			encrypted, err := crypto.Encrypt(data, a.DEK)
			if err != nil {
				return err
			}

			rec := &pb.EncryptedRecord{
				Id:        uuid.New().String(),
				Type:      pb.RecordType_RECORD_TYPE_CARD,
				Blob:      encrypted,
				Revision:  0,
				UpdatedAt: time.Now().UnixMilli(),
			}

			a.Storage.UpsertRecord(rec)
			return a.Storage.Save()
		},
	}
	cmd.Flags().StringVar(&number, "number", "", "card number")
	cmd.Flags().StringVar(&expiry, "expiry", "", "expiry date (MM/YY)")
	cmd.Flags().StringVar(&cvc, "cvc", "", "CVC code")
	cmd.Flags().StringVar(&holder, "holder", "", "card holder name")
	cmd.Flags().StringVar(&meta, "meta", "", "metadata")
	if err := cmd.MarkFlagRequired("number"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("expiry"); err != nil {
		panic(err)
	}
	return cmd
}

// ListCmd returns the command for listing all locally cached records.
func ListCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all records",
		Run: func(cmd *cobra.Command, args []string) {
			for r := range a.Storage.All() {
				if r.IsDeleted {
					continue
				}
				fmt.Printf("ID: %s | Type: %s | Revision: %d | Updated: %s\n",
					r.Id, r.Type, r.Revision, time.UnixMilli(r.UpdatedAt).Format(time.RFC3339))
			}
		},
	}
}

// GetCmd returns the command for retrieving and decrypting a record by ID.
func GetCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get and decrypt a record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.EnsureDEK(); err != nil {
				return err
			}

			rec := a.Storage.GetRecord(args[0])
			if rec == nil {
				return fmt.Errorf("record not found")
			}

			decrypted, err := crypto.Decrypt(rec.Blob, a.DEK)
			if err != nil {
				return err
			}

			fmt.Printf("Decrypted data: %s\n", string(decrypted))
			return nil
		},
	}
}

// DeleteCmd returns the command for marking a record as deleted locally.
// The deletion is propagated to the server on the next sync.
func DeleteCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "delete [id]",
		Short: "Mark a record as deleted (synced on next sync)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rec := a.Storage.GetRecord(args[0])
			if rec == nil {
				return fmt.Errorf("record not found")
			}

			rec.IsDeleted = true
			rec.UpdatedAt = time.Now().UnixMilli()
			a.Storage.UpsertRecord(rec)
			return a.Storage.Save()
		},
	}
}

// SyncCmd returns the command for synchronizing local records with the server.
func SyncCmd(a *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync records with server",
		RunE: func(cmd *cobra.Command, args []string) error {
			updates, lastRev, err := a.Client.Sync(cmd.Context(), slices.Collect(a.Storage.All()), a.Storage.LastRevision)
			if err != nil {
				return err
			}

			for _, rec := range updates {
				a.Storage.UpsertRecord(rec)
			}
			a.Storage.LastRevision = lastRev
			return a.Storage.Save()
		},
	}
}

// readFile reads the file at path using os.Root to prevent symlink/TOCTOU attacks.
func readFile(path string) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(filepath.Dir(absPath))
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	f, err := root.Open(filepath.Base(absPath))
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return io.ReadAll(f)
}
