package storage

import (
	"encoding/json"
	"os"
	"strings"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
)

// LocalStore represents the client-side local cache.
type LocalStore struct {
	path         string
	Records      []*pb.EncryptedRecord `json:"records"`
	Token        string                `json:"token"`
	Salt         []byte                `json:"salt"`
	LastRevision int64                 `json:"last_revision"`
}

// Load loads the local storage from the given path.
func Load(path string) (*LocalStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &LocalStore{path: path}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return &LocalStore{path: path}, nil
	}
	var store LocalStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	store.path = path
	return &store, nil
}

// Save saves the local storage to the file.
func (s *LocalStore) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600)
}

// GetRecord returns a record by ID.
func (s *LocalStore) GetRecord(id string) *pb.EncryptedRecord {
	for _, r := range s.Records {
		if r.Id == id {
			return r
		}
	}
	return nil
}

// UpsertRecord updates or inserts a record into the local storage.
func (s *LocalStore) UpsertRecord(rec *pb.EncryptedRecord) {
	for i, r := range s.Records {
		if r.Id == rec.Id {
			if rec.UpdatedAt > r.UpdatedAt {
				s.Records[i] = rec
			}
			return
		}
	}
	s.Records = append(s.Records, rec)
}
