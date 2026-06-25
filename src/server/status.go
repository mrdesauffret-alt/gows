package server

import (
	"context"
	"strings"
	"time"

	__ "github.com/devlikeapro/gows/proto"
	"go.mau.fi/whatsmeow/types"
)

func parseStatusMessageID(statusID string) string {
	parts := strings.Split(statusID, "_")
	if len(parts) >= 3 {
		return parts[2]
	}
	return statusID
}

func (s *Server) ReadStatus(ctx context.Context, req *__.ReadStatusRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	contact, err := types.ParseJID(req.GetContactId())
	if err != nil {
		return nil, err
	}
	messageID := parseStatusMessageID(req.GetStatusId())
	err = cli.MarkRead(
		ctx,
		[]types.MessageID{messageID},
		time.Now(),
		types.StatusBroadcastJID,
		contact,
	)
	if err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}
