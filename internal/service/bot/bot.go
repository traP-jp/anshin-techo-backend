package bot

import (
	"context"
	"fmt"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
)

type Config struct {
	Origin      string
	AccessToken string
}

type Service struct {
	bot *traqwsbot.Bot
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

	return &Service{bot: bot}, nil
}

func (s *Service) Start() error {
	return s.bot.Start()
}

func (s *Service) API() *traq.APIClient {
	return s.bot.API()
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
	dm, _, err := s.bot.API().UserAPI.
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
	_ = dm

	return nil
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
