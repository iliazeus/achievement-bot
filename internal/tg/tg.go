package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

type Bot struct {
	baseUrl string
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewBot(token string) (*Bot, error) {
	return NewBotWithContext(context.Background(), token)
}

func NewBotWithContext(ctx context.Context, token string) (*Bot, error) {
	if token == "" {
		return nil, fmt.Errorf("bot token not provided")
	}

	baseUrl := fmt.Sprint("https://api.telegram.org/bot", token)
	ctx, cancel := context.WithCancel(ctx)

	return &Bot{baseUrl, ctx, cancel}, nil
}

func (bot *Bot) Shutdown() {
	bot.cancel()
}

type Envelope[T any] struct {
	OK          bool    `json:"ok"`
	Description *string `json:"description"`
	Result      *T      `json:"result"`
}

func Send[TResult any](cl *http.Client, req *http.Request, result **TResult) error {
	res, err := cl.Do(req)
	if err != nil {
		return err
	}

	var body Envelope[TResult]
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		return err
	}

	if body.Description != nil {
		return fmt.Errorf(*body.Description)
	}

	if result != nil {
		*result = body.Result
	}
	return nil
}

type Handler = func(upd *Update, err error)

func (bot *Bot) Run(handler Handler) error {
	url := fmt.Sprint(bot.baseUrl, "/deleteWebhook")
	req, err := http.NewRequestWithContext(bot.ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	_ = Send[bool](http.DefaultClient, req, nil)

	cl := *http.DefaultClient
	cl.Timeout = 0

	offset := 0

	for {
		if bot.ctx.Err() != nil {
			return bot.ctx.Err()
		}

		url := fmt.Sprint(bot.baseUrl, "/getUpdates?timeout=60&offset=", offset)
		req, err := http.NewRequestWithContext(bot.ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		var upds *[]Update
		err = Send(&cl, req, &upds)
		if err != nil {
			handler(nil, err)
			continue
		}

		for i := range *upds {
			offset = (*upds)[i].UpdateID + 1
			go handler(&(*upds)[i], nil)
		}
	}
}

type Update struct {
	UpdateID    int          `json:"update_id"`
	InlineQuery *InlineQuery `json:"inline_query"`
}

type InlineQuery struct {
	ID    string `json:"id"`
	Query string `json:"query"`
}

type Message struct {
	MessageID int      `json:"message_id"`
	Sticker   *Sticker `json:"sticker"`
}

type Sticker struct {
	FileID string `json:"file_id"`
}

type SendStickerBytes struct {
	ChatId       int
	StickerBytes []byte
	Emoji        string
}

func (data SendStickerBytes) Request(bot *Bot) (*Message, error) {
	url := fmt.Sprint(bot.baseUrl, "/sendSticker")

	buf := &bytes.Buffer{}
	form := multipart.NewWriter(buf)

	err := form.WriteField("chat_id", fmt.Sprint(data.ChatId))
	if err != nil {
		return nil, err
	}

	if data.Emoji != "" {
		err = form.WriteField("emoji", data.Emoji)
		if err != nil {
			return nil, err
		}
	}

	sticker, err := form.CreateFormFile("sticker", "sticker.webp")
	if err != nil {
		return nil, err
	}

	_, err = sticker.Write(data.StickerBytes)
	if err != nil {
		return nil, err
	}

	err = form.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(bot.ctx, "POST", url, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", form.FormDataContentType())

	var message *Message
	err = Send(http.DefaultClient, req, &message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

type AnswerInlineQueryWithStickers struct {
	InlineQueryID  string
	StickerFileIDs []string
}

func (data AnswerInlineQueryWithStickers) Request(bot *Bot) error {
	url := fmt.Sprint(bot.baseUrl, "/answerInlineQuery")

	body := map[string]any{
		"inline_query_id": data.InlineQueryID,
	}

	results := make([]map[string]any, 0, len(data.StickerFileIDs))

	for i, fileId := range data.StickerFileIDs {
		results = append(results, map[string]any{
			"type":            "sticker",
			"id":              i,
			"sticker_file_id": fileId,
		})
	}

	body["results"] = results

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(bot.ctx, "POST", url, buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	return Send[bool](http.DefaultClient, req, nil)
}
