package user

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"massrouter.ai/backend/internal/service"
)

type Controller struct {
	userService    service.UserService
	authService    service.AuthService
	billingService service.BillingService
	validator      *validator.Validate
}

func NewController(userService service.UserService, authService service.AuthService, billingService service.BillingService) *Controller {
	return &Controller{
		userService:    userService,
		authService:    authService,
		billingService: billingService,
		validator:      validator.New(),
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile information
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Profile retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/profile [get]
func (c *Controller) GetProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	profile, err := c.userService.GetProfile(ctx.Request.Context(), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get profile",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current user's profile information
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.UpdateProfileRequest true "Profile update request"
// @Success 200 {object} map[string]interface{} "Profile updated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 409 {object} map[string]interface{} "Username already taken"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/profile [put]
func (c *Controller) UpdateProfile(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	var req service.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	if err := c.validator.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_VALIDATION",
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	if err := c.userService.UpdateProfile(ctx.Request.Context(), userID.(string), &req); err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		if err.Error() == "user not found" {
			status = http.StatusNotFound
			errorCode = "ERR_USER_NOT_FOUND"
		} else if err.Error() == "username already taken" {
			status = http.StatusConflict
			errorCode = "ERR_USERNAME_EXISTS"
		}

		ctx.JSON(status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Profile updated successfully",
		},
	})
}

// ChangePassword godoc
// @Summary Change password
// @Description Change current user's password
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]string true "Password change request"
// @Success 200 {object} map[string]interface{} "Password changed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input or incorrect current password"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/password/change [post]
func (c *Controller) ChangePassword(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,min=8"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	if err := c.validator.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_VALIDATION",
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	if err := c.userService.ChangePassword(ctx.Request.Context(), userID.(string), req.CurrentPassword, req.NewPassword); err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		switch err.Error() {
		case "user not found":
			status = http.StatusNotFound
			errorCode = "ERR_USER_NOT_FOUND"
		case "current password is incorrect":
			status = http.StatusBadRequest
			errorCode = "ERR_INVALID_PASSWORD"
		}

		ctx.JSON(status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password changed successfully",
		},
	})
}

// ListAPIKeys godoc
// @Summary List API keys
// @Description Get list of current user's API keys
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "API keys retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/api-keys [get]
func (c *Controller) ListAPIKeys(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	keys, err := c.userService.ListAPIKeys(ctx.Request.Context(), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to list API keys",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"api_keys": keys,
		},
	})
}

// CreateAPIKey godoc
// @Summary Create API key
// @Description Create a new API key for current user
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateAPIKeyRequest true "API key creation request"
// @Success 201 {object} map[string]interface{} "API key created successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/api-keys [post]
func (c *Controller) CreateAPIKey(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	var req service.CreateAPIKeyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	if err := c.validator.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_VALIDATION",
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	key, err := c.userService.CreateAPIKey(ctx.Request.Context(), userID.(string), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to create API key",
			},
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"api_key":  key.APIKey, // The actual API key string
			"key_info": key,        // The full key information
			"message":  "API key created successfully. Save this key now as it won't be shown again.",
			"warning":  "For security reasons, the full API key will only be shown once.",
		},
	})
}

// DeleteAPIKey godoc
// @Summary Delete API key
// @Description Delete an API key by ID
// @Tags user
// @Produce json
// @Security BearerAuth
// @Param id path string true "API key ID"
// @Success 200 {object} map[string]interface{} "API key deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - missing ID"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - unauthorized to delete this key"
// @Failure 404 {object} map[string]interface{} "API key not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/api-keys/{id} [delete]
func (c *Controller) DeleteAPIKey(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	keyID := ctx.Param("id")
	if keyID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "API key ID is required",
			},
		})
		return
	}

	if err := c.userService.DeleteAPIKey(ctx.Request.Context(), userID.(string), keyID); err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		switch err.Error() {
		case "API key not found":
			status = http.StatusNotFound
			errorCode = "ERR_KEY_NOT_FOUND"
		case "unauthorized to delete this API key":
			status = http.StatusForbidden
			errorCode = "ERR_UNAUTHORIZED"
		}

		ctx.JSON(status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "API key deleted successfully",
		},
	})
}

// RotateAPIKey godoc
// @Summary Rotate API key
// @Description Rotate an existing API key by creating a new version and optionally deactivating the old one
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "API key ID"
// @Param request body service.RotateAPIKeyRequest true "Rotation request"
// @Success 201 {object} map[string]interface{} "API key rotated successfully"
// @Failure 400 {object} map[string]interface{} "Bad request - invalid input"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden - unauthorized to rotate this key"
// @Failure 404 {object} map[string]interface{} "API key not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/api-keys/{id}/rotate [post]
func (c *Controller) RotateAPIKey(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	keyID := ctx.Param("id")
	if keyID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "API key ID is required",
			},
		})
		return
	}

	var req service.RotateAPIKeyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_BAD_REQUEST",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate request if needed
	if err := c.validator.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_VALIDATION",
				"message": "Validation failed",
				"details": err.Error(),
			},
		})
		return
	}

	newKey, err := c.userService.RotateAPIKey(ctx.Request.Context(), userID.(string), keyID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		errorCode := "ERR_INTERNAL"

		switch err.Error() {
		case "API key not found":
			status = http.StatusNotFound
			errorCode = "ERR_KEY_NOT_FOUND"
		case "unauthorized to rotate this API key":
			status = http.StatusForbidden
			errorCode = "ERR_UNAUTHORIZED"
		}

		ctx.JSON(status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    errorCode,
				"message": err.Error(),
			},
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"api_key":  newKey.APIKey, // The actual API key string
			"key_info": newKey,        // The full key information
			"message":  "API key rotated successfully. Save this key now as it won't be shown again.",
			"warning":  "For security reasons, the full API key will only be shown once.",
		},
	})
}

// GetBalance godoc
// @Summary Get user balance
// @Description Get current user's balance information
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Balance retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/balance [get]
func (c *Controller) GetBalance(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	balance, err := c.userService.GetUserBalance(ctx.Request.Context(), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get balance",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    balance,
	})
}

// GetUsageStatistics godoc
// @Summary Get usage statistics
// @Description Get current user's usage statistics for a date range
// @Tags user
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Usage statistics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/v1/user/usage [get]
func (c *Controller) GetUsageStatistics(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_UNAUTHORIZED",
				"message": "Authentication required",
			},
		})
		return
	}

	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	var startDate, endDate *time.Time
	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			endDate = &parsed
		}
	}

	stats, err := c.userService.GetUsageStatistics(ctx.Request.Context(), userID.(string), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_INTERNAL",
				"message": "Failed to get usage statistics",
			},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
