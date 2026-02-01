package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"massrouter.ai/backend/internal/model"
	"massrouter.ai/backend/internal/repository"
	"massrouter.ai/backend/pkg/utils"
)

type userService struct {
	userRepo    repository.UserRepository
	apiKeyRepo  repository.UserAPIKeyRepository
	billingRepo repository.BillingRecordRepository
	paymentRepo repository.PaymentRecordRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	apiKeyRepo repository.UserAPIKeyRepository,
	billingRepo repository.BillingRecordRepository,
	paymentRepo repository.PaymentRecordRepository,
) UserService {
	return &userService{
		userRepo:    userRepo,
		apiKeyRepo:  apiKeyRepo,
		billingRepo: billingRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	apiKeys, err := s.apiKeyRepo.FindActiveKeysByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	balance, err := s.getUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	totalUsage, err := s.billingRepo.GetTotalCostByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total usage: %w", err)
	}

	return &UserProfile{
		User:       user,
		APIKeys:    apiKeys,
		Balance:    balance,
		TotalUsage: totalUsage,
	}, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if req.Username != "" && req.Username != user.Username {
		existingUser, err := s.userRepo.FindByUsername(ctx, req.Username)
		if err != nil {
			return fmt.Errorf("failed to check username: %w", err)
		}
		if existingUser != nil {
			return fmt.Errorf("username already taken")
		}
		user.Username = req.Username
	}

	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if !utils.VerifyPassword(oldPassword, user.PasswordHash) {
		return fmt.Errorf("current password is incorrect")
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *userService) ListAPIKeys(ctx context.Context, userID string) ([]*APIKeyResponse, error) {
	keys, err := s.apiKeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	responses := make([]*APIKeyResponse, 0, len(keys))
	for _, key := range keys {
		responses = append(responses, convertToAPIKeyResponse(key))
	}

	return responses, nil
}

func (s *userService) CreateAPIKey(ctx context.Context, userID string, req *CreateAPIKeyRequest) (*APIKeyResponse, error) {
	apiKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Determine permission set from request
	var permissionSet *model.PermissionSet

	if req.PermissionSet != nil {
		// Use the provided permission set
		permissionSet = req.PermissionSet
	} else if len(req.Permissions) > 0 {
		// Convert legacy permissions format to permission set
		permissionSet = convertLegacyPermissionsToSet(req.Permissions)
	} else {
		// Default to full access
		permissionSet = model.NewFullAccessPermissionSet()
	}

	// Validate the permission set
	if err := permissionSet.Validate(); err != nil {
		return nil, fmt.Errorf("invalid permission set: %w", err)
	}

	key := &model.UserAPIKey{
		UserID:      userID,
		Name:        req.Name,
		APIKey:      apiKey,
		Prefix:      apiKey[:10],
		Permissions: permissionSet.ToJSONB(),
		RateLimit:   req.RateLimit,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if req.ExpiresIn > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		key.ExpiresAt = &expiresAt
	}

	if err := s.apiKeyRepo.Create(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return convertToAPIKeyResponse(key), nil
}

// convertLegacyPermissionsToSet converts legacy permissions array to PermissionSet
func convertLegacyPermissionsToSet(permissions []string) *model.PermissionSet {
	// Map common permission names to operations
	operations := []string{}
	for _, perm := range permissions {
		switch perm {
		case "read":
			operations = append(operations, "read")
		case "write":
			operations = append(operations, "write")
		case "admin":
			// Admin gets all operations
			return model.NewFullAccessPermissionSet()
		}
	}

	// If no specific operations or all operations, create full access
	if len(operations) == 0 || containsAllOperations(operations) {
		return model.NewFullAccessPermissionSet()
	}

	// Create model-specific permission set with all models (*) and the specified operations
	return model.NewModelSpecificPermissionSet([]string{"*"}, operations)
}

// containsAllOperations checks if the operations list contains all necessary operations
func containsAllOperations(operations []string) bool {
	hasRead := false
	hasWrite := false
	for _, op := range operations {
		if op == "read" {
			hasRead = true
		}
		if op == "write" {
			hasWrite = true
		}
	}
	return hasRead && hasWrite
}

func (s *userService) DeleteAPIKey(ctx context.Context, userID, keyID string) error {
	key, err := s.apiKeyRepo.FindByID(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed to find API key: %w", err)
	}
	if key == nil {
		return fmt.Errorf("API key not found")
	}

	if key.UserID != userID {
		return fmt.Errorf("unauthorized to delete this API key")
	}

	if err := s.apiKeyRepo.RevokeKey(ctx, keyID); err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

func (s *userService) RotateAPIKey(ctx context.Context, userID, keyID string, req *RotateAPIKeyRequest) (*APIKeyResponse, error) {
	// Find the existing key
	oldKey, err := s.apiKeyRepo.FindByID(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find API key: %w", err)
	}
	if oldKey == nil {
		return nil, fmt.Errorf("API key not found")
	}

	// Verify ownership
	if oldKey.UserID != userID {
		return nil, fmt.Errorf("unauthorized to rotate this API key")
	}

	// Generate new API key
	newAPIKey, err := utils.GenerateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Determine expiration for new key
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(req.ExpiresIn) * time.Second)
		expiresAt = &exp
	} else if oldKey.ExpiresAt != nil {
		// Keep the same expiration if not specified
		expiresAt = oldKey.ExpiresAt
	}

	// Create new key with same permissions and settings
	newKey := &model.UserAPIKey{
		UserID:         userID,
		Name:           oldKey.Name + " (rotated)",
		APIKey:         newAPIKey,
		Prefix:         newAPIKey[:10],
		Permissions:    oldKey.Permissions,
		RateLimit:      oldKey.RateLimit,
		ExpiresAt:      expiresAt,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		ParentKeyID:    &oldKey.ID,
		Version:        oldKey.Version + 1,
		RotationReason: &req.Reason,
		RotatedAt:      &time.Time{},
	}

	// Set rotated_at to current time
	now := time.Now()
	newKey.RotatedAt = &now

	// Deactivate old key if not keeping it active
	if !req.KeepOldActive {
		oldKey.IsActive = false
		oldKey.UpdatedAt = time.Now()
		if err := s.apiKeyRepo.Update(ctx, oldKey); err != nil {
			return nil, fmt.Errorf("failed to deactivate old key: %w", err)
		}
	}

	// Create new key
	if err := s.apiKeyRepo.Create(ctx, newKey); err != nil {
		return nil, fmt.Errorf("failed to create rotated API key: %w", err)
	}

	return convertToAPIKeyResponse(newKey), nil
}

func (s *userService) GetUserBalance(ctx context.Context, userID string) (*UserBalance, error) {
	totalPaid, err := s.paymentRepo.GetUserTotalPaid(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total paid: %w", err)
	}

	totalUsed, err := s.billingRepo.GetTotalCostByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total used: %w", err)
	}

	balance := totalPaid - totalUsed

	var lastPayment *time.Time
	var lastActivity *time.Time

	payments, err := s.paymentRepo.FindByUserID(ctx, userID)
	if err == nil && len(payments) > 0 {
		lastPayment = &payments[0].CreatedAt
	}

	billingRecords, err := s.billingRepo.FindByUserID(ctx, userID)
	if err == nil && len(billingRecords) > 0 {
		lastActivity = &billingRecords[0].CreatedAt
	}

	return &UserBalance{
		Balance:      balance,
		TotalPaid:    totalPaid,
		TotalUsed:    totalUsed,
		LastPayment:  lastPayment,
		LastActivity: lastActivity,
	}, nil
}

