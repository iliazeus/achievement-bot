package main

import (
	"context"
	"net/http"

	// "crypto/tls"

	"log/slog"
	"os"
	"strconv"
	"time"

	// "net/http"
	// "net/url"

	"github.com/iliazeus/achievement-bot/internal/sticker"
	"github.com/iliazeus/achievement-bot/internal/tg"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	slog := slog.Default()

	// proxyUrl, _ := url.Parse("http://127.0.0.1:8080")
	// http.DefaultClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	stickerMaker, err := sticker.NewStickerMaker()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var tempChatId int
	{
		val, err := strconv.Atoi(os.Getenv("TELEGRAM_TEMP_CHAT_ID"))
		if err != nil {
			slog.Error("TELEGRAM_TEMP_CHAT_ID not set")
			os.Exit(1)
		}
		tempChatId = val
	}

	client, err := tg.NewClient(http.DefaultClient, os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = client.RunUpdateLoop(context.Background(), func(ctx context.Context, upd *tg.Update, err error) {
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

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
