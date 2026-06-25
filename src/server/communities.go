package server

import (
	"context"

	__ "github.com/devlikeapro/gows/proto"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func (s *Server) toJson(data interface{}) *__.Json {
	return &__.Json{Data: s.safeMarshal(data)}
}

func (s *Server) toJsonList(items []interface{}) *__.JsonList {
	elements := make([]*__.Json, len(items))
	for i, item := range items {
		elements[i] = s.toJson(item)
	}
	return &__.JsonList{Elements: elements}
}

func (s *Server) ListCommunities(ctx context.Context, req *__.Session) (*__.JsonList, error) {
	cli, err := s.Sm.Get(req.GetId())
	if err != nil {
		return nil, err
	}
	groups, err := cli.GetJoinedGroups(ctx)
	if err != nil {
		return nil, err
	}
	var communities []interface{}
	for _, group := range groups {
		if group.IsParent {
			communities = append(communities, group)
		}
	}
	return s.toJsonList(communities), nil
}

func (s *Server) GetCommunity(ctx context.Context, req *__.JidRequest) (*__.Json, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	info, err := cli.GetGroupInfo(ctx, jid)
	if err != nil {
		return nil, err
	}
	return s.toJson(info), nil
}

func (s *Server) CreateCommunity(ctx context.Context, req *__.CreateCommunityRequest) (*__.Json, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	createReq := whatsmeow.ReqCreateGroup{
		Name: req.GetSubject(),
	}
	createReq.IsParent = true
	info, err := cli.CreateGroup(ctx, createReq)
	if err != nil {
		return nil, err
	}
	if description := req.GetDescription(); description != "" {
		if err := cli.SetGroupDescription(ctx, info.JID, description); err != nil {
			return nil, err
		}
		info, err = cli.GetGroupInfo(ctx, info.JID)
		if err != nil {
			return nil, err
		}
	}
	return s.toJson(info), nil
}

func (s *Server) LeaveCommunity(ctx context.Context, req *__.JidRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	if err := cli.LeaveGroup(ctx, jid); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

type communityLinkedGroupsResponse struct {
	CommunityJID string                   `json:"communityJid"`
	IsCommunity  bool                     `json:"isCommunity"`
	LinkedGroups []*types.GroupLinkTarget `json:"linkedGroups"`
}

func (s *Server) GetCommunityLinkedGroups(ctx context.Context, req *__.JidRequest) (*__.Json, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	info, err := cli.GetGroupInfo(ctx, jid)
	if err != nil {
		return nil, err
	}
	linked, err := cli.GetSubGroups(ctx, jid)
	if err != nil {
		return nil, err
	}
	return s.toJson(communityLinkedGroupsResponse{
		CommunityJID: jid.String(),
		IsCommunity:  info.IsParent,
		LinkedGroups: linked,
	}), nil
}

func (s *Server) LinkCommunityGroup(ctx context.Context, req *__.CommunityGroupLinkRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	parent, err := types.ParseJID(req.GetCommunityJid())
	if err != nil {
		return nil, err
	}
	child, err := types.ParseJID(req.GetGroupJid())
	if err != nil {
		return nil, err
	}
	if err := cli.LinkGroup(ctx, parent, child); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) UnlinkCommunityGroup(ctx context.Context, req *__.CommunityGroupLinkRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	parent, err := types.ParseJID(req.GetCommunityJid())
	if err != nil {
		return nil, err
	}
	child, err := types.ParseJID(req.GetGroupJid())
	if err != nil {
		return nil, err
	}
	if err := cli.UnlinkGroup(ctx, parent, child); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) CreateCommunityGroup(ctx context.Context, req *__.CreateCommunityGroupRequest) (*__.Json, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	parent, err := types.ParseJID(req.GetCommunityJid())
	if err != nil {
		return nil, err
	}
	participants := make([]types.JID, len(req.GetParticipants()))
	for i, p := range req.GetParticipants() {
		jid, parseErr := types.ParseJID(p)
		if parseErr != nil {
			return nil, parseErr
		}
		participants[i] = jid
	}
	createReq := whatsmeow.ReqCreateGroup{
		Name:         req.GetSubject(),
		Participants: participants,
	}
	createReq.LinkedParentJID = parent
	info, err := cli.CreateGroup(ctx, createReq)
	if err != nil {
		return nil, err
	}
	return s.toJson(info), nil
}

func (s *Server) SetCommunitySubject(ctx context.Context, req *__.JidStringRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	if err := cli.SetGroupName(ctx, jid, req.GetValue()); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) SetCommunityDescription(ctx context.Context, req *__.JidStringRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	if err := cli.SetGroupDescription(ctx, jid, req.GetValue()); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) GetCommunityInviteLink(ctx context.Context, req *__.JidRequest) (*__.OptionalString, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	link, err := cli.GetGroupInviteLink(ctx, jid, false)
	if err != nil {
		return nil, err
	}
	return &__.OptionalString{Value: link}, nil
}

func (s *Server) JoinCommunity(ctx context.Context, req *__.GroupCodeRequest) (*__.Json, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := cli.JoinGroupWithLink(ctx, req.GetCode())
	if err != nil {
		return nil, err
	}
	return s.toJson(map[string]string{"jid": jid.String()}), nil
}
