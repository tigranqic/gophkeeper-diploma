package gservice

import (
	"github.com/google/uuid"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Sync handles bidirectional synchronization of encrypted records.
// It upserts the client's records on the server, then streams back all records
// updated since the client's last known revision.
func (s *Server) Sync(req *pb.SyncRequest, stream pb.StorageService_SyncServer) error {
	ctx := stream.Context()
	userIDStr, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return status.Error(codes.Unauthenticated, "unauthenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return status.Error(codes.Internal, "invalid user id")
	}

	updates, maxRevision, err := s.repo.SyncRecords(ctx, userID, req.Records, req.LastRevision)
	if err != nil {
		return status.Error(codes.Internal, "failed to sync records")
	}

	// Stream updates in chunks to avoid message size limits
	chunkSize := 100
	for i := 0; i < len(updates); i += chunkSize {
		end := i + chunkSize
		if end > len(updates) {
			end = len(updates)
		}

		err := stream.Send(&pb.SyncResponse{
			Records:         updates[i:end],
			CurrentRevision: maxRevision,
		})
		if err != nil {
			return err
		}
	}

	// If no updates, still send current revision
	if len(updates) == 0 {
		return stream.Send(&pb.SyncResponse{
			CurrentRevision: maxRevision,
		})
	}

	return nil
}
