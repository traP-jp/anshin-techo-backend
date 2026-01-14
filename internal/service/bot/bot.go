package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/traP-jp/anshin-techo-backend/internal/repository"
	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

type Config struct {
	Origin      string
	AccessToken string

	TicketCreateChannelID string // チケット作成通知
	TicketUpdateChannelID string // チケット編集通知
	NoteIncomingChannelID string // 受信ノート通知
	NoteOutgoingChannelID string // 発信ノート通知
	NoteOtherChannelID    string // その他ノート通知
	ReviewNotifyChannelID string // レビュー通知

	// ユーザー設定
	ManagerID string // 本職（マネージャー）のTraQ ID
}

type Service struct {
	bot    *traqwsbot.Bot
	config Config

	// 内部キャッシュ
	yokunasasouStampID string
}

var (
	_ Client        = (*Service)(nil)
	_ MessageSender = (*Service)(nil)
	_ EventHandler  = (*Service)(nil)
)

func NewService(cfg Config) (*Service, error) {
	if cfg.Origin == "" || cfg.AccessToken == "" {
		return nil, fmt.Errorf("bot config is incomplete: origin and access token are required")
	}

	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		Origin:               cfg.Origin,
		AccessToken:          cfg.AccessToken,
		DisableAutoReconnect: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	s := &Service{
		bot:                bot,
		config:             cfg,
		yokunasasouStampID: "",
	}

	s.setupInternalHandlers()

	return s, nil
}

func (s *Service) Start() error {

	go func() {
		ctx := context.Background()
		if err := s.fetchYokunasasouStampID(ctx); err != nil {
			fmt.Printf("failed to fetch stamp id: %v\n", err)
		}
	}()

	return s.bot.Start()
}

func (s *Service) API() *traq.APIClient {
	return s.bot.API()
}

func (s *Service) setupInternalHandlers() {
	s.bot.OnBotMessageStampsUpdated(func(p *payload.BotMessageStampsUpdated) {
		for _, stamp := range p.Stamps {
			if s.yokunasasouStampID != "" && stamp.StampID == s.yokunasasouStampID {
				_, _ = s.bot.API().MessageAPI.DeleteMessage(context.Background(), p.MessageID).Execute()

				return
			}
		}
	})
}

func (s *Service) fetchYokunasasouStampID(ctx context.Context) error {
	stamps, _, err := s.bot.API().StampAPI.GetStamps(ctx).Execute()
	if err != nil {
		return err
	}
	for _, stamp := range stamps {
		if stamp.Name == "yokunasasou" {
			s.yokunasasouStampID = stamp.Id

			return nil
		}
	}

	return fmt.Errorf("stamp :yokunasasou: not found")
}

func (s *Service) generateMention(ctx context.Context, userID string) string {
	if userID == "" {
		return ""
	}
	user, _, err := s.bot.API().UserAPI.GetUser(ctx, userID).Execute()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("@%s", user.Name)
}

func (s *Service) PostMessage(ctx context.Context, channelID string, content string) error {
	embedTrue := true
	_, _, err := s.bot.API().MessageAPI.
		PostMessage(ctx, channelID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: content,
			Embed:   &embedTrue,
			Nonce:   nil,
		}).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}

	return nil
}

func (s *Service) PostDirectMessage(ctx context.Context, userID string, content string) error {
	embedTrue := true
	_, _, err := s.bot.API().UserAPI.
		PostDirectMessage(ctx, userID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: content,
			Embed:   &embedTrue,
			Nonce:   nil,
		}).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to post direct message: %w", err)
	}

	return nil
}

// NotifyTicketCreated : チケット作成通知
func (s *Service) NotifyTicketCreated(ctx context.Context, ticket *repository.Ticket) error {
	dueStr := "未設定"
	if ticket.Due.Valid {
		dueStr = ticket.Due.Time.Format("2006/01/02")
	}

	message := fmt.Sprintf(`## 新しいチケット(ID: %d)が作成されました
タイトル: %s
担当者: @%s
副担当: %v
関係者: %v
タグ: %v
締め切り: %s

%s`,
		ticket.ID,
		ticket.Title,
		ticket.Assignee,
		ticket.SubAssignees,
		ticket.Stakeholders,
		ticket.Tags,
		dueStr,
		ticket.Description.String,
	)

	targetID := s.config.TicketCreateChannelID
	return s.PostMessage(ctx, targetID, message)
}

// NotifyTicketUpdated : チケット編集通知
func (s *Service) NotifyTicketUpdated(ctx context.Context, ticket *repository.Ticket) error {
	targets := []string{ticket.Assignee}
	targets = append(targets, ticket.SubAssignees...)
	targets = append(targets, ticket.Stakeholders...)

	uniqueTargets := make(map[string]bool)
	var mentions []string
	for _, name := range targets {
		if name != "" && !uniqueTargets[name] {
			uniqueTargets[name] = true
			mentions = append(mentions, "@"+name)
		}
	}
	mentionsStr := strings.Join(mentions, " ")

	dueStr := "未設定"
	if ticket.Due.Valid {
		dueStr = ticket.Due.Time.Format("2006/01/02")
	}

	content := fmt.Sprintf(`### チケット更新(ID: %d)
%s
タイトル: %s
状態: %s
期限: %s
担当: @%s

内容が更新されました。詳細は確認してください。`,
		ticket.ID,
		mentionsStr,
		ticket.Title,
		ticket.Status,
		dueStr,
		ticket.Assignee,
	)

	return s.PostMessage(ctx, s.config.TicketUpdateChannelID, content)
}

