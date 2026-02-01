package service

import (
	"context"
	"time"

	"massrouter.ai/backend/internal/model"
)

type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (*model.User, error)
	Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error)
	Logout(ctx context.Context, userID string) error
	VerifyEmail(ctx context.Context, userID string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
}

type UserService interface {
	GetProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) error
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ListAPIKeys(ctx context.Context, userID string) ([]*APIKeyResponse, error)
	CreateAPIKey(ctx context.Context, userID string, req *CreateAPIKeyRequest) (*APIKeyResponse, error)
	DeleteAPIKey(ctx context.Context, userID, keyID string) error
	RotateAPIKey(ctx context.Context, userID, keyID string, req *RotateAPIKeyRequest) (*APIKeyResponse, error)
	GetUserBalance(ctx context.Context, userID string) (*UserBalance, error)
	GetUsageStatistics(ctx context.Context, userID string, startDate, endDate *time.Time) (*UsageStatistics, error)
}

type ModelService interface {
	ListModels(ctx context.Context, req *ListModelsRequest) (*ListModelsResponse, error)
	GetModelDetails(ctx context.Context, modelID string) (*ModelDetails, error)
	SearchModels(ctx context.Context, query string, filters *ModelFilters) (*ListModelsResponse, error)
	GetModelProviders(ctx context.Context) ([]*model.ModelProvider, error)
	GetModelCategories(ctx context.Context) ([]string, error)
}

type BillingService interface {
	GetBalance(ctx context.Context, userID string) (*BalanceInfo, error)
	GetPaymentHistory(ctx context.Context, userID string, page, limit int) (*PaymentHistoryResponse, error)
	CreatePayment(ctx context.Context, userID string, req *CreatePaymentRequest) (*PaymentInfo, error)
	ProcessPaymentWebhook(ctx context.Context, payload []byte, signature string) error
	GetBillingRecords(ctx context.Context, userID string, page, limit int) (*BillingRecordsResponse, error)
	CalculateCost(ctx context.Context, modelID string, inputTokens, outputTokens int) (*CostCalculation, error)
	CreateBillingRecord(ctx context.Context, req *CreateBillingRecordRequest) error
	StartBillingWorker()
	StopBillingWorker()
	GetQueueStatus(ctx context.Context) (*QueueStatus, error)
}

type AdminService interface {
	ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	GetUserDetails(ctx context.Context, userID string) (*AdminUserDetails, error)
	UpdateUser(ctx context.Context, userID string, req *AdminUpdateUserRequest) error
	CreateModelProvider(ctx context.Context, req *CreateModelProviderRequest) (*model.ModelProvider, error)
	UpdateModelProvider(ctx context.Context, providerID string, req *UpdateModelProviderRequest) error
	CreateModel(ctx context.Context, req *CreateModelRequest) (*model.Model, error)
	UpdateModel(ctx context.Context, modelID string, req *UpdateModelRequest) error
	GetSystemStats(ctx context.Context) (*SystemStats, error)
	UpdateSystemConfig(ctx context.Context, key, value string) error
}

type QuotaService interface {
	// Get user quota configuration
	GetUserQuota(ctx context.Context, userID string) (*model.UserQuota, error)
	UpdateUserQuota(ctx context.Context, userID string, req *UpdateQuotaRequest) error

	// Check if user has quota for an API call
	CheckQuota(ctx context.Context, userID string, tokens int, cost float64) (*QuotaCheckResult, error)

	// Record usage after successful API call
	RecordUsage(ctx context.Context, userID string, modelID string, tokens int, cost float64) error

	// Get usage statistics
	GetDailyUsage(ctx context.Context, userID string, date time.Time) (*model.UserUsage, error)
	GetMonthlyUsage(ctx context.Context, userID string, yearMonth string) (*model.MonthlyUsage, error)
	GetUsageHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]*model.UserUsage, error)

	// Reset quotas (admin only)
	ResetDailyQuota(ctx context.Context, userID string) error
	ResetMonthlyQuota(ctx context.Context, userID string) error
}

// Request/Response types
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50"`
}

type CreateAPIKeyRequest struct {
	Name          string               `json:"name" validate:"required,min=1,max=255"`
	Permissions   []string             `json:"permissions,omitempty" validate:"omitempty"` // Legacy format
	PermissionSet *model.PermissionSet `json:"permission_set,omitempty"`
	RateLimit     int                  `json:"rate_limit" validate:"required,min=1,max=10000"`
	ExpiresIn     int                  `json:"expires_in" validate:"omitempty,min=3600"` // seconds
}

type RotateAPIKeyRequest struct {
	Reason        string `json:"reason,omitempty"`
	KeepOldActive bool   `json:"keep_old_active,omitempty"`                          // Whether to keep old key active during grace period
	ExpiresIn     int    `json:"expires_in,omitempty" validate:"omitempty,min=3600"` // seconds, optional override
}

