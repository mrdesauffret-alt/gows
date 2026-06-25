package server

import (
	"context"
	"errors"
	"time"

	__ "github.com/devlikeapro/gows/proto"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func (s *Server) MuteChat(ctx context.Context, req *__.ChatMuteRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}

	var muteDuration time.Duration
	switch {
	case req.GetMuteEndMs() != nil:
		muteEndMs := req.GetMuteEndMs().GetValue()
		if muteEndMs > 0 && muteEndMs < 1_000_000_000_000 {
			muteEndMs *= 1000
		}
		remaining := time.Duration(muteEndMs-time.Now().UnixMilli()) * time.Millisecond
		if remaining < 0 {
			remaining = 0
		}
		muteDuration = remaining
	case req.GetDurationSeconds() != nil:
		sec := req.GetDurationSeconds().GetValue()
		muteDuration = time.Duration(sec) * time.Second
	default:
		muteDuration = 8 * time.Hour
	}

	patch := appstate.BuildMute(jid, true, muteDuration)
	if err := cli.SendAppState(patch); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) UnmuteChat(ctx context.Context, req *__.JidRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	patch := appstate.BuildMute(jid, false, 0)
	if err := cli.SendAppState(patch); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) UpdateBlockStatus(ctx context.Context, req *__.UpdateBlockStatusRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	var action events.BlocklistChangeAction
	switch req.GetAction() {
	case __.BlockAction_BLOCK:
		action = events.BlocklistChangeActionBlock
	case __.BlockAction_UNBLOCK:
		action = events.BlocklistChangeActionUnblock
	default:
		return nil, errors.New("invalid block action")
	}
	_, err = cli.UpdateBlocklist(jid, action)
	if err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) GetBlocklist(ctx context.Context, req *__.Session) (*__.BlocklistResponse, error) {
	cli, err := s.Sm.Get(req.GetId())
	if err != nil {
		return nil, err
	}
	list, err := cli.GetBlocklist()
	if err != nil {
		return nil, err
	}
	jids := make([]string, 0, len(list.JIDs))
	for _, j := range list.JIDs {
		jids = append(jids, j.String())
	}
	return &__.BlocklistResponse{Jids: jids}, nil
}
