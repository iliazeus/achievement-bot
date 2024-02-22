package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"

	// "crypto/tls"
	// "net/http"
	// "net/url"

	"github.com/iliazeus/achievement-bot/internal/tg/v2"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	slog := slog.Default()

	// proxyUrl, _ := url.Parse("http://127.0.0.1:8080")
	// http.DefaultClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		slog.Error("TELEGRAM_BOT_TOKEN not set")
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

	bot, err := newBot(tempChatId)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	cl, err := tg.NewClient(botToken)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info("starting bot")

	err = bot.RunWithSource(
		context.Background(),
		cl, cl.LongPollEventSource(10000),
	)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
