package bot

import (
	"context"
	"log"

	"github.com/traPtitech/traq-ws-bot/payload"
)

// HandlerService は Bot のイベントハンドラを管理するサービス
type HandlerService struct {
	// 将来的に repository や他のサービスを注入する可能性がある
	messageSender MessageSender
}

// NewHandlerService は新しい HandlerService を作成する
func NewHandlerService(messageSender MessageSender) *HandlerService {
	return &HandlerService{
		messageSender: messageSender,
	}
}

// RegisterHandlers は Bot のイベントハンドラを登録する
func (h *HandlerService) RegisterHandlers(eventHandler EventHandler) {
	// チャンネル作成イベントのハンドラ
	eventHandler.OnBotMessageStampsUpdated(h.handleBotMessageStampsUpdated)

	// メッセージ作成イベントのハンドラ
	eventHandler.OnMessageCreated(h.handleMessageCreated)
}

// handleMessageCreated はメッセージ作成時の処理を実行する
func (h *HandlerService) handleMessageCreated(messageID, channelID, userID, content string) {
	log.Printf("Message created: %s in %s by %s", messageID, channelID, userID)

	// TODO: 実際のビジネスロジックを実装

	// 簡単なエコーボットの例（実際の実装では条件分岐が必要）
	if content == "ping" {
		ctx := context.Background()
		if err := h.messageSender.PostMessage(ctx, channelID, "pong!"); err != nil {
			log.Printf("Failed to respond to ping in channel %s: %v", channelID, err)
		}
	}
}

func (h *HandlerService) handleBotMessageStampsUpdated(messageID string, stamps []payload.MessageStamp) {
	log.Printf("Bot message stamps updated: %s with stamps %v", messageID, stamps)

	// TODO: 実際のビジネスロジックを実装
	/*
		// 例：特定のスタンプが付けられたらメッセージを送信
		for _, stamp := range stamps {
			if stamp.StampID == "👍" {
				ctx := context.Background()
				if err := h.messageSender.PostMessage(ctx, messageID, "いいねが付きました！"); err != nil {
					log.Printf("Failed to respond to stamp on message %s: %v", messageID, err)
				}
			}
		}
	*/
}
