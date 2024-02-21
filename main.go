package main

import (
	"log/slog"
	"os"
	"strconv"

	// "crypto/tls"
	// "net/http"
	// "net/url"

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

	err := run(slog, botToken, tempChatId)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
