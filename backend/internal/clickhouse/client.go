package clickhouse

import (
	"context"
	"fmt"
	"time"

	clickhousego "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Client wraps the native ClickHouse driver for analytics queries.
// Writes are handled by the OTel Collector's clickhouse exporter.
type Client struct {
	conn driver.Conn
}

func NewClient(addr, user, password, database string) (*Client, error) {
	conn, err := clickhousego.Open(&clickhousego.Options{
		Addr: []string{addr},
		Auth: clickhousego.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
		Compression: &clickhousego.Compression{
			Method: clickhousego.CompressionLZ4,
		},
		DialTimeout:  10 * time.Second,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse: open: %w", err)
	}
	return &Client{conn: conn}, nil
}

// RuleFire holds the fires-in-24h count for one rule.
type RuleFire struct {
	RuleID   string `ch:"rule_id"`
	Fires24h uint32 `ch:"fires_24h"`
}

// GuardrailFireCounts returns the number of times each rule fired in the last 24 h
// for the given tenant.
func (c *Client) GuardrailFireCounts(ctx context.Context, tenantID string) ([]RuleFire, error) {
	var rows []RuleFire
	err := c.conn.Select(ctx, &rows, `
		SELECT
			LogAttributes['guardrail.rule_id'] AS rule_id,
			toUInt32(count())                  AS fires_24h
		FROM otel_logs
		WHERE ServiceName = 'ai-gateway'
		  AND Body         = 'guardrail.fired'
		  AND LogAttributes['tenant_id'] = ?
		  AND Timestamp   >= now() - INTERVAL 24 HOUR
		GROUP BY rule_id
	`, tenantID)
	return rows, err
}

// GuardrailEvent is one row from the audit log.
type GuardrailEvent struct {
	Timestamp string `ch:"timestamp"`
	TraceID   string `ch:"trace_id"`
	RuleID    string `ch:"rule_id"`
	RuleName  string `ch:"rule_name"`
	Action    string `ch:"action"`
	Reason    string `ch:"reason"`
	Scope     string `ch:"scope"`
	AgentID   string `ch:"agent_id"`
}

// TraceRow holds the summary of one LLM call (one llm.generation span).
type TraceRow struct {
	TraceID      string  `ch:"trace_id"`
	Timestamp    string  `ch:"ts"`
	Model        string  `ch:"model"`
	InputTokens  int32   `ch:"input_tokens"`
	OutputTokens int32   `ch:"output_tokens"`
	Cost         float64 `ch:"cost"`
	TtftMs       float64 `ch:"ttft_ms"`
	Tps          float64 `ch:"tps"`
	DurationMs   float64 `ch:"duration_ms"`
	Status       string  `ch:"status"`
	AgentID      string  `ch:"agent_id"`
	ThreadID     string  `ch:"thread_id"`
}

// TraceSpan holds one OTel span within a trace.
type TraceSpan struct {
	SpanID       string            `ch:"span_id"`
	ParentSpanID string            `ch:"parent_span_id"`
	SpanName     string            `ch:"span_name"`
	DurationMs   float64           `ch:"duration_ms"`
	StartTimeMs  float64           `ch:"start_time_ms"`
	StatusCode   string            `ch:"status_code"`
	Attributes   map[string]string `ch:"attributes"`
}

// ListTraces returns the most recent llm.generation spans for a tenant.
// Pass agentID = "" to fetch across all agents.
func (c *Client) ListTraces(ctx context.Context, tenantID, agentID string, limit int) ([]TraceRow, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	agentFilter := ""
	if agentID != "" {
		agentFilter = "AND llm.SpanAttributes['agent.id'] = ?"
	}

	sql := fmt.Sprintf(`
		SELECT
			llm.TraceId                                                     AS trace_id,
			formatDateTime(llm.Timestamp, '%%Y-%%m-%%dT%%H:%%i:%%SZ')      AS ts,
			llm.SpanAttributes['gen_ai.model']                              AS model,
			toInt32OrZero(llm.SpanAttributes['gen_ai.input.tokens'])        AS input_tokens,
			toInt32OrZero(llm.SpanAttributes['gen_ai.output.tokens'])       AS output_tokens,
			toFloat64OrZero(llm.SpanAttributes['gen_ai.cost.usd'])          AS cost,
			toFloat64OrZero(llm.SpanAttributes['gen_ai.ttft_ms'])           AS ttft_ms,
			toFloat64OrZero(llm.SpanAttributes['gen_ai.tps'])               AS tps,
			toFloat64(llm.Duration) / 1e6                                    AS duration_ms,
			if(llm.StatusCode = 'STATUS_CODE_ERROR', 'ERROR',
			   if(isNotNull(gr.TraceId), 'GUARDRAIL', 'OK')
			) AS status,
			llm.SpanAttributes['agent.id']                                  AS agent_id,
			llm.SpanAttributes['thread.id']                                 AS thread_id
		FROM otel_traces llm
		LEFT JOIN (
			SELECT DISTINCT TraceId
			FROM otel_logs
			WHERE ServiceName = 'ai-gateway'
			  AND Body = 'guardrail.fired'
		) gr ON llm.TraceId = gr.TraceId
		WHERE llm.ServiceName = 'ai-gateway'
		  AND llm.SpanName = 'llm.generation'
		  AND llm.SpanAttributes['tenant.id'] = ?
		  %s
		ORDER BY llm.Timestamp DESC
		LIMIT ?
	`, agentFilter)

	var rows []TraceRow
	var err error
	if agentID != "" {
		err = c.conn.Select(ctx, &rows, sql, tenantID, agentID, limit)
	} else {
		err = c.conn.Select(ctx, &rows, sql, tenantID, limit)
	}
	return rows, err
}

// GetTraceSpans returns all spans for a given trace, ordered by start time.
func (c *Client) GetTraceSpans(ctx context.Context, traceID string) ([]TraceSpan, error) {
	var rows []TraceSpan
	err := c.conn.Select(ctx, &rows, `
		SELECT
			SpanId                                      AS span_id,
			ParentSpanId                                AS parent_span_id,
			SpanName                                    AS span_name,
			toFloat64(Duration) / 1e6                   AS duration_ms,
			toFloat64(toInt64(Timestamp)) / 1e6         AS start_time_ms,
			StatusCode                                  AS status_code,
			SpanAttributes                              AS attributes
		FROM otel_traces
		WHERE TraceId = ?
		ORDER BY Timestamp ASC
	`, traceID)
	return rows, err
}

// GuardrailEvents returns the most recent audit events for a tenant, optionally
// filtered to a single rule. Pass ruleID = "" to fetch across all rules.
func (c *Client) GuardrailEvents(ctx context.Context, tenantID, ruleID string, limit int) ([]GuardrailEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	ruleFilter := "1=1"
	if ruleID != "" {
		ruleFilter = "LogAttributes['guardrail.rule_id'] = ?"
	}

	sql := fmt.Sprintf(`
		SELECT
			formatDateTime(Timestamp, '%%Y-%%m-%%dT%%H:%%i:%%SZ') AS timestamp,
			TraceId                                AS trace_id,
			LogAttributes['guardrail.rule_id']     AS rule_id,
			LogAttributes['guardrail.rule_name']   AS rule_name,
			LogAttributes['guardrail.action']      AS action,
			LogAttributes['guardrail.reason']      AS reason,
			LogAttributes['scope']                 AS scope,
			LogAttributes['agent_id']              AS agent_id
		FROM otel_logs
		WHERE ServiceName = 'ai-gateway'
		  AND Body         = 'guardrail.fired'
		  AND LogAttributes['tenant_id'] = ?
		  AND %s
		ORDER BY Timestamp DESC
		LIMIT ?
	`, ruleFilter)

	var rows []GuardrailEvent
	var err error
	if ruleID != "" {
		err = c.conn.Select(ctx, &rows, sql, tenantID, ruleID, limit)
	} else {
		err = c.conn.Select(ctx, &rows, sql, tenantID, limit)
	}
	return rows, err
}
