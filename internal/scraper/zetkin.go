package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const zetkinAPIURL = "https://app.zetkin.die-linke.de/api/rpc"

type ZetkinClient struct {
	cookie string
	client *http.Client
}

type ZetkinRequest struct {
	Func   string                 `json:"func"`
	Params map[string]interface{} `json:"params"`
}

type ZetkinResponse struct {
	Result []ZetkinEvent `json:"result"`
}

type ZetkinEvent struct {
	ID           int                `json:"id"`
	Title        string             `json:"title"`
	StartTime    string             `json:"start_time"`
	EndTime      string             `json:"end_time"`
	InfoText     string             `json:"info_text"`
	URL          string             `json:"url"`
	Location     *ZetkinLocation    `json:"location"`
	Organization ZetkinOrganization `json:"organization"`
	Cancelled    *string            `json:"cancelled"`
}

type ZetkinLocation struct {
	ID    int     `json:"id"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Title string  `json:"title"`
}

type ZetkinOrganization struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func NewZetkinClient(cookie string, timeout time.Duration) *ZetkinClient {
	return &ZetkinClient{
		cookie: cookie,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (z *ZetkinClient) FetchAllEvents() ([]ZetkinEvent, error) {
	reqBody := ZetkinRequest{
		Func:   "getAllEvents",
		Params: map[string]interface{}{},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", zetkinAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:145.0) Gecko/20100101 Firefox/145.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", fmt.Sprintf("zsid=%s", z.cookie))

	resp, err := z.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		preview := string(body)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		log.Printf("Zetkin API error response: %s", preview)
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var zetkinResp ZetkinResponse
	if err := json.Unmarshal(body, &zetkinResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return zetkinResp.Result, nil
}

func (z *ZetkinClient) FilterByOrganization(events []ZetkinEvent, organization string) []ZetkinEvent {
	var filtered []ZetkinEvent
	for _, event := range events {
		if event.Organization.Title == organization && event.Cancelled == nil {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func parseZetkinTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}
