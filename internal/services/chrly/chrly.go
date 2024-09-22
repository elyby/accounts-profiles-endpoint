package chrly

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/viper"
)

type Chrly struct {
	baseUrl    string
	httpClient *http.Client
}

func New(baseUrl string, httpClient *http.Client) *Chrly {
	return &Chrly{
		baseUrl,
		httpClient,
	}
}

func NewWithConfig(config *viper.Viper) (*Chrly, error) {
	config.SetDefault("chrly.url", "http://skinsystem.ely.by")
	chrlyUrl := strings.Trim(config.GetString("chrly.url"), "/")

	return New(chrlyUrl, &http.Client{}), nil
}

func (c *Chrly) GetTexturesByUsername(ctx context.Context, username string) ([]byte, error) {
	req, err := c.newRequestWithContext(ctx, "GET", fmt.Sprintf("%s/textures/%s", c.baseUrl, username), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to perform a request to Chrly: %w", err)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	textures, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response from Chrly: %w", err)
	}

	return textures, nil
}

func (c *Chrly) newRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	hub := sentry.GetHubFromContext(ctx)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to form a correct request to Chrly: %w", err)
	}

	req.Header.Add(sentry.SentryTraceHeader, hub.GetTraceparent())
	req.Header.Add(sentry.SentryBaggageHeader, hub.GetBaggage())

	return req, nil
}
