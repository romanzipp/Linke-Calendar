package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const zetkinAPIBaseURL = "https://api.zetkin.die-linke.de/v1/orgs"

type ZetkinClient struct {
	orgID  int
	client *http.Client
}

type ZetkinResponse struct {
	Data []ZetkinEvent `json:"data"`
}

type ZetkinEvent struct {
	ID           int                `json:"id"`
	Title        string             `json:"title"`
	StartTime    string             `json:"start_time"`
	EndTime      string             `json:"end_time"`
	InfoText     string             `json:"info_text"`
	URL          string             `json:"url"`
	Activity     *ZetkinActivity    `json:"activity"`
	Location     *ZetkinLocation    `json:"location"`
	Contact      *ZetkinContact     `json:"contact"`
	Organization ZetkinOrganization `json:"organization"`
	Cancelled    *string            `json:"cancelled"`
}

type ZetkinActivity struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type ZetkinLocation struct {
	ID    int     `json:"id"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Title string  `json:"title"`
}

type ZetkinContact struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ZetkinOrganization struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func NewZetkinClient(orgID int, timeout time.Duration) *ZetkinClient {
	return &ZetkinClient{
		orgID: orgID,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (z *ZetkinClient) FetchAllEvents() ([]ZetkinEvent, error) {
	url := fmt.Sprintf("%s/%d/actions", zetkinAPIBaseURL, z.orgID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:145.0) Gecko/20100101 Firefox/145.0")
	req.Header.Set("Accept", "application/json")

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

	var filtered []ZetkinEvent
	for _, event := range zetkinResp.Data {
		if event.Cancelled == nil {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}

func parseZetkinTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}
