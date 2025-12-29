package air

import (
	"context"
	"fmt"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// processOfflineQueues processes all pending offline actions.
func (s *Server) processOfflineQueues() {
	for email, queue := range s.offlineQueues {
		s.processOfflineQueue(email, queue)
	}
}

// processOfflineQueue processes a single account's offline queue.
func (s *Server) processOfflineQueue(email string, queue *cache.OfflineQueue) {
	if s.nylasClient == nil || !s.IsOnline() {
		return
	}

	// Get the grant ID for this email
	var grantID string
	grants, err := s.grantStore.ListGrants()
	if err != nil {
		return
	}
	for _, g := range grants {
		if g.Email == email {
			grantID = g.ID
			break
		}
	}
	if grantID == "" {
		return
	}

	ctx := context.Background()

	for {
		action, err := queue.Dequeue()
		if err != nil || action == nil {
			break
		}

		// Process the action
		err = s.processOfflineAction(ctx, grantID, action)
		if err != nil {
			// Mark as failed and re-queue if retries left
			if action.Attempts < 3 {
				_ = queue.MarkFailed(action.ID, err)
			}
		}
	}
}

// processOfflineAction processes a single offline action.
func (s *Server) processOfflineAction(ctx context.Context, grantID string, action *cache.QueuedAction) error {
	switch action.Type {
	case cache.ActionMarkRead, cache.ActionMarkUnread:
		var payload cache.MarkReadPayload
		if err := action.GetActionData(&payload); err != nil {
			return err
		}
		_, err := s.nylasClient.UpdateMessage(ctx, grantID, payload.EmailID, &domain.UpdateMessageRequest{
			Unread: &payload.Unread,
		})
		return err

	case cache.ActionStar, cache.ActionUnstar:
		var payload cache.StarPayload
		if err := action.GetActionData(&payload); err != nil {
			return err
		}
		_, err := s.nylasClient.UpdateMessage(ctx, grantID, payload.EmailID, &domain.UpdateMessageRequest{
			Starred: &payload.Starred,
		})
		return err

	case cache.ActionDelete:
		return s.nylasClient.DeleteMessage(ctx, grantID, action.ResourceID)

	case cache.ActionMove:
		var payload cache.MovePayload
		if err := action.GetActionData(&payload); err != nil {
			return err
		}
		_, err := s.nylasClient.UpdateMessage(ctx, grantID, payload.EmailID, &domain.UpdateMessageRequest{
			Folders: []string{payload.FolderID},
		})
		return err

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}
