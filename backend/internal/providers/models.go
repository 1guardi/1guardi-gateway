package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chaitanyabankanhal/ai-gateway/internal/auth"
	"github.com/redis/go-redis/v9"
)

type ModelProviderService struct {
	httpClient *http.Client
	redis      *redis.Client
}

func NewModelProviderService(redis *redis.Client) *ModelProviderService {
	return &ModelProviderService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		redis:      redis,
	}
}

func (s *ModelProviderService) GetModels(ctx context.Context, provider, apiKey string) ([]string, error) {
	// Cache by provider and a hash of the API key
	hash := auth.HashKey(apiKey)
	cacheKey := fmt.Sprintf("models:cache:%s:%s", provider, hash)

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
	case "openai":
		models, err = s.fetchOpenAIModels(ctx, apiKey)
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

func (s *ModelProviderService) fetchOpenAIModels(ctx context.Context, apiKey string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.openai.com/v1/models", nil)
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
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
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
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.anthropic.com/v1/models", nil)
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
