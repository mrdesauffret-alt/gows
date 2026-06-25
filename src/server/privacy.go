package server

import (
	"context"

	__ "github.com/devlikeapro/gows/proto"
	"github.com/golang/protobuf/proto"
	"go.mau.fi/whatsmeow/appstate"
	waSyncAction "go.mau.fi/whatsmeow/proto/waSyncAction"
	"go.mau.fi/whatsmeow/types"
)

func privacySettingsToResponse(settings *types.PrivacySettings, linkPreviewsDisabled bool) *__.PrivacySettingsResponse {
	categories := map[string]string{}
	if settings != nil {
		if settings.LastSeen != "" {
			categories["last"] = string(settings.LastSeen)
		}
		if settings.Profile != "" {
			categories["profile"] = string(settings.Profile)
		}
		if settings.Status != "" {
			categories["status"] = string(settings.Status)
		}
		if settings.ReadReceipts != "" {
			categories["readreceipts"] = string(settings.ReadReceipts)
		}
		if settings.GroupAdd != "" {
			categories["groupadd"] = string(settings.GroupAdd)
		}
		if settings.Online != "" {
			categories["online"] = string(settings.Online)
		}
		if settings.CallAdd != "" {
			categories["calladd"] = string(settings.CallAdd)
		}
		if settings.Messages != "" {
			categories["messages"] = string(settings.Messages)
		}
	}
	return &__.PrivacySettingsResponse{
		Categories:           categories,
		LinkPreviewsDisabled: linkPreviewsDisabled,
	}
}

func buildDisableLinkPreviews(disabled bool) appstate.PatchInfo {
	return appstate.PatchInfo{
		Type: appstate.WAPatchCriticalBlock,
		Mutations: []appstate.MutationInfo{{
			Index:   []string{appstate.IndexSettingDisableLinkPreviews},
			Version: 7,
			Value: &waSyncAction.SyncActionValue{
				PrivacySettingDisableLinkPreviewsAction: &waSyncAction.PrivacySettingDisableLinkPreviewsAction{
					IsPreviewsDisabled: proto.Bool(disabled),
				},
			},
		}},
	}
}

func (s *Server) GetPrivacySettings(ctx context.Context, req *__.Session) (*__.PrivacySettingsResponse, error) {
	cli, err := s.Sm.Get(req.GetId())
	if err != nil {
		return nil, err
	}
	settings, err := cli.TryFetchPrivacySettings(ctx, true)
	if err != nil {
		return nil, err
	}
	return privacySettingsToResponse(settings, false), nil
}

func (s *Server) UpdatePrivacySettings(ctx context.Context, req *__.UpdatePrivacySettingsRequest) (*__.PrivacySettingsResponse, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}

	linkPreviewsDisabled := false
	if req.LinkPreviewsDisabled != nil {
		linkPreviewsDisabled = req.LinkPreviewsDisabled.GetValue()
		patch := buildDisableLinkPreviews(linkPreviewsDisabled)
		if err := cli.SendAppState(ctx, patch); err != nil {
			return nil, err
		}
	}

	updates := []struct {
		field *__.OptionalString
		name  types.PrivacySettingType
	}{
		{req.LastSeen, types.PrivacySettingTypeLastSeen},
		{req.ProfilePicture, types.PrivacySettingTypeProfile},
		{req.Status, types.PrivacySettingTypeStatus},
		{req.ReadReceipts, types.PrivacySettingTypeReadReceipts},
		{req.GroupsAdd, types.PrivacySettingTypeGroupAdd},
		{req.Online, types.PrivacySettingTypeOnline},
		{req.Calls, types.PrivacySettingTypeCallAdd},
		{req.Messages, types.PrivacySettingTypeMessages},
	}

	for _, update := range updates {
		if update.field == nil {
			continue
		}
		_, err = cli.SetPrivacySetting(ctx, update.name, types.PrivacySetting(update.field.GetValue()))
		if err != nil {
			return nil, err
		}
	}

	settings, err := cli.TryFetchPrivacySettings(ctx, true)
	if err != nil {
		return nil, err
	}
	return privacySettingsToResponse(settings, linkPreviewsDisabled), nil
}
