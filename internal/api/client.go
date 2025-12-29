package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sketchfab-forecasts/internal/models"
	"time"

	"github.com/sirupsen/logrus"
)

// Client для работы с Sketchfab API
type Client struct {
	baseURL    string
	apiToken   string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewClient создает новый API клиент
func NewClient(baseURL, apiToken string, logger *logrus.Logger) *Client {
	return &Client{
		baseURL:  baseURL,
		apiToken: apiToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// FetchModels получает модели с Sketchfab API
func (c *Client) FetchModels(params map[string]string, limit int) ([]models.SketchfabModel, error) {
	allModels := []models.SketchfabModel{}
	nextURL := c.buildURL("/models", params)

	for len(allModels) < limit && nextURL != "" {
		c.logger.Infof("Fetching models from: %s", nextURL)

		resp, err := c.makeRequest(nextURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch models: %w", err)
		}

		var apiResp models.APIResponse
		// Read full body so we can include a snippet on decode errors (helps debug timestamp formats)
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			// Save full body to file for debugging
			_ = os.MkdirAll("data", 0755)
			_ = os.WriteFile("data/failed_response.json", body, 0644)
			return nil, fmt.Errorf("failed to decode response: %w; full body saved to data/failed_response.json", err)
		}

		allModels = append(allModels, apiResp.Results...)
		nextURL = apiResp.Next

		c.logger.Infof("Fetched %d models, total: %d", len(apiResp.Results), len(allModels))

		// Rate limiting - избегаем блокировки API
		time.Sleep(1 * time.Second)

		if len(allModels) >= limit {
			break
		}
	}

	if len(allModels) > limit {
		allModels = allModels[:limit]
	}

	return allModels, nil
}

// GetModelByUID получает конкретную модель по UID
func (c *Client) GetModelByUID(uid string) (*models.SketchfabModel, error) {
	url := c.buildURL(fmt.Sprintf("/models/%s", uid), nil)

	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch model: %w", err)
	}
	defer resp.Body.Close()

	var model models.SketchfabModel
	if err := json.NewDecoder(resp.Body).Decode(&model); err != nil {
		return nil, fmt.Errorf("failed to decode model: %w", err)
	}

	return &model, nil
}

// buildURL создает URL с параметрами
func (c *Client) buildURL(endpoint string, params map[string]string) string {
	u, _ := url.Parse(c.baseURL + endpoint)

	if params != nil {
		q := u.Query()
		for key, value := range params {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// makeRequest выполняет HTTP запрос к API
func (c *Client) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.apiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// SearchParams параметры для поиска моделей
type SearchParams struct {
	Query        string
	Categories   []string
	Tags         []string
	Sort         string // "-likeCount", "-viewCount", "-publishedAt"
	Downloadable bool
	Animated     bool
	DateFilter   string // количество дней для фильтрации (например, "30" для последнего месяца)
}

// BuildSearchParams создает map параметров для API
func BuildSearchParams(sp SearchParams) map[string]string {
	params := make(map[string]string)

	if sp.Query != "" {
		params["q"] = sp.Query
	}

	if len(sp.Categories) > 0 {
		for _, cat := range sp.Categories {
			params["categories"] = cat
		}
	}

	if len(sp.Tags) > 0 {
		for _, tag := range sp.Tags {
			params["tags"] = tag
		}
	}

	if sp.Sort != "" {
		params["sort_by"] = sp.Sort
	}

	if sp.Downloadable {
		params["downloadable"] = "true"
	}

	if sp.Animated {
		params["animated"] = "true"
	}

	// Добавляем фильтр по дате для последнего месяца
	if sp.DateFilter != "" {
		params["date"] = sp.DateFilter
	}

	return params
}
