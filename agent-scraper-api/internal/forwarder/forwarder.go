package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ArminDashti/agent-scraper-api/internal/store"
)

type Client struct {
	store      *store.Store
	httpClient *http.Client
	envURL     string
}

func NewClient(st *store.Store, envURL string) *Client {
	return &Client{
		store:      st,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		envURL:     envURL,
	}
}

type expensePayload struct {
	Shop    string `json:"shop"`
	Item    string `json:"item"`
	Expense string `json:"expense"`
}

func (c *Client) Deliver(ctx context.Context, row store.ExpenseRow) error {
	targetURL, err := c.store.TargetAPIURL(ctx, c.envURL)
	if err != nil {
		return err
	}
	if targetURL == "" {
		return c.store.InsertDelivery(ctx, row.ID, "", "skipped", nil, "TARGET_API_URL is empty")
	}

	body, err := json.Marshal(expensePayload{Shop: row.Shop, Item: row.Item, Expense: row.Expense})
	if err != nil {
		return c.store.InsertDelivery(ctx, row.ID, targetURL, "failed", nil, err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(body))
	if err != nil {
		return c.store.InsertDelivery(ctx, row.ID, targetURL, "failed", nil, err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return c.store.InsertDelivery(ctx, row.ID, targetURL, "failed", nil, err.Error())
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code < 200 || code >= 300 {
		return c.store.InsertDelivery(ctx, row.ID, targetURL, "failed", &code, fmt.Sprintf("unexpected status %d", code))
	}
	return c.store.InsertDelivery(ctx, row.ID, targetURL, "sent", &code, "")
}