type UserProfile struct {
	User       *model.User         `json:"user"`
	APIKeys    []*model.UserAPIKey `json:"api_keys,omitempty"`
	Balance    float64             `json:"balance"`
	TotalUsage float64             `json:"total_usage"`
}

type UserBalance struct {
	Balance      float64    `json:"balance"`
	TotalPaid    float64    `json:"total_paid"`
	TotalUsed    float64    `json:"total_used"`
	LastPayment  *time.Time `json:"last_payment,omitempty"`
	LastActivity *time.Time `json:"last_activity,omitempty"`
}

type UsageStatistics struct {
	DailyUsage  []*DailyUsage `json:"daily_usage"`
	TotalCost   float64       `json:"total_cost"`
	TotalTokens int64         `json:"total_tokens"`
	TopModels   []*ModelUsage `json:"top_models"`
}

type DailyUsage struct {
	Date     time.Time `json:"date"`
	Cost     float64   `json:"cost"`
	Tokens   int64     `json:"tokens"`
	Requests int       `json:"requests"`
}

type ModelUsage struct {
	ModelID   string  `json:"model_id"`
	ModelName string  `json:"model_name"`
	Cost      float64 `json:"cost"`
	Tokens    int64   `json:"tokens"`
	Requests  int     `json:"requests"`
}

type ListModelsRequest struct {
	Page      int    `json:"page" validate:"min=1"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Category  string `json:"category,omitempty"`
	Provider  string `json:"provider,omitempty"`
	Search    string `json:"search,omitempty"`
	IsFree    *bool  `json:"is_free,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`    // name, price, created_at
	SortOrder string `json:"sort_order,omitempty"` // asc, desc
}

