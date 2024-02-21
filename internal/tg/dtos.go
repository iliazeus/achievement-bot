package tg

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
