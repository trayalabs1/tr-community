package moengage

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/fctx"
	"github.com/Southclaws/fault/fmsg"
	"go.uber.org/fx"

	"github.com/Southclaws/storyden/internal/config"
)

const defaultPlatform = "web"

var errMoEngageRequestFailed = fault.New("moengage request failed")

type Sender interface {
	AddEvent(ctx context.Context, data EventData) error
}

type EventData struct {
	CaseID    string
	Event     string
	EventAttr map[string]any
	Platform  string
}

type Client struct {
	logger     *slog.Logger
	httpClient *http.Client
	url        string
	authHeader string
}

func Build() fx.Option {
	return fx.Provide(newClient)
}

func newClient(cfg config.Config, logger *slog.Logger) (Sender, error) {
	if cfg.MoEngageAppKey == "" || cfg.MoEngageAuthKey == "" {
		return nil, nil
	}

	auth := base64.StdEncoding.EncodeToString([]byte(cfg.MoEngageAppKey + ":" + cfg.MoEngageAuthKey))

	return &Client{
		logger:     logger.With(slog.String("client", "moengage")),
		httpClient: &http.Client{Timeout: 10 * time.Second},
		url:        fmt.Sprintf("https://api-03.moengage.com/v1/event/%s", cfg.MoEngageAppKey),
		authHeader: "Basic " + auth,
	}, nil
}

func (c *Client) AddEvent(ctx context.Context, data EventData) error {
	if c == nil {
		return fault.New("moengage client is nil")
	}

	platform := data.Platform
	if platform == "" {
		platform = defaultPlatform
	}

	payload := struct {
		Type       string        `json:"type"`
		CustomerID string        `json:"customer_id"`
		Actions    []eventAction `json:"actions"`
	}{
		Type:       "event",
		CustomerID: data.CaseID,
		Actions: []eventAction{
			{
				Action:     data.Event,
				Attributes: data.EventAttr,
				Platform:   platform,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fault.Wrap(err, fctx.With(ctx))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := parseErrorMessage(resp.Body)
		if message == "" {
			message = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		}

		return fault.Wrap(errMoEngageRequestFailed, fctx.With(ctx), fmsg.With(message))
	}

	c.logger.Debug("moengage event sent successfully",
		slog.String("event", data.Event),
		slog.String("case_id", data.CaseID),
	)

	return nil
}

type eventAction struct {
	Action     string         `json:"action"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Platform   string         `json:"platform"`
}

func parseErrorMessage(body io.Reader) string {
	payload, err := io.ReadAll(body)
	if err != nil || len(payload) == 0 {
		return ""
	}

	var envelope struct {
		ErrorMessage string `json:"error_message"`
		Message      string `json:"message"`
	}
	if err := json.Unmarshal(payload, &envelope); err == nil {
		if envelope.ErrorMessage != "" {
			return envelope.ErrorMessage
		}
		if envelope.Message != "" {
			return envelope.Message
		}
	}

	return string(payload)
}
