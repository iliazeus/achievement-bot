package tg

import "encoding/json"

type envelope struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

type Error struct {
	Description string
}

func (err *Error) Error() string {
	return err.Description
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

type InlineQueryAnswer struct {
	Type          string `json:"type"`
	ID            int    `json:"id"`
	StickerFileID string `json:"sticker_file_id"`
}
