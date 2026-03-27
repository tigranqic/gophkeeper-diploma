package storage

import (
	"encoding/json"
	"iter"
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
	for r := range s.All() {
		if r.Id == id {
			return r
		}
	}
	return nil
}

// All returns an iterator over every record in the store.
// Returning iter.Seq keeps callers decoupled from the underlying slice so the
// storage layout can change without touching every call site.
func (s *LocalStore) All() iter.Seq[*pb.EncryptedRecord] {
	return func(yield func(*pb.EncryptedRecord) bool) {
		for _, r := range s.Records {
			if !yield(r) {
				return
			}
		}
	}
}

// AllIndexed returns an iterator over (index, record) pairs.
func (s *LocalStore) AllIndexed() iter.Seq2[int, *pb.EncryptedRecord] {
	return func(yield func(int, *pb.EncryptedRecord) bool) {
		for i, r := range s.Records {
			if !yield(i, r) {
				return
			}
		}
	}
}

// UpsertRecord updates or inserts a record into the local storage.
// A server-returned record wins if its UpdatedAt is newer OR its Revision is higher
// (the revision is only assigned server-side, so a higher revision always means
// a more authoritative version of the record).
func (s *LocalStore) UpsertRecord(rec *pb.EncryptedRecord) {
	// AllIndexed yields (position, record) pairs so we can overwrite
	// the existing slot directly without a second lookup.
	for i, r := range s.AllIndexed() {
		if r.Id == rec.Id {
			if rec.UpdatedAt > r.UpdatedAt || rec.Revision > r.Revision {
				s.Records[i] = rec
			}
			return
		}
	}
	s.Records = append(s.Records, rec)
}
