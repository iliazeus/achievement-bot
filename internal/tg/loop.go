package tg

import (
	"context"
	"time"
)

type UpdateHandler = func(context.Context, *Update, error)

func (cl *Client) RunUpdateLoop(ctx context.Context, handler UpdateHandler) error {
	_ = cl.DeleteWebhook(ctx)

	offset := 0

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		updates, err := cl.GetUpdates(ctx, offset, 10000)
		if err != nil {
			handler(ctx, nil, err)
			time.Sleep(1 * time.Second)
			continue
		}

		for i := range updates {
			offset = updates[i].UpdateID + 1
			go handler(ctx, &updates[i], nil)
		}
	}
}
