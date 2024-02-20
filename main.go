package main

import (
	"bytes"
	// "crypto/tls"
	"image"
	"image/color"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/fogleman/gg"
	"github.com/kolesa-team/go-webp/webp"

	// "net/http"
	// "net/url"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"

	"github.com/iliazeus/achievement-bot/internal/assets"
	"github.com/iliazeus/achievement-bot/internal/tg"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	slog := slog.Default()

	// proxyUrl, _ := url.Parse("http://127.0.0.1:8080")
	// http.DefaultClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	var templateImage image.Image
	{
		file, err := assets.FS.Open("template.png")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		templateImage, _, err = image.Decode(file)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}

	var fontFace font.Face
	{
		font, err := opentype.Parse(goregular.TTF)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		fontFace, err = opentype.NewFace(font, &opentype.FaceOptions{
			Size: 48,
			DPI:  72,
		})
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}

	// var fontFace = opentype.NewFace()

	var tempChatId int
	{
		val, err := strconv.Atoi(os.Getenv("TELEGRAM_TEMP_CHAT_ID"))
		if err != nil {
			slog.Error("TELEGRAM_TEMP_CHAT_ID not set")
			os.Exit(1)
		}
		tempChatId = val
	}

	bot, err := tg.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	err = bot.Run(func(upd *tg.Update, err error) {
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

			canvas := gg.NewContextForImage(templateImage)

			canvas.SetColor(color.White)
			canvas.SetFontFace(fontFace)
			canvas.DrawStringWrapped(
				query.Query,
				180, 80, // x, y
				0.0, 0.7, // anchorX, anchorY; mostly found by trial & error
				500, 1.0, // width, lineSpacing
				gg.AlignLeft,
			)

			buf := &bytes.Buffer{}
			err := webp.Encode(buf, canvas.Image(), nil)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			msg, err := tg.SendStickerBytes{
				ChatId:       tempChatId,
				StickerBytes: buf.Bytes(),
			}.Request(bot)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			err = tg.AnswerInlineQueryWithStickers{
				InlineQueryID:  query.ID,
				StickerFileIDs: []string{msg.Sticker.FileID},
			}.Request(bot)
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
