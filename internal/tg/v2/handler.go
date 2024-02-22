package tg

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

type ErrorHandler = func(context.Context, *Client, error) error
type UpdateHandler = func(context.Context, *Client, *Update) error
type InlineQueryHandler = func(context.Context, *Client, *InlineQuery) error

type Handlers struct {
	OnAnyUpdate       UpdateHandler
	OnUnhandledUpdate UpdateHandler

	OnInlineQuery InlineQueryHandler
}

// does not own channel
type EventSource = func(context.Context, chan<- *Update) error

func (cl *Client) LongPollEventSource(timeout int) EventSource {
	return func(ctx context.Context, upds chan<- *Update) error {
		err := cl.DeleteWebhook(ctx)
		if err != nil {
			return err
		}

		offset := 0

		for {
			batch, err := cl.GetUpdates(ctx, offset, timeout)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				slog.Error(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}

			for i := range batch {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case upds <- &batch[i]:
					offset = batch[i].UpdateID + 1
					continue
				}
			}
		}
	}
}

func (hs *Handlers) RunWithSource(ctx context.Context, cl *Client, srcFn EventSource) error {
	upds := make(chan *Update)
	defer close(upds)

	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return srcFn(ctx, upds)
	})

loop:
	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			break loop

		case upd := <-upds:
			var err error

			if hs.OnAnyUpdate != nil {
				err = hs.OnAnyUpdate(ctx, cl, upd)
			}

			if err == nil {
				switch {
				case upd.InlineQuery != nil && hs.OnInlineQuery != nil:
					err = hs.OnInlineQuery(ctx, cl, upd.InlineQuery)
				case hs.OnUnhandledUpdate != nil:
					err = hs.OnUnhandledUpdate(ctx, cl, upd)
				}
			}

			if err != nil {
				if err == ctx.Err() {
					return err
				} else {
					slog.Error(err.Error())
				}
			}
		}
	}

	return g.Wait()
}
