package server

import (
	"context"

	__ "github.com/devlikeapro/gows/proto"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
)

func (s *Server) AddChatLabel(ctx context.Context, req *__.ChatLabelRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetChatId())
	if err != nil {
		return nil, err
	}
	patch := appstate.BuildLabelChat(jid, req.GetLabelId(), true)
	if err := cli.SendAppState(ctx, patch); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) RemoveChatLabel(ctx context.Context, req *__.ChatLabelRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetChatId())
	if err != nil {
		return nil, err
	}
	patch := appstate.BuildLabelChat(jid, req.GetLabelId(), false)
	if err := cli.SendAppState(ctx, patch); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}
