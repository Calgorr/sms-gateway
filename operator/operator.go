package operator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/calgorr/sms-gateway/config"
)

type Client interface {
	Send(ctx context.Context, to, text string) error
}

// httpOperator implements Client against the operator's real HTTP API.
type httpOperator struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewOperator(config config.Operator) Client {
	return &httpOperator{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

type sendRequest struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

type sendResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func (o *httpOperator) Send(ctx context.Context, to, text string) error {
	body, err := json.Marshal(sendRequest{To: to, Text: text})
	if err != nil {
		return fmt.Errorf("marshal operator request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build operator request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("operator request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var sr sendResponse
		_ = json.NewDecoder(resp.Body).Decode(&sr)
		if sr.Error != "" {
			return fmt.Errorf("operator rejected message: %s (status %d)", sr.Error, resp.StatusCode)
		}
		return fmt.Errorf("operator returned status %d", resp.StatusCode)
	}

	return nil
}
