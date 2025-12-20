package bot

import (
	"context"

	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/traq-ws-bot/payload"
)

// MockService はテスト用の Bot Service モック
type MockService struct {
	StartFunc             func() error
	APIFunc               func() *traq.APIClient
	PostMessageFunc       func(ctx context.Context, channelID string, content string) error
	PostDirectMessageFunc func(ctx context.Context, userID string, content string) error

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
