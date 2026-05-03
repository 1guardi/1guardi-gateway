package guardrails

// Scope constants for rule filtering.
const (
	ScopeInput    = "input"
	ScopeOutput   = "output"
	ScopeToolCall = "tool_call"
)

// Valid action values and their tiebreak priority (lower = higher precedence).
var actionPriority = map[string]int{
	"block":      0,
	"rewrite":    1,
	"tag":        2,
	"shadow":     3,
	"substitute": 4,
	"log":        5,
}

// Condition is the decoded form of GuardrailRule.Condition (stored as JSON).
type Condition struct {
	Type     string   `json:"type"`      // "regex" | "keyword" | "managed"
	Patterns []string `json:"patterns"`  // used by regex and keyword types
	MatchAll bool     `json:"match_all"` // AND semantics; default is OR
	RuleID   string   `json:"rule_id"`   // used by managed type
}

// EvalContext carries the content and identity info for a single evaluation.
type EvalContext struct {
	TenantID string
	AgentID  string
	Scope    string // one of the Scope constants
	Content  string
}

// FireEvent records a single rule that fired during evaluation.
type FireEvent struct {
	RuleID   uint
	RuleName string
	Action   string
	Priority int
	Reason   string
}

// Decision is the output of Evaluate.
type Decision struct {
	Allowed    bool
	Action     string       // winning action; empty when no rules fired
	FiredRules []FireEvent
}