func (s *userService) GetUsageStatistics(ctx context.Context, userID string, startDate, endDate *time.Time) (*UsageStatistics, error) {
	records, err := s.billingRepo.GetUserUsage(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	usageMap := make(map[time.Time]*DailyUsage)
	modelUsageMap := make(map[string]*ModelUsage)
	var totalCost float64
	var totalTokens int64

	for _, record := range records {
		date := record.CreatedAt.Truncate(24 * time.Hour)
		if daily, exists := usageMap[date]; exists {
			daily.Cost += record.Cost
			daily.Tokens += int64(record.TotalTokens)
			daily.Requests++
		} else {
			usageMap[date] = &DailyUsage{
				Date:     date,
				Cost:     record.Cost,
				Tokens:   int64(record.TotalTokens),
				Requests: 1,
			}
		}

		modelKey := record.ModelID
		if modelUsage, exists := modelUsageMap[modelKey]; exists {
			modelUsage.Cost += record.Cost
			modelUsage.Tokens += int64(record.TotalTokens)
			modelUsage.Requests++
		} else {
			modelUsageMap[modelKey] = &ModelUsage{
				ModelID:   record.ModelID,
				ModelName: "Unknown",
				Cost:      record.Cost,
				Tokens:    int64(record.TotalTokens),
				Requests:  1,
			}
		}

		totalCost += record.Cost
		totalTokens += int64(record.TotalTokens)
	}

	dailyUsage := make([]*DailyUsage, 0, len(usageMap))
	for _, daily := range usageMap {
		dailyUsage = append(dailyUsage, daily)
	}

	topModels := make([]*ModelUsage, 0, len(modelUsageMap))
	for _, modelUsage := range modelUsageMap {
		topModels = append(topModels, modelUsage)
	}

	return &UsageStatistics{
		DailyUsage:  dailyUsage,
		TotalCost:   totalCost,
		TotalTokens: totalTokens,
		TopModels:   topModels[:min(5, len(topModels))],
	}, nil
}

func (s *userService) getUserBalance(ctx context.Context, userID string) (float64, error) {
	balance, err := s.GetUserBalance(ctx, userID)
	if err != nil {
		return 0, err
	}
	return balance.Balance, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// convertToAPIKeyResponse converts a UserAPIKey to APIKeyResponse for backward compatibility
func convertToAPIKeyResponse(key *model.UserAPIKey) *APIKeyResponse {
	// Extract permissions from PermissionSet
	var permissions []string

	// Try to parse as PermissionSet
	permissionSet, err := model.PermissionSetFromJSONB(key.Permissions)
	if err == nil && permissionSet != nil {
		// Convert PermissionSet to legacy permissions format
		permissions = convertPermissionSetToLegacy(permissionSet)
	} else {
		// Fallback to old format
		if perms, ok := key.Permissions["permissions"].([]interface{}); ok {
			for _, p := range perms {
				if str, ok := p.(string); ok {
					permissions = append(permissions, str)
				}
			}
		}
	}

	// If no permissions found, default to empty array
	if permissions == nil {
		permissions = []string{}
	}

	return &APIKeyResponse{
		ID:             key.ID,
		UserID:         key.UserID,
		Name:           key.Name,
		APIKey:         key.APIKey,
		Prefix:         key.Prefix,
		Permissions:    permissions,
		RateLimit:      key.RateLimit,
		ExpiresAt:      key.ExpiresAt,
		LastUsedAt:     key.LastUsedAt,
		IsActive:       key.IsActive,
		CreatedAt:      key.CreatedAt,
		ParentKeyID:    key.ParentKeyID,
		Version:        key.Version,
		RotationReason: key.RotationReason,
		RotatedAt:      key.RotatedAt,
	}
}

// convertPermissionSetToLegacy converts a PermissionSet to legacy permissions array
func convertPermissionSetToLegacy(ps *model.PermissionSet) []string {
	// Check if it's full access
	if ps.DefaultAllow && len(ps.Permissions) == 0 {
		return []string{"read", "write", "admin"}
	}

	// Extract unique actions from permissions
	actions := make(map[string]bool)
	for _, perm := range ps.Permissions {
		if perm.ResourceType == "model" && (perm.ResourceID == "*" || perm.ResourceID != "") {
			if perm.Action == "*" {
				// Wildcard action means all operations
				return []string{"read", "write", "admin"}
			}
			actions[perm.Action] = true
		}
	}

	// Convert actions to array
	var result []string
	for action := range actions {
		result = append(result, action)
	}

	// If no actions found but has deprecated fields, use them
	if len(result) == 0 && len(ps.AllowedOperations) > 0 {
		return ps.AllowedOperations
	}

	// Sort for consistency
	sort.Strings(result)
	return result
}
