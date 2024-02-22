package tg

import (
	"bytes"
	"context"
)

func (cl *Client) DeleteWebhook(ctx context.Context) error {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	res, err := cl.rest.R().
		SetContext(ctx).
		Post("/deleteWebhook")

	_, err = unwrapResponse(res, err, true)
	return err
}

func (cl *Client) GetUpdates(
	ctx context.Context,
	offset int,
	timeout int,
) ([]Update, error) {
	ctx, cancel := context.WithTimeout(ctx, seconds(timeout+10))
	defer cancel()

	var result []Update
	res, err := cl.rest.R().
		SetContext(ctx).
		SetQueryParam("offset", itoa(offset)).
		SetQueryParam("timeout", itoa(timeout)).
		Get("/getUpdates")

	return unwrapResponse(res, err, result)
}

func (cl *Client) SendSticker(
	ctx context.Context,
	chatId int,
	sticker []byte,
) (*Message, error) {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	var result *Message
	res, err := cl.rest.R().
		SetContext(ctx).
		SetFileReader("sticker", "sticker.webp", bytes.NewReader(sticker)).
		SetFormData(fM{"chat_id": itoa(chatId)}).
		Post("/sendSticker")

	return unwrapResponse(res, err, result)
}

func (cl *Client) AnswerInlineQuery(
	ctx context.Context,
	inlineQueryId string,
	answers []InlineQueryAnswer,
) error {
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	res, err := cl.rest.R().
		SetContext(ctx).
		SetBody(jM{"inline_query_id": inlineQueryId, "results": answers}).
		Post("/answerInlineQuery")

	_, err = unwrapResponse(res, err, true)
	return err
}
