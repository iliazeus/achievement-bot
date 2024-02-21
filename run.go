package main

import (
	"context"
	"net/http"

	"log/slog"
	"time"

	"github.com/iliazeus/achievement-bot/internal/sticker"
	"github.com/iliazeus/achievement-bot/internal/tg"
	_ "github.com/joho/godotenv/autoload"
)

func run(slog *slog.Logger, botToken string, tempChatId int) error {
	stickerMaker, err := sticker.NewStickerMaker()
	if err != nil {
		return err
	}

	client, err := tg.NewClient(http.DefaultClient, botToken)
	if err != nil {
		return err
	}

	slog.Info("starting update loop")

	return client.RunUpdateLoop(context.Background(), func(ctx context.Context, upd *tg.Update, err error) {
		if err != nil {
			slog.Error(err.Error())
			time.Sleep(1 * time.Second)
			return
		}

		slog := slog.With("update_id", upd.UpdateID)
		slog.Debug("got update")

		switch {
		case upd.InlineQuery != nil:
			query := upd.InlineQuery

			slog := slog.With("inline_query.id", query.ID)
			slog.Info("got inline query", "query", query.Query)

			sticker, err := stickerMaker.MakeSticker(query.Query)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			msg, err := client.SendSticker(ctx, tempChatId, sticker)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			err = client.AnswerInlineQuery(
				ctx, query.ID,
				tg.InlineQueryAnswer{ID: 0, Type: "sticker", StickerFileID: msg.Sticker.FileID},
			)
			if err != nil {
				slog.Error(err.Error())
				return
			}
		}
	})
}
