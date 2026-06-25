package server

import (
	"context"

	__ "github.com/devlikeapro/gows/proto"
	"go.mau.fi/whatsmeow/types"
)

func (s *Server) GetBusinessProfile(ctx context.Context, req *__.JidRequest) (*__.Json, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	profile, err := cli.GetBusinessProfile(ctx, jid)
	if err != nil {
		return nil, err
	}
	return s.toJson(profile), nil
}
