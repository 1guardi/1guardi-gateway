package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
	"github.com/redis/go-redis/v9"
)

type ModelProvider interface {
	GetModels(ctx context.Context, provider, apiKey, baseURL string) ([]string, error)
}

type ModelProviderService struct {
	httpClient       *http.Client
	redis            *redis.Client
	openaiBaseURL    string
	geminiBaseURL    string
	anthropicBaseURL string
}

func NewModelProviderService(redis *redis.Client) *ModelProviderService {
	return &ModelProviderService{
		httpClient:       &http.Client{Timeout: 10 * time.Second},
		redis:            redis,
		openaiBaseURL:    "https://api.openai.com",
		geminiBaseURL:    "https://generativelanguage.googleapis.com",
		anthropicBaseURL: "https://api.anthropic.com",
	}
}

func (s *ModelProviderService) GetModels(ctx context.Context, provider, apiKey, baseURL string) ([]string, error) {
	// Cache by provider, baseURL and a hash of the API key
	hash := auth.HashKey(apiKey)
	cacheKey := fmt.Sprintf("models:cache:%s:%s:%s", provider, baseURL, hash)

	if s.redis != nil {
		val, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var models []string
			if err := json.Unmarshal([]byte(val), &models); err == nil {
				return models, nil
			}
		}
	}

	var models []string
	var err error

	switch provider {
	case "openai", "openai-compatible":
		models, err = s.fetchOpenAIModels(ctx, apiKey, baseURL)
	case "gemini":
		models, err = s.fetchGeminiModels(ctx, apiKey)
	case "anthropic":
		models, err = s.fetchAnthropicModels(ctx, apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err != nil {
		return nil, err
	}

	// Cache for 24 hours
	if s.redis != nil && len(models) > 0 {
		data, _ := json.Marshal(models)
		s.redis.Set(ctx, cacheKey, data, 24*time.Hour)
	}

	return models, nil
}

func (s *ModelProviderService) fetchOpenAIModels(ctx context.Context, apiKey, baseURL string) ([]string, error) {
	u := s.openaiBaseURL + "/v1/models"
	if baseURL != "" {
		baseURL = strings.TrimSuffix(baseURL, "/")
		if strings.HasSuffix(baseURL, "/v1/models") {
			u = baseURL
		} else if strings.HasSuffix(baseURL, "/v1") {
			u = fmt.Sprintf("%s/models", baseURL)
		} else {
			u = fmt.Sprintf("%s/v1/models", baseURL)
		}
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api error: %s", resp.Status)
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("provider returned non-json response: %s", string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}
	return models, nil
}

func (s *ModelProviderService) fetchGeminiModels(ctx context.Context, apiKey string) ([]string, error) {
	// Gemini list models endpoint
	url := fmt.Sprintf("%s/v1beta/models?key=%s", s.geminiBaseURL, apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini api error: %s", resp.Status)
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range result.Models {
		// Names look like "models/gemini-pro", we might want to trim "models/"
		models = append(models, m.Name)
	}
	return models, nil
}

func (s *ModelProviderService) fetchAnthropicModels(ctx context.Context, apiKey string) ([]string, error) {
	u := fmt.Sprintf("%s/v1/models", s.anthropicBaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic api error: %s", resp.Status)
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Data))
	for i, m := range result.Data {
		models[i] = m.ID
	}
	return models, nil
}
