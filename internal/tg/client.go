package tg

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HTTPClient
	baseUrl    string
}

var ErrInvalidBotToken = errors.New("invalid bot token")

func NewClient(httpClient HTTPClient, botToken string) (*Client, error) {
	if botToken == "" {
		return nil, ErrInvalidBotToken
	}

	client := Client{
		httpClient,
		"https://api.telegram.org/bot" + botToken,
	}

	return &client, nil
}

func (cl *Client) newRequest(ctx context.Context) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "", cl.baseUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "achievement-bot")

	return req, nil
}

type envelope struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

type APIError struct {
	StatusCode  int
	Description string
}

func (err APIError) Error() string {
	return err.Description
}

func (cl *Client) doRequest(req *http.Request, result any) error {
	if req == nil {
		return nil
	}

	res, resErr := cl.httpClient.Do(req)
	if resErr != nil {
		return resErr
	}

	var env envelope
	jsonErr := json.NewDecoder(res.Body).Decode(&env)

	if jsonErr != nil {
		return APIError{res.StatusCode, res.Status}
	}

	if res.StatusCode >= 400 || !env.OK {
		return APIError{res.StatusCode, env.Description}
	}

	return json.Unmarshal(env.Result, result)
}
