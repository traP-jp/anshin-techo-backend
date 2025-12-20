package bot

import (
	"context"

	"github.com/traPtitech/go-traq"
	"github.com/traPtitech/traq-ws-bot/payload"
)

// BotClient は traQ Bot の操作を抽象化したインターフェース
// テスト時にはこのインターフェースをモックすることで、実際の traQ との通信を避けられる
type BotClient interface {
	// Start は Bot の WebSocket 接続を開始する
	Start() error
	// API は traQ API クライアントを返す
	API() *traq.APIClient

	// MessageSender インターフェースを埋め込み
	MessageSender

	// EventHandler インターフェースを埋め込み
	EventHandler
}

// MessageSender はメッセージ送信機能を抽象化したインターフェース
type MessageSender interface {
	// PostMessage は指定されたチャンネルにメッセージを送信する
	PostMessage(ctx context.Context, channelID string, content string) error
	// PostDirectMessage は指定されたユーザーにダイレクトメッセージを送信する
	PostDirectMessage(ctx context.Context, userID string, content string) error
}

// EventHandler は Bot イベントのハンドラを抽象化したインターフェース
type EventHandler interface {
	// OnMessageCreated はメッセージ作成イベントのハンドラを登録する
	OnMessageCreated(handler func(messageID, channelID, userID, content string))
	// OnChannelCreated はチャンネル作成イベントのハンドラを登録する
	OnBotMessageStampsUpdated(handler func(messageID string, stamps []payload.MessageStamp))
}
