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
