package secllm

import (
	"context"
	"encoding/json"

	"github.com/chaitanyabankanhal/ai-gateway/internal/inference"
)

// Detector wraps the generic inference client with prompt-injection classification.
type Detector struct {
	client    *inference.Client
	threshold float64
}

func NewDetector(client *inference.Client, threshold float64) *Detector {
	return &Detector{client: client, threshold: threshold}
}

// IsInjection returns (true, score, nil) when the classifier labels the text as
// INJECTION with confidence >= threshold. On sidecar errors it returns
// (false, 0, err) — callers decide whether to fail-open or fail-closed.
func (d *Detector) IsInjection(ctx context.Context, text string) (bool, float64, error) {
	res, err := d.client.Analyze(ctx, "prompt-injection", text)
	if err != nil {
		return false, 0, err
	}

	// wolf-defender response shape: [{"label":"INJECTION","score":0.99}]
	var items []struct {
		Label string  `json:"label"`
		Score float64 `json:"score"`
	}
	if err := json.Unmarshal(res.Raw, &items); err != nil || len(items) == 0 {
		return false, 0, err
	}

	top := items[0]
	return top.Label == "INJECTION" && top.Score >= d.threshold, top.Score, nil
}
