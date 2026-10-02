package bot

import (
	"context"

	"github.com/traP-jp/anshin-techo-backend/internal/repository"
	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/traq-ws-bot/payload"
)

// MockService はテスト用の Bot Service モック
type MockService struct {
	StartFunc             func() error
	APIFunc               func() *traq.APIClient
	PostMessageFunc       func(ctx context.Context, channelID string, content string) error
	PostDirectMessageFunc func(ctx context.Context, userID string, content string) error

	NotifyTicketCreatedFunc     func(ctx context.Context, ticket *repository.Ticket) error
	NotifyTicketUpdatedFunc     func(ctx context.Context, ticket *repository.Ticket) error
	NotifyNoteCreatedFunc       func(ctx context.Context, noteType string, contentPreview string, authorID string, shouldMention bool) error
	NotifyReviewCreatedFunc     func(ctx context.Context, noteTitle string, noteAuthorID string, reviewerID string, comment string) error
	SendDeadlineReminderFunc    func(ctx context.Context, ticket *repository.Ticket, daysOverdue int) error
	SendWaitingSentReminderFunc func(ctx context.Context, ticket *repository.Ticket) error
	GetUserIDByNameFunc         func(ctx context.Context, name string) (string, error)

	// イベントハンドラの記録用
	MessageCreatedHandler       func(messageID, channelID, userID, content string)
	MessageStampsUpdatedHandler func(messageID string, stamps []payload.MessageStamp)
}

var (
	_ Client        = (*MockService)(nil)
	_ MessageSender = (*MockService)(nil)
	_ EventHandler  = (*MockService)(nil)
)

// NewMockService はテスト用のモックサービスを作成する
func NewMockService() *MockService {
	return &MockService{
		StartFunc: func() error { return nil },
		APIFunc:   func() *traq.APIClient { return nil },
		PostMessageFunc: func(_ context.Context, _ string, _ string) error {
			return nil
		},
		PostDirectMessageFunc: func(_ context.Context, _ string, _ string) error {
			return nil
		},

		NotifyTicketCreatedFunc:     func(_ context.Context, _ *repository.Ticket) error { return nil },
		NotifyTicketUpdatedFunc:     func(_ context.Context, _ *repository.Ticket) error { return nil },
		NotifyNoteCreatedFunc:       func(_ context.Context, _ string, _ string, _ string, _ bool) error { return nil },
		NotifyReviewCreatedFunc:     func(_ context.Context, _ string, _ string, _ string, _ string) error { return nil },
		SendDeadlineReminderFunc:    func(_ context.Context, _ *repository.Ticket, _ int) error { return nil },
		SendWaitingSentReminderFunc: func(_ context.Context, _ *repository.Ticket) error { return nil },
		GetUserIDByNameFunc:         func(_ context.Context, _ string) (string, error) { return "user-id", nil },

		MessageCreatedHandler:       func(_, _, _, _ string) {},
		MessageStampsUpdatedHandler: func(_ string, _ []payload.MessageStamp) {},
	}
}

func (m *MockService) Start() error {
	return m.StartFunc()
}

func (m *MockService) API() *traq.APIClient {
	return m.APIFunc()
}

func (m *MockService) PostMessage(ctx context.Context, channelID string, content string) error {
	return m.PostMessageFunc(ctx, channelID, content)
}

func (m *MockService) PostDirectMessage(ctx context.Context, userID string, content string) error {
	return m.PostDirectMessageFunc(ctx, userID, content)
}

func (m *MockService) NotifyTicketCreated(ctx context.Context, ticket *repository.Ticket) error {
	return m.NotifyTicketCreatedFunc(ctx, ticket)
}

func (m *MockService) NotifyTicketUpdated(ctx context.Context, ticket *repository.Ticket) error {
	return m.NotifyTicketUpdatedFunc(ctx, ticket)
}

func (m *MockService) NotifyNoteCreated(ctx context.Context, noteType string, contentPreview string, authorID string, shouldMention bool) error {
	return m.NotifyNoteCreatedFunc(ctx, noteType, contentPreview, authorID, shouldMention)
}

func (m *MockService) NotifyReviewCreated(ctx context.Context, noteTitle string, noteAuthorID string, reviewerID string, comment string) error {
	return m.NotifyReviewCreatedFunc(ctx, noteTitle, noteAuthorID, reviewerID, comment)
}

func (m *MockService) SendDeadlineReminder(ctx context.Context, ticket *repository.Ticket, daysOverdue int) error {
	return m.SendDeadlineReminderFunc(ctx, ticket, daysOverdue)
}

func (m *MockService) SendWaitingSentReminder(ctx context.Context, ticket *repository.Ticket) error {
	return m.SendWaitingSentReminderFunc(ctx, ticket)
}

func (m *MockService) GetUserIDByName(ctx context.Context, name string) (string, error) {
	return m.GetUserIDByNameFunc(ctx, name)
}

func (m *MockService) OnMessageCreated(handler func(messageID, channelID, userID, content string)) {
	m.MessageCreatedHandler = handler
}

func (m *MockService) OnBotMessageStampsUpdated(handler func(messageID string, stamps []payload.MessageStamp)) {
	m.MessageStampsUpdatedHandler = handler
}

// SimulateMessageCreated はテストでメッセージ作成イベントをシミュレートする
func (m *MockService) SimulateMessageCreated(messageID, channelID, userID, content string) {
	if m.MessageCreatedHandler != nil {
		m.MessageCreatedHandler(messageID, channelID, userID, content)
	}
}
