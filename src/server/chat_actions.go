package server

import (
	"context"
	"errors"
	"time"

	__ "github.com/devlikeapro/gows/proto"
	"github.com/golang/protobuf/proto"
	"go.mau.fi/whatsmeow/appstate"
	waCommon "go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	waSyncAction "go.mau.fi/whatsmeow/proto/waSyncAction"
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
		if remaining <= 0 {
			patch := appstate.BuildMute(jid, false, 0)
			if err := cli.SendAppState(ctx, patch); err != nil {
				return nil, err
			}
			return &__.Empty{}, nil
		}
		muteDuration = remaining
	case req.GetDurationSeconds() != nil:
		sec := req.GetDurationSeconds().GetValue()
		muteDuration = time.Duration(sec) * time.Second
	default:
		muteDuration = 8 * time.Hour
	}

	patch := appstate.BuildMute(jid, true, muteDuration)
	if err := cli.SendAppState(ctx, patch); err != nil {
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
	if err := cli.SendAppState(ctx, patch); err != nil {
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
	_, err = cli.UpdateBlocklist(ctx, jid, action)
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
	list, err := cli.GetBlocklist(ctx)
	if err != nil {
		return nil, err
	}
	jids := make([]string, 0, len(list.JIDs))
	for _, j := range list.JIDs {
		jids = append(jids, j.String())
	}
	return &__.BlocklistResponse{Jids: jids}, nil
}

func syncActionMessageRange(lastMessageTimestamp time.Time, lastMessageKey *waCommon.MessageKey) *waSyncAction.SyncActionMessageRange {
	if lastMessageTimestamp.IsZero() {
		lastMessageTimestamp = time.Now()
	}
	messageRange := &waSyncAction.SyncActionMessageRange{
		LastMessageTimestamp: proto.Int64(lastMessageTimestamp.Unix()),
	}
	if lastMessageKey != nil {
		messageRange.Messages = []*waSyncAction.SyncActionMessage{{
			Key:       lastMessageKey,
			Timestamp: proto.Int64(lastMessageTimestamp.Unix()),
		}}
	}
	return messageRange
}

func buildClearChat(target types.JID, lastMessageTimestamp time.Time, lastMessageKey *waCommon.MessageKey) appstate.PatchInfo {
	action := &waSyncAction.ClearChatAction{
		MessageRange: syncActionMessageRange(lastMessageTimestamp, lastMessageKey),
	}
	return appstate.PatchInfo{
		Type: appstate.WAPatchRegularHigh,
		Mutations: []appstate.MutationInfo{{
			Index:   []string{appstate.IndexClearChat, target.String()},
			Version: 6,
			Value: &waSyncAction.SyncActionValue{
				ClearChatAction: action,
			},
		}},
	}
}

func (s *Server) DeleteChat(ctx context.Context, req *__.JidRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	patch := appstate.BuildDeleteChat(jid, time.Now(), nil, false)
	if err := cli.SendAppState(ctx, patch); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) ClearChat(ctx context.Context, req *__.JidRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	jid, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	patch := buildClearChat(jid, time.Now(), nil)
	if err := cli.SendAppState(ctx, patch); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) StarMessage(ctx context.Context, req *__.StarMessageRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	target, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	sender, err := types.ParseJID(req.GetSender())
	if err != nil {
		return nil, err
	}
	fromMe := false
	if cli.Store.ID != nil && sender.User == cli.Store.ID.User {
		fromMe = true
	} else if sender.User == cli.Store.LID.User && cli.Store.LID.User != "" {
		fromMe = true
	}
	patch := appstate.BuildStar(target, sender, req.GetMessageId(), fromMe, req.GetStar())
	if err := cli.SendAppState(ctx, patch); err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func buildPinInChatMessage(
	cli interface {
		BuildMessageKey(chat, sender types.JID, id types.MessageID) *waCommon.MessageKey
	},
	chat types.JID,
	sender types.JID,
	messageID types.MessageID,
	durationSeconds int64,
	pin bool,
) *waE2E.Message {
	pinType := waE2E.PinInChatMessage_PIN_FOR_ALL
	if !pin {
		pinType = waE2E.PinInChatMessage_UNPIN_FOR_ALL
	}
	message := &waE2E.Message{
		PinInChatMessage: &waE2E.PinInChatMessage{
			Key:               cli.BuildMessageKey(chat, sender, messageID),
			Type:              pinType.Enum(),
			SenderTimestampMS: proto.Int64(time.Now().UnixMilli()),
		},
	}
	if pin && durationSeconds > 0 {
		message.MessageContextInfo = &waE2E.MessageContextInfo{
			MessageAddOnDurationInSecs: proto.Uint32(uint32(durationSeconds)),
			MessageAddOnExpiryType:     waE2E.MessageContextInfo_STATIC.Enum(),
		}
	}
	return message
}

func (s *Server) PinMessage(ctx context.Context, req *__.PinMessageRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	chat, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	sender, err := types.ParseJID(req.GetSender())
	if err != nil {
		return nil, err
	}
	message := buildPinInChatMessage(cli, chat, sender, req.GetMessageId(), req.GetDurationSeconds(), true)
	_, err = cli.SendMessage(ctx, chat, message)
	if err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}

func (s *Server) UnpinMessage(ctx context.Context, req *__.UnpinMessageRequest) (*__.Empty, error) {
	cli, err := s.Sm.Get(req.GetSession().GetId())
	if err != nil {
		return nil, err
	}
	chat, err := types.ParseJID(req.GetJid())
	if err != nil {
		return nil, err
	}
	sender, err := types.ParseJID(req.GetSender())
	if err != nil {
		return nil, err
	}
	message := buildPinInChatMessage(cli, chat, sender, req.GetMessageId(), 0, false)
	_, err = cli.SendMessage(ctx, chat, message)
	if err != nil {
		return nil, err
	}
	return &__.Empty{}, nil
}
