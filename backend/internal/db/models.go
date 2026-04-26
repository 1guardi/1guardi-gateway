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
	Name       string `gorm:"not null"`
	TenantID   uint   `gorm:"not null;index"`
	Tenant     Tenant
	AgentID    *uint `gorm:"index"` // Optional: if set, key is scoped to this agent
	Agent      Agent `gorm:"foreignKey:AgentID"`
	LastUsedAt *time.Time
	IsActive   bool `gorm:"default:true"`
}

// Upstream represents an LLM provider endpoint.
type Upstream struct {
	gorm.Model
	KeyID         string `gorm:"not null;index" json:"key_id"`
	ProviderModel string `gorm:"not null" json:"model"`
	BaseURL       string `gorm:"not null" json:"base_url"`
	APIKey        string `gorm:"not null" json:"-"` // Never export API Key
	TenantID      uint   `gorm:"not null;index" json:"tenant_id"`
	Tenant        Tenant `gorm:"-" json:"-"`
}

// AutoMigrate runs schema migrations for all models.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tenant{},
		&Agent{},
		&APIKey{},
		&Upstream{},
	)
}
