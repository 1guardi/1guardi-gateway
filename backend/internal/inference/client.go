package inference

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client sends text to the mlrunner sidecar and caches verdicts in Redis.
type Client struct {
	http     *http.Client
	redis    *redis.Client
	baseURL  string
	cacheTTL time.Duration
}

func NewClient(baseURL string, timeoutMS int, cacheTTL time.Duration, redisClient *redis.Client) *Client {
	return &Client{
		http: &http.Client{
			Transport: &http.Transport{MaxIdleConnsPerHost: 10},
			Timeout:   time.Duration(timeoutMS) * time.Millisecond,
		},
		redis:    redisClient,
		baseURL:  baseURL,
		cacheTTL: cacheTTL,
	}
}

// Analyze calls the named analyzer on the mlrunner sidecar.
// Redis is checked first; on a miss the sidecar is called and the result cached.
func (c *Client) Analyze(ctx context.Context, analyzer, text string) (Result, error) {
	hash := sha256.Sum256([]byte(text))
	cacheKey := fmt.Sprintf("gw:inference:%s:%x", analyzer, hash)

	if c.redis != nil {
		if val, err := c.redis.Get(ctx, cacheKey).Bytes(); err == nil {
			return Result{Analyzer: analyzer, Raw: json.RawMessage(val), Cached: true}, nil
		}
	}

	reqBody, err := json.Marshal(analyzeRequest{Text: text})
	if err != nil {
		return Result{}, fmt.Errorf("inference: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/analyze/%s", c.baseURL, analyzer)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return Result{}, fmt.Errorf("inference: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("inference: sidecar call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("inference: sidecar returned %d", resp.StatusCode)
	}

	var ar analyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return Result{}, fmt.Errorf("inference: decode response: %w", err)
	}

	if c.redis != nil && len(ar.Result) > 0 {
		c.redis.Set(ctx, cacheKey, []byte(ar.Result), c.cacheTTL) //nolint:errcheck
	}

	return Result{Analyzer: analyzer, Raw: ar.Result, Cached: false}, nil
}
