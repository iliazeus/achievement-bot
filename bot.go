package main

import (
	"context"

	"log/slog"

	"github.com/iliazeus/achievement-bot/internal/sticker"
	"github.com/iliazeus/achievement-bot/internal/tg/v2"
	_ "github.com/joho/godotenv/autoload"
)

func newBot(tempChatId int) (*tg.Handlers, error) {
	stickerMaker, err := sticker.NewStickerMaker()
	if err != nil {
		return nil, err
	}

	bot := &tg.Handlers{
		OnAnyUpdate: func(ctx context.Context, c *tg.Client, u *tg.Update) error {
			ctx = withLoggedValue(ctx, "update_id", u.UpdateID)
			slog.DebugContext(ctx, "got update")
			return nil
		},

		OnInlineQuery: func(ctx context.Context, c *tg.Client, iq *tg.InlineQuery) error {
			ctx = withLoggedValue(ctx, "inline_query_id", iq.ID)
			slog.InfoContext(ctx, "got inline query")

			sticker, err := stickerMaker.MakeSticker(iq.Query)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
				return err
			}

			msg, err := c.SendSticker(ctx, tempChatId, sticker)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
				return err
			}

			err = c.AnswerInlineQuery(
				ctx, iq.ID,
				[]tg.InlineQueryAnswer{
					{ID: 0, Type: "sticker", StickerFileID: msg.Sticker.FileID},
				},
			)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
				return err
			}

			return nil
		},
	}

	return bot, nil
}
