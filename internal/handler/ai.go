package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/traP-jp/anshin-techo-backend/internal/api"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

func newAIClient() *openai.Client {
	apiKey := os.Getenv("LITELLM_API_KEY")
	baseURL := os.Getenv("LITELLM_BASE_URL")

	if baseURL == "" {
		baseURL = "https://llm-proxy.trap.jp"
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return openai.NewClientWithConfig(config)
}

// POST /tickets/{ticketId}/ai/generate
func (h *Handler) TicketsTicketIdAiGeneratePost(ctx context.Context, req *api.TicketsTicketIdAiGeneratePostReq, params api.TicketsTicketIdAiGeneratePostParams) (api.TicketsTicketIdAiGeneratePostRes, error) {
	ticket, err := h.repo.GetTicketByID(ctx, params.TicketId)
	if err != nil {
		if err == repository.ErrTicketNotFound {
			return &api.TicketsTicketIdAiGeneratePostNotFound{}, nil
		}
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	notes, err := h.repo.GetNotes(ctx, params.TicketId)
	if err != nil {
		return nil, fmt.Errorf("get notes: %w", err)
	}

	systemPrompt := `
あなたはtraPの渉外担当をサポートするAIアシスタントです。
ユーザーから提供される「案件情報」と「これまでの経緯」を元に、次に送るべき返信メールのドラフトを作成してください。
`
	contextText := fmt.Sprintf("【案件名】: %s\n【詳細】: %s\n\n【これまでの経緯】:\n", ticket.Title, ticket.Description.String)
	for _, n := range notes {
		if n.Status == "sent" {
			contextText += fmt.Sprintf("- %s (%s): %s\n", n.UserID, n.Type, n.Content)
		}
	}

	instruction := "特になし"
	if req.Instruction.Set {
		instruction = req.Instruction.Value
	}
	userPrompt := fmt.Sprintf("%s\n\n【今回の指示】: %s\n\n返信ドラフトを作成してください。", contextText, instruction)

	client := newAIClient()

	streamReq := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		Stream: true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, streamReq)
	if err != nil {
		return nil, fmt.Errorf("ai stream error: %w", err)
	}

	reader, writer := io.Pipe()

	go func() {
		defer stream.Close()
		defer writer.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				fmt.Printf("Stream error: %v\n", err)
				return
			}

			chunk := response.Choices[0].Delta.Content
			if chunk == "" {
				continue
			}

			msg := fmt.Sprintf("data: %s\n\n", chunk)
			if _, err := writer.Write([]byte(msg)); err != nil {
				return
			}
		}
	}()

	return &api.TicketsTicketIdAiGeneratePostOK{
		Data: reader,
	}, nil
}

// POST /tickets/{ticketId}/notes/{noteId}/ai/review
func (h *Handler) TicketsTicketIdNotesNoteIdAiReviewPost(ctx context.Context, params api.TicketsTicketIdNotesNoteIdAiReviewPostParams) (api.TicketsTicketIdNotesNoteIdAiReviewPostRes, error) {
	note, err := h.repo.GetNoteByID(ctx, params.TicketId, params.NoteId)
	if err != nil {
		return &api.TicketsTicketIdNotesNoteIdAiReviewPostNotFound{}, nil
	}
	ticket, err := h.repo.GetTicketByID(ctx, params.TicketId)
	if err != nil {
		return nil, fmt.Errorf("get ticket: %w", err)
	}

	systemPrompt := `
あなたはtraPの渉外担当の補佐役です。
部員が作成した「外部への返信メールの下書き」をレビューしてください。
指摘事項を箇条書きで、リアルタイムにフィードバックしてください。
`
	userPrompt := fmt.Sprintf(
		"【案件概要】: %s\n\n【レビュー対象のメール下書き】:\n%s\n\nレビューをお願いします。",
		ticket.Title,
		note.Content,
	)

	client := newAIClient()

	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		},
		Stream: true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ai stream error: %w", err)
	}

	reader, writer := io.Pipe()

	go func() {
		defer stream.Close()
		defer writer.Close()

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				fmt.Printf("Stream error: %v\n", err)
				return
			}

			chunk := response.Choices[0].Delta.Content
			if chunk == "" {
				continue
			}

			msg := fmt.Sprintf("data: %s\n\n", chunk)
			if _, err := writer.Write([]byte(msg)); err != nil {
				return
			}
		}
	}()

	return &api.TicketsTicketIdNotesNoteIdAiReviewPostOK{
		Data: reader,
	}, nil
}