type ListModelsResponse struct {
	Models     []*ModelInfo `json:"models"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	TotalPages int          `json:"total_pages"`
}

type ModelInfo struct {
	*model.Model
	ProviderName string `json:"provider_name"`
}

type ModelDetails struct {
	*model.Model
	Provider      *model.ModelProvider `json:"provider"`
	DailyRequests int                  `json:"daily_requests"`
	SuccessRate   float64              `json:"success_rate"`
	AvgLatency    float64              `json:"avg_latency"`
}

type ModelFilters struct {
	Categories []string `json:"categories,omitempty"`
	Providers  []string `json:"providers,omitempty"`
	MinPrice   float64  `json:"min_price,omitempty"`
	MaxPrice   float64  `json:"max_price,omitempty"`
	IsFree     *bool    `json:"is_free,omitempty"`
}

type BalanceInfo struct {
	Balance     float64    `json:"balance"`
	CreditLimit float64    `json:"credit_limit,omitempty"`
	NextBilling *time.Time `json:"next_billing,omitempty"`
	IsOverdue   bool       `json:"is_overdue"`
}

type CreatePaymentRequest struct {
	Amount        float64 `json:"amount" validate:"required,min=0.01"`
	Currency      string  `json:"currency" validate:"required,len=3"`
	PaymentMethod string  `json:"payment_method" validate:"required"`
	ReturnURL     string  `json:"return_url,omitempty"`
}

type PaymentInfo struct {
	ID         string     `json:"id"`
	Amount     float64    `json:"amount"`
	Currency   string     `json:"currency"`
	Status     string     `json:"status"`
	PaymentURL string     `json:"payment_url,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

type PaymentHistoryResponse struct {
	Payments []*PaymentItem `json:"payments"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	Limit    int            `json:"limit"`
}

type PaymentItem struct {
	ID            string     `json:"id"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	PaymentMethod string     `json:"payment_method"`
	Status        string     `json:"status"`
	TransactionID string     `json:"transaction_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
}

type BillingRecordsResponse struct {
	Records []*BillingRecordItem `json:"records"`
	Total   int64                `json:"total"`
	Page    int                  `json:"page"`
	Limit   int                  `json:"limit"`
}

type BillingRecordItem struct {
	*model.BillingRecord
	ModelName    string `json:"model_name"`
	ProviderName string `json:"provider_name"`
}

type CostCalculation struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	InputCost    float64 `json:"input_cost"`
	OutputCost   float64 `json:"output_cost"`
	TotalCost    float64 `json:"total_cost"`
	ModelName    string  `json:"model_name"`
	ProviderName string  `json:"provider_name"`
}

type ListUsersRequest struct {
	Page      int    `json:"page" validate:"min=1"`
	Limit     int    `json:"limit" validate:"min=1,max=100"`
	Search    string `json:"search,omitempty"`
	Role      string `json:"role,omitempty"`
	Status    string `json:"status,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

type ListUsersResponse struct {
	Users      []*AdminUserInfo `json:"users"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

type AdminUserInfo struct {
	*model.User
	TotalPaid      float64    `json:"total_paid"`
	TotalUsed      float64    `json:"total_used"`
	CurrentBalance float64    `json:"current_balance"`
	APIKeysCount   int        `json:"api_keys_count"`
	LastActivity   *time.Time `json:"last_activity,omitempty"`
}

type AdminUserDetails struct {
	*model.User
	PaymentHistory []*PaymentItem       `json:"payment_history,omitempty"`
	BillingRecords []*BillingRecordItem `json:"billing_records,omitempty"`
	APIKeys        []*model.UserAPIKey  `json:"api_keys,omitempty"`
	OAuthAccounts  []*OAuthAccountInfo  `json:"oauth_accounts,omitempty"`
	Statistics     *UserStatistics      `json:"statistics,omitempty"`
}

type OAuthAccountInfo struct {
	ProviderName string     `json:"provider_name"`
	ConnectedAt  time.Time  `json:"connected_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
}

type UserStatistics struct {
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	AvgDailyCost  float64 `json:"avg_daily_cost"`
	MostUsedModel string  `json:"most_used_model"`
}

type AdminUpdateUserRequest struct {
	Role   string `json:"role,omitempty" validate:"omitempty,oneof=user admin"`
	Status string `json:"status,omitempty" validate:"omitempty,oneof=active suspended deleted"`
}

type CreateModelProviderRequest struct {
	Name       string                 `json:"name" validate:"required"`
	APIBaseURL string                 `json:"api_base_url" validate:"required,url"`
	APIKey     string                 `json:"api_key,omitempty"`
	Config     map[string]interface{} `json:"config" validate:"required"`
}

type UpdateModelProviderRequest struct {
	APIBaseURL string                 `json:"api_base_url,omitempty"`
	APIKey     string                 `json:"api_key,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Status     string                 `json:"status,omitempty" validate:"omitempty,oneof=active disabled"`
}

type CreateModelRequest struct {
	ProviderID    string                 `json:"provider_id" validate:"required"`
	Name          string                 `json:"name" validate:"required"`
	Description   string                 `json:"description,omitempty"`
	ContextLength *int                   `json:"context_length,omitempty"`
	MaxTokens     *int                   `json:"max_tokens,omitempty"`
	Capabilities  map[string]interface{} `json:"capabilities" validate:"required"`
	Category      string                 `json:"category,omitempty"`
	PricingTier   string                 `json:"pricing_tier,omitempty"`
	InputPrice    float64                `json:"input_price" validate:"required,min=0"`
	OutputPrice   float64                `json:"output_price" validate:"required,min=0"`
	IsFree        bool                   `json:"is_free"`
}

type UpdateModelRequest struct {
	Name          string                 `json:"name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	ContextLength *int                   `json:"context_length,omitempty"`
	MaxTokens     *int                   `json:"max_tokens,omitempty"`
	Capabilities  map[string]interface{} `json:"capabilities,omitempty"`
	Category      string                 `json:"category,omitempty"`
	PricingTier   string                 `json:"pricing_tier,omitempty"`
	InputPrice    float64                `json:"input_price,omitempty" validate:"omitempty,min=0"`
	OutputPrice   float64                `json:"output_price,omitempty" validate:"omitempty,min=0"`
	IsActive      *bool                  `json:"is_active,omitempty"`
}

type SystemStats struct {
	TotalUsers     int64          `json:"total_users"`
	ActiveUsers    int64          `json:"active_users"`
	TotalRequests  int64          `json:"total_requests"`
	TotalRevenue   float64        `json:"total_revenue"`
	DailyRequests  int64          `json:"daily_requests"`
	DailyRevenue   float64        `json:"daily_revenue"`
	TopModels      []*ModelStats  `json:"top_models"`
	RecentPayments []*PaymentItem `json:"recent_payments"`
	ServerStatus   *ServerStatus  `json:"server_status"`
}

type ModelStats struct {
	ModelID     string  `json:"model_id"`
	ModelName   string  `json:"model_name"`
	Requests    int64   `json:"requests"`
	Revenue     float64 `json:"revenue"`
	SuccessRate float64 `json:"success_rate"`
}

type ServerStatus struct {
	Database bool    `json:"database"`
	Redis    bool    `json:"redis"`
	Uptime   string  `json:"uptime"`
	Memory   float64 `json:"memory_usage"`
	CPU      float64 `json:"cpu_usage"`
}

// Quota-related request/response types
type UpdateQuotaRequest struct {
	DailyRequestLimit   *int                   `json:"daily_request_limit,omitempty" validate:"omitempty,min=0"`
	DailyTokenLimit     *int                   `json:"daily_token_limit,omitempty" validate:"omitempty,min=0"`
	DailyCostLimit      *float64               `json:"daily_cost_limit,omitempty" validate:"omitempty,min=0"`
	MonthlyRequestLimit *int                   `json:"monthly_request_limit,omitempty" validate:"omitempty,min=0"`
	MonthlyTokenLimit   *int                   `json:"monthly_token_limit,omitempty" validate:"omitempty,min=0"`
	MonthlyCostLimit    *float64               `json:"monthly_cost_limit,omitempty" validate:"omitempty,min=0"`
	PerMinuteRateLimit  *int                   `json:"per_minute_rate_limit,omitempty" validate:"omitempty,min=1,max=1000"`
	PerHourRateLimit    *int                   `json:"per_hour_rate_limit,omitempty" validate:"omitempty,min=1,max=10000"`
	ResetDay            *int                   `json:"reset_day,omitempty" validate:"omitempty,min=1,max=31"`
	Timezone            string                 `json:"timezone,omitempty"`
	ModelLimits         map[string]interface{} `json:"model_limits,omitempty"`
	IsActive            *bool                  `json:"is_active,omitempty"`
}

type QuotaCheckResult struct {
	Allowed              bool      `json:"allowed"`
	Reason               string    `json:"reason,omitempty"`
	DailyRequests        int       `json:"daily_requests"`
	DailyRequestsLimit   int       `json:"daily_requests_limit"`
	DailyTokens          int       `json:"daily_tokens"`
	DailyTokensLimit     int       `json:"daily_tokens_limit"`
	DailyCost            float64   `json:"daily_cost"`
	DailyCostLimit       float64   `json:"daily_cost_limit"`
	MonthlyRequests      int       `json:"monthly_requests"`
	MonthlyRequestsLimit int       `json:"monthly_requests_limit"`
	MonthlyTokens        int       `json:"monthly_tokens"`
	MonthlyTokensLimit   int       `json:"monthly_tokens_limit"`
	MonthlyCost          float64   `json:"monthly_cost"`
	MonthlyCostLimit     float64   `json:"monthly_cost_limit"`
	NextReset            time.Time `json:"next_reset"`
}

type CreateBillingRecordRequest struct {
	UserID         string                 `json:"user_id"`
	APIKeyID       *string                `json:"api_key_id,omitempty"`
	ModelID        string                 `json:"model_id"`
	RequestTokens  int                    `json:"request_tokens"`
	ResponseTokens int                    `json:"response_tokens"`
	TotalTokens    int                    `json:"total_tokens"`
	Cost           float64                `json:"cost"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type QueueStatus struct {
	QueueLength       int64     `json:"queue_length"`
	WorkerRunning     bool      `json:"worker_running"`
	LastProcessedAt   time.Time `json:"last_processed_at,omitempty"`
	TotalProcessed    int64     `json:"total_processed"`
	ErrorsLastHour    int       `json:"errors_last_hour"`
	AvgProcessingTime float64   `json:"avg_processing_time"`
}

// OAuth-related request/response types
type OAuthService interface {
	GetEnabledProviders(ctx context.Context) ([]*OAuthProviderInfo, error)
	StartOAuthFlow(ctx context.Context, req *StartOAuthFlowRequest) (string, error)
	HandleOAuthCallback(ctx context.Context, req *HandleOAuthCallbackRequest) (*OAuthLoginResponse, error)
	DisconnectOAuthAccount(ctx context.Context, userID string, req *DisconnectOAuthAccountRequest) error
}

type OAuthProviderInfo struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	ClientID    string                 `json:"client_id"`
	AuthURL     string                 `json:"auth_url"`
	TokenURL    string                 `json:"token_url"`
	UserInfoURL string                 `json:"user_info_url"`
	Scopes      string                 `json:"scopes"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
}

type StartOAuthFlowRequest struct {
	Provider    string `json:"provider" validate:"required"`
	CallbackURL string `json:"callback_url" validate:"required,url"`
	State       string `json:"state,omitempty"`
}

type HandleOAuthCallbackRequest struct {
	Provider string `json:"provider" validate:"required"`
	Code     string `json:"code" validate:"required"`
	State    string `json:"state" validate:"required"`
}

type DisconnectOAuthAccountRequest struct {
	Provider string `json:"provider" validate:"required"`
}

type OAuthLoginResponse struct {
	User         *model.User `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	IsNewUser    bool        `json:"is_new_user"`
}

// APIKeyResponse is the response structure for API keys that maintains backward compatibility
type APIKeyResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	APIKey      string     `json:"api_key,omitempty"`
	Prefix      string     `json:"prefix"`
	Permissions []string   `json:"permissions"` // Legacy format for backward compatibility
	RateLimit   int        `json:"rate_limit"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`

	// Rotation fields
	ParentKeyID    *string    `json:"parent_key_id,omitempty"`
	Version        int        `json:"version"`
	RotationReason *string    `json:"rotation_reason,omitempty"`
	RotatedAt      *time.Time `json:"rotated_at,omitempty"`
}
