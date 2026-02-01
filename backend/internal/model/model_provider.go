package model

import (
	"time"
)

type ModelProvider struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name       string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	APIBaseURL string    `gorm:"type:varchar(512);not null" json:"api_base_url"`
	APIKey     string    `gorm:"type:varchar(512)" json:"-"`
	Config     JSONB     `gorm:"type:jsonb;not null;default:'{}'" json:"config"`
	Status     string    `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time `gorm:"not null" json:"updated_at"`

	Models []Model `gorm:"foreignKey:ProviderID" json:"models,omitempty"`
}

type Model struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ProviderID    string    `gorm:"type:uuid;not null;index" json:"provider_id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description,omitempty"`
	ContextLength *int      `gorm:"type:integer" json:"context_length,omitempty"`
	MaxTokens     *int      `gorm:"type:integer" json:"max_tokens,omitempty"`
	Capabilities  JSONB     `gorm:"type:jsonb;not null;default:'{}'" json:"capabilities"`
	Category      string    `gorm:"type:varchar(100);index" json:"category,omitempty"`
	PricingTier   string    `gorm:"type:varchar(50)" json:"pricing_tier,omitempty"`
	InputPrice    float64   `gorm:"type:decimal(10,8);not null;default:0" json:"input_price"`
	OutputPrice   float64   `gorm:"type:decimal(10,8);not null;default:0" json:"output_price"`
	IsFree        bool      `gorm:"not null;default:false" json:"is_free"`
	IsActive      bool      `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`

	Provider        ModelProvider    `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	BillingRecords  []BillingRecord  `gorm:"foreignKey:ModelID" json:"billing_records,omitempty"`
	ModelStatistics []ModelStatistic `gorm:"foreignKey:ModelID" json:"statistics,omitempty"`
}

type UserAPIKey struct {
	ID          string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID      string     `gorm:"type:uuid;not null;index" json:"user_id"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	APIKey      string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	Prefix      string     `gorm:"type:varchar(10);not null" json:"prefix"`
	Permissions JSONB      `gorm:"type:jsonb;not null;default:'[]'" json:"permissions"`
	RateLimit   int        `gorm:"not null;default:1000" json:"rate_limit"`
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	IsActive    bool       `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt   time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updated_at"`

	// Rotation fields
	ParentKeyID    *string    `gorm:"type:uuid;index" json:"parent_key_id,omitempty"`
	Version        int        `gorm:"not null;default:1" json:"version"`
	RotationReason *string    `gorm:"type:text" json:"rotation_reason,omitempty"`
	RotatedAt      *time.Time `json:"rotated_at,omitempty"`

	User           User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BillingRecords []BillingRecord `gorm:"foreignKey:APIKeyID" json:"billing_records,omitempty"`
}

func (ModelProvider) TableName() string {
	return "model_providers"
}

func (Model) TableName() string {
	return "models"
}

func (UserAPIKey) TableName() string {
	return "user_api_keys"
}
