package db

import (
	"time"

	"gorm.io/gorm"
)

// Tenant represents an LLM consumer or project.
type Tenant struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	APIKey      string `gorm:"uniqueIndex;not null"` // Internal key for this tenant to use the gateway
	Agents      []Agent
	APIKeys     []APIKey
	Upstreams   []Upstream
}

// Agent represents a specific AI agent within a tenant.
type Agent struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	TenantID    uint `gorm:"not null;index"`
	Tenant      Tenant
	APIKeys     []APIKey
}

// APIKey represents a key used to authenticate requests to the gateway.
type APIKey struct {
	gorm.Model
	KeyHash    string `gorm:"uniqueIndex;not null"`
	Prefix     string `gorm:"not null"` // e.g. "sk_"
	Suffix     string `gorm:"not null"` // Last 4 characters of the key
	Name       string `gorm:"not null"`
	TenantID   uint   `gorm:"not null;index"`
	Tenant     Tenant
	AgentID    *uint `gorm:"index"` // Optional: if set, key is scoped to this agent
	Agent      Agent `gorm:"foreignKey:AgentID"`
	UserID     *uint `gorm:"index"` // Optional: if set, key is scoped to this user
	User       User  `gorm:"foreignKey:UserID"`
	LastUsedAt *time.Time
	IsActive   bool `gorm:"default:true"`
}

// Upstream represents an LLM provider endpoint.
type Upstream struct {
	gorm.Model
	KeyID    string `gorm:"not null;index" json:"key_id"`
	Provider string `gorm:"not null;default:'openai'" json:"provider"`
	Models   string `gorm:"not null" json:"models"` // Comma-separated list of models
	BaseURL  string `gorm:"not null" json:"base_url"`
	APIKey   string `gorm:"not null" json:"-"` // Never export API Key
	TenantID uint   `gorm:"not null;index" json:"tenant_id"`
	Tenant   Tenant `gorm:"-" json:"-"`
}

// User represents a user who can log in to the gateway.
type User struct {
	gorm.Model
	Name         string `gorm:"not null;default:''"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	IsSuperAdmin bool   `gorm:"default:false"`
}

// Role represents a collection of permissions.
type Role struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

// Permission represents an atomic action that can be performed.
type Permission struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null"` // e.g., "tenant.read"
}

// TenantMember maps a user to a tenant with a specific role.
type TenantMember struct {
	gorm.Model
	UserID   uint `gorm:"uniqueIndex:idx_user_tenant;not null"`
	User     User
	TenantID uint `gorm:"uniqueIndex:idx_user_tenant;not null"`
	Tenant   Tenant
	RoleID   uint `gorm:"not null"`
	Role     Role
}

// GuardrailRule defines a policy evaluated against inbound/outbound LLM content.
type GuardrailRule struct {
	gorm.Model
	TenantID  uint   `gorm:"not null;index" json:"tenant_id"`
	AgentID   *uint  `gorm:"index" json:"agent_id"`                          // nil = applies to all agents
	Name      string `gorm:"not null" json:"name"`
	Priority  int    `gorm:"not null;default:100" json:"priority"`
	Scope     string `gorm:"not null" json:"scope"`                          // CSV: "input", "output", "tool_call"
	Direction string `gorm:"not null;default:'both'" json:"direction"`       // "inbound"|"outbound"|"both"
	Condition string `gorm:"type:text" json:"condition"`                     // JSON-encoded Condition struct
	Action    string `gorm:"not null" json:"action"`                         // "block"|"log"|"tag"|"rewrite"|"shadow"|"substitute"
	Mode      string `gorm:"not null;default:'parallel'" json:"mode"`
	Managed   bool   `gorm:"default:false" json:"managed"`
	ManagedID string `json:"managed_id"`                                     // e.g., "prompt-injection"
	Version   string `json:"version"`
	Enabled   bool   `gorm:"default:true" json:"enabled"`
}

// AutoMigrate runs schema migrations for all models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tenant{},
		&Agent{},
		&APIKey{},
		&Upstream{},
		&User{},
		&Role{},
		&Permission{},
		&TenantMember{},
		&GuardrailRule{},
	)
}
