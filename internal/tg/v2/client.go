package tg

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	rest *resty.Client
}

var ErrEmptyBotToken = errors.New("empty bot token")

func NewClient(botToken string) (*Client, error) {
	return NewClientWithResty(resty.New(), botToken)
}

// takes ownership of rest
func NewClientWithResty(rest *resty.Client, botToken string) (*Client, error) {
	if botToken == "" {
		return nil, ErrEmptyBotToken
	}

	// rest.SetDebug(true)

	rest.SetBaseURL("https://api.telegram.org/bot" + botToken)
	rest.SetHeader("User-Agent", "achievement-bot")

	rest.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetResult(&envelope{}).SetError(&Error{})
		return nil
	})

	rest.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		if r.IsSuccess() {
			// to guard against 200 OK-errors
			return r.Result().(*envelope).Err()
		} else {
			return nil
		}
	})

	return &Client{rest}, nil
}

func withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, seconds(10))
}

func (err *Error) UnmarshalJSON(data []byte) error {
	var env envelope

	envErr := json.Unmarshal(data, &env)
	if envErr != nil {
		return envErr
	}

	if !env.OK {
		*err = Error{env.Description}
	}

	return nil
}

func (env *envelope) Err() error {
	if env.OK {
		return nil
	} else {
		return &Error{env.Description}
	}
}

func unwrapResponse[TResult any](res *resty.Response, resErr error, result TResult) (TResult, error) {
	if resErr != nil {
		return result, resErr
	}

	if res.IsError() {
		return result, res.Error().(*Error)
	}

	env := res.Result().(*envelope)
	jsonErr := json.Unmarshal(env.Result, &result)
	return result, jsonErr
}