// NotifyNoteCreated : ノート作成通知
func (s *Service) NotifyNoteCreated(ctx context.Context, noteType string, contentPreview string, authorID string, shouldMention bool) error {
	var channelID string
	var typeLabel string

	switch noteType {
	case "incoming":
		channelID = s.config.NoteIncomingChannelID
		typeLabel = "受信"
	case "outgoing":
		channelID = s.config.NoteOutgoingChannelID
		typeLabel = "発信"
	default:
		channelID = s.config.NoteOtherChannelID
		typeLabel = "その他"
	}

	authorName := ""
	if shouldMention {
		authorName = s.generateMention(ctx, authorID)
	} else {
		user, _, err := s.bot.API().UserAPI.GetUser(ctx, authorID).Execute()
		if err == nil {
			authorName = user.Name
		} else {
			authorName = "不明なユーザー"
		}
	}

	msg := fmt.Sprintf("＃＃＃ ノート作成 (%s)\n作成者: %s\n\n%s", typeLabel, authorName, contentPreview)

	return s.PostMessage(ctx, channelID, msg)
}

// NotifyReviewCreated : レビュー通知
func (s *Service) NotifyReviewCreated(ctx context.Context, noteTitle string, noteAuthorID string, reviewerID string, comment string) error {
	reviewerMention := s.generateMention(ctx, reviewerID)

	targetMention := ""
	if comment != "" {
		targetMention = s.generateMention(ctx, noteAuthorID)
	} else {
		user, _, err := s.bot.API().UserAPI.GetUser(ctx, noteAuthorID).Execute()
		if err == nil {
			targetMention = user.Name
		}
	}

	msg := fmt.Sprintf("### レビュー通知\n案件: %s\nレビュワー: %s -> %s\n\n%s", noteTitle, reviewerMention, targetMention, comment)

	return s.PostMessage(ctx, s.config.ReviewNotifyChannelID, msg)
}

// SendDeadlineReminder : 期限超過リマインダー (担当者と本職にDM)
func (s *Service) SendDeadlineReminder(ctx context.Context, ticket *repository.Ticket, daysOverdue int) error {
	dueStr := "未設定"
	if ticket.Due.Valid {
		dueStr = ticket.Due.Time.Format("2006/01/02")
	}

	msg := fmt.Sprintf(` ###【期限超過リマインド】 
@%s
案件「**%s**」(ID: %d) の期限(%s)から **%d日** が経過しました。
至急、対応状況を確認・報告してください。

**状態**: %s`,
		ticket.Assignee,
		ticket.Title,
		ticket.ID,
		dueStr,
		daysOverdue,
		ticket.Status,
	)

	if s.config.ManagerID != "" {
		msg += fmt.Sprintf(" @%s", s.config.ManagerID)
	}

	return s.PostMessage(ctx, s.config.TicketCreateChannelID, msg)
}

// SendWaitingSentReminder : 送信待ちリマインダー
// targetUserID: 送信相手のユーザーID
func (s *Service) SendWaitingSentReminder(ctx context.Context, ticket *repository.Ticket) error {
	msg := fmt.Sprintf(`##【送信待ちリマインド】
@%s
案件「**%s**」(ID: %d) が「送信待ち」ステータスになっています。
送信作業をお願いします。

**期限**: %s`,
		ticket.Assignee,
		ticket.Title,
		ticket.ID,
		func() string {
			if ticket.Due.Valid {
				return ticket.Due.Time.Format("2006/01/02")
			}
			return "未設定"
		}(),
	)

	return s.PostMessage(ctx, s.config.TicketCreateChannelID, msg)
}

func AddBusinessHours(start time.Time, duration time.Duration) time.Time {
	current := start
	remaining := duration

	for remaining > 0 {
		h := current.Hour()
		if h >= 0 && h < 8 {
			nextStart := time.Date(current.Year(), current.Month(), current.Day(), 8, 0, 0, 0, current.Location())
			current = nextStart

			continue
		}

		tomorrow := current.AddDate(0, 0, 1)
		nextStop := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, current.Location())

		canWork := nextStop.Sub(current)

		if remaining <= canWork {
			return current.Add(remaining)
		}

		current = nextStop
		remaining -= canWork
	}

	return current
}

func (s *Service) OnMessageCreated(handler func(messageID, channelID, userID, content string)) {
	s.bot.OnMessageCreated(func(p *payload.MessageCreated) {
		handler(p.Message.ID, p.Message.ChannelID, p.Message.User.ID, p.Message.Text)
	})
}

func (s *Service) OnBotMessageStampsUpdated(handler func(messageID string, stamps []payload.MessageStamp)) {
	s.bot.OnBotMessageStampsUpdated(func(p *payload.BotMessageStampsUpdated) {
		handler(p.MessageID, p.Stamps)
	})
}

func (s *Service) Config() Config {
	return s.config
}

func (s *Service) GetUserIDByName(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("username is empty")
	}

	users, _, err := s.bot.API().UserAPI.GetUsers(ctx).Name(name).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to search user '%s': %w", name, err)
	}

	if len(users) == 0 {
		return "", fmt.Errorf("user '%s' not found", name)
	}

	return users[0].Id, nil
}
