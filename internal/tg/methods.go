package tg

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"strconv"
)

func (cl *Client) DeleteWebhook(ctx context.Context) error {
	req, err := cl.newRequest(ctx)
	if err != nil {
		return err
	}

	req.Method = "POST"
	req.URL = req.URL.JoinPath("deleteWebhook")

	var result bool
	return cl.doRequest(req, &result)
}

func (cl *Client) GetUpdates(
	ctx context.Context,
	offset int,
	timeout int,
) ([]Update, error) {
	req, err := cl.newRequest(ctx)
	if err != nil {
		return nil, err
	}

	req.Method = "GET"
	req.URL = req.URL.JoinPath("getUpdates")

	query := req.URL.Query()
	query.Set("offset", strconv.Itoa(offset))
	query.Set("timeout", strconv.Itoa(timeout))
	req.URL.RawQuery = query.Encode()

	var result []Update
	err = cl.doRequest(req, &result)
	return result, err
}

func (cl *Client) SendSticker(
	ctx context.Context,
	chatId int,
	sticker []byte,
) (*Message, error) {
	req, err := cl.newRequest(ctx)
	if err != nil {
		return nil, err
	}

	req.Method = "POST"
	req.URL = req.URL.JoinPath("sendSticker")

	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	_ = form.WriteField("chat_id", strconv.Itoa(chatId))
	stickerFile, _ := form.CreateFormFile("sticker", "sticker.webp")
	_, _ = stickerFile.Write(sticker)

	err = form.Close()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Body = io.NopCloser(body)
	req.ContentLength = int64(body.Len())

	var result *Message
	err = cl.doRequest(req, &result)
	return result, err
}

func (cl *Client) AnswerInlineQuery(
	ctx context.Context,
	inlineQueryId string,
	answers ...InlineQueryAnswer,
) error {
	req, err := cl.newRequest(ctx)
	if err != nil {
		return err
	}

	req.Method = "POST"
	req.URL = req.URL.JoinPath("answerInlineQuery")

	body := &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(map[string]any{
		"inline_query_id": inlineQueryId,
		"results":         answers,
	})
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(body)
	req.ContentLength = int64(body.Len())

	var result bool
	return cl.doRequest(req, &result)
}
