package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Permission represents a granular permission for an API key
type Permission struct {
	// ResourceType indicates the type of resource (e.g., "model", "billing", "key")
	ResourceType string `json:"resource_type"`

	// ResourceID is the specific resource identifier (e.g., model ID, "*" for all)
	ResourceID string `json:"resource_id"`

	// Action is the allowed operation (e.g., "read", "write", "create", "delete")
	Action string `json:"action"`

	// Constraints are optional constraints for the permission (e.g., max_tokens_per_month)
	Constraints map[string]interface{} `json:"constraints,omitempty"`
}

// PermissionSet represents a collection of permissions for an API key
type PermissionSet struct {
	Permissions []Permission `json:"permissions"`

	// DefaultPermissions are applied if no specific permission matches
	DefaultAllow bool `json:"default_allow"`

	// MaxRequestsPerMonth limits total requests
	MaxRequestsPerMonth *int64 `json:"max_requests_per_month,omitempty"`

	// MaxTokensPerMonth limits total tokens
	MaxTokensPerMonth *int64 `json:"max_tokens_per_month,omitempty"`

	// AllowedModels is a list of model IDs that can be accessed (deprecated, use Permissions instead)
	AllowedModels []string `json:"allowed_models,omitempty"`

	// AllowedOperations is a list of operations that can be performed (deprecated, use Permissions instead)
	AllowedOperations []string `json:"allowed_operations,omitempty"`
}

// NewFullAccessPermissionSet creates a permission set with full access to all resources
func NewFullAccessPermissionSet() *PermissionSet {
	return &PermissionSet{
		Permissions: []Permission{
			{
				ResourceType: "model",
				ResourceID:   "*",
				Action:       "*",
			},
			{
				ResourceType: "billing",
				ResourceID:   "*",
				Action:       "read",
			},
		},
		DefaultAllow:      true,
		AllowedModels:     []string{"*"},
		AllowedOperations: []string{"*"},
	}
}

// NewModelSpecificPermissionSet creates a permission set for specific models
func NewModelSpecificPermissionSet(modelIDs []string, operations []string) *PermissionSet {
	permissions := make([]Permission, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		for _, op := range operations {
			permissions = append(permissions, Permission{
				ResourceType: "model",
				ResourceID:   modelID,
				Action:       op,
			})
		}
	}

	return &PermissionSet{
		Permissions:       permissions,
		DefaultAllow:      false,
		AllowedModels:     modelIDs,
		AllowedOperations: operations,
	}
}

// HasPermission checks if the permission set allows the requested action on the resource
func (ps *PermissionSet) HasPermission(resourceType, resourceID, action string) bool {
	// First check specific permissions
	for _, perm := range ps.Permissions {
		if perm.ResourceType == resourceType || perm.ResourceType == "*" {
			if perm.ResourceID == resourceID || perm.ResourceID == "*" {
				if perm.Action == action || perm.Action == "*" {
					return true
				}
			}
		}
	}

	// Fallback to deprecated fields for backward compatibility
	if resourceType == "model" {
		// Check if model is in allowed models
		modelAllowed := false
		for _, allowed := range ps.AllowedModels {
			if allowed == resourceID || allowed == "*" {
				modelAllowed = true
				break
			}
		}

		// Check if operation is in allowed operations
		operationAllowed := false
		for _, allowed := range ps.AllowedOperations {
			if allowed == action || allowed == "*" {
				operationAllowed = true
				break
			}
		}

		if modelAllowed && operationAllowed {
			return true
		}
	}

	// Finally check default allow
	return ps.DefaultAllow
}

// CheckModelAccess checks if a specific model can be accessed with the given operation
func (ps *PermissionSet) CheckModelAccess(modelID, operation string) bool {
	return ps.HasPermission("model", modelID, operation)
}

// ToJSONB converts PermissionSet to JSONB for database storage
func (ps *PermissionSet) ToJSONB() JSONB {
	data, _ := json.Marshal(ps)
	var result JSONB
	json.Unmarshal(data, &result)
	return result
}

// PermissionSetFromJSONB creates a PermissionSet from JSONB
func PermissionSetFromJSONB(j JSONB) (*PermissionSet, error) {
	if j == nil {
		return NewFullAccessPermissionSet(), nil
	}

	data, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	var ps PermissionSet
	if err := json.Unmarshal(data, &ps); err != nil {
		// Try to parse legacy format
		return parseLegacyPermissions(j)
	}

	return &ps, nil
}

// parseLegacyPermissions parses the old permissions format
func parseLegacyPermissions(j JSONB) (*PermissionSet, error) {
	ps := NewFullAccessPermissionSet()

	// Old format: {"permissions": ["read", "write"]}
	if perms, ok := j["permissions"].([]interface{}); ok {
		operations := make([]string, 0, len(perms))
		for _, p := range perms {
			if op, ok := p.(string); ok {
				operations = append(operations, op)
			}
		}
		ps.AllowedOperations = operations
	}

	// Check for allowed_models in old format
	if models, ok := j["allowed_models"].([]interface{}); ok {
		modelIDs := make([]string, 0, len(models))
		for _, m := range models {
			if id, ok := m.(string); ok {
				modelIDs = append(modelIDs, id)
			}
		}
		ps.AllowedModels = modelIDs
	}

	// Convert old format to new permissions
	ps.convertLegacyToPermissions()

	return ps, nil
}

// convertLegacyToPermissions converts legacy fields to the new permissions structure
func (ps *PermissionSet) convertLegacyToPermissions() {
	if len(ps.Permissions) > 0 {
		// Already has new format permissions
		return
	}

	// Convert allowed models and operations to permissions
	for _, modelID := range ps.AllowedModels {
		for _, operation := range ps.AllowedOperations {
			ps.Permissions = append(ps.Permissions, Permission{
				ResourceType: "model",
				ResourceID:   modelID,
				Action:       operation,
			})
		}
	}
}

// Validate checks if the permission set is valid
func (ps *PermissionSet) Validate() error {
	for _, perm := range ps.Permissions {
		if strings.TrimSpace(perm.ResourceType) == "" {
			return fmt.Errorf("permission resource_type cannot be empty")
		}
		if strings.TrimSpace(perm.ResourceID) == "" {
			return fmt.Errorf("permission resource_id cannot be empty")
		}
		if strings.TrimSpace(perm.Action) == "" {
			return fmt.Errorf("permission action cannot be empty")
		}
	}
	return nil
}
