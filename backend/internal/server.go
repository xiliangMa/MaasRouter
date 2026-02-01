package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"massrouter.ai/backend/api/swagger"
	"massrouter.ai/backend/internal/config"
	"massrouter.ai/backend/internal/controller/admin"
	"massrouter.ai/backend/internal/controller/auth"
	"massrouter.ai/backend/internal/controller/billing"
	"massrouter.ai/backend/internal/controller/health"
	"massrouter.ai/backend/internal/controller/model"
	"massrouter.ai/backend/internal/controller/oauth"
	proxyController "massrouter.ai/backend/internal/controller/proxy"
	"massrouter.ai/backend/internal/controller/user"
	"massrouter.ai/backend/internal/middleware"
	"massrouter.ai/backend/internal/service"
	pkgAuth "massrouter.ai/backend/pkg/auth"
	"massrouter.ai/backend/pkg/cache"
	"massrouter.ai/backend/pkg/database"
)

type Server struct {
	cfg         *config.Config
	logger      zerolog.Logger
	db          *database.Database
	redisClient *cache.RedisClient
	jwtManager  *pkgAuth.JWTManager
	router      *gin.Engine
	httpServer  *http.Server

	// Controllers
	healthController  *health.Controller
	authController    *auth.Controller
	oauthController   *oauth.Controller
	userController    *user.Controller
	modelController   *model.Controller
	billingController *billing.Controller
	adminController   *admin.Controller
	proxyController   *proxyController.Controller

	// Services
	billingService service.BillingService
}

func NewServer(
	cfg *config.Config,
	logger zerolog.Logger,
	db *database.Database,
	redisClient *cache.RedisClient,
	jwtManager *pkgAuth.JWTManager,
	healthController *health.Controller,
	authController *auth.Controller,
	oauthController *oauth.Controller,
	userController *user.Controller,
	modelController *model.Controller,
	billingController *billing.Controller,
	adminController *admin.Controller,
	proxyController *proxyController.Controller,
	billingService service.BillingService,
) *Server {
	server := &Server{
		cfg:               cfg,
		logger:            logger,
		db:                db,
		redisClient:       redisClient,
		jwtManager:        jwtManager,
		healthController:  healthController,
		authController:    authController,
		oauthController:   oauthController,
		userController:    userController,
		modelController:   modelController,
		billingController: billingController,
		adminController:   adminController,
		proxyController:   proxyController,
		billingService:    billingService,
	}

	server.setupRouter()
	return server
}

func (s *Server) setupRouter() {
	if s.cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	s.router = gin.New()

	// Middleware
	s.router.Use(middleware.Recovery(s.logger))
	s.router.Use(middleware.Logger(s.logger))
	s.router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowedOrigins:   s.cfg.CORS.AllowedOrigins,
		AllowedMethods:   s.cfg.CORS.AllowedMethods,
		AllowedHeaders:   s.cfg.CORS.AllowedHeaders,
		AllowCredentials: s.cfg.CORS.AllowCredentials,
	}))
	s.router.Use(middleware.RequestID())

	if s.redisClient != nil {
		s.router.Use(middleware.DefaultRateLimit(s.redisClient.Client))
	}

	// Swagger documentation
	swagger.RegisterSwagger(s.router)

	// Public routes
	api := s.router.Group("/api/v1")
	{
		// Health checks
		api.GET("/health", s.healthController.HealthCheck)
		api.GET("/ready", s.healthController.ReadinessCheck)
		api.GET("/live", s.healthController.LivenessCheck)

		// Version
		api.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"version": "1.0.0",
					"name":    "MassRouter API",
					"build":   time.Now().Format("20060102.150405"),
				},
			})
		})

		// Authentication routes (public)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", s.authController.Register)
			authGroup.POST("/login", s.authController.Login)
			authGroup.POST("/refresh", s.authController.RefreshToken)
			authGroup.POST("/logout", middleware.JWTAuth(s.jwtManager), s.authController.Logout)
			authGroup.POST("/password/reset/request", s.authController.RequestPasswordReset)
			authGroup.POST("/password/reset", s.authController.ResetPassword)

			// OAuth routes
			authGroup.GET("/oauth/providers", s.oauthController.GetOAuthProviders)
			authGroup.POST("/oauth/start", s.oauthController.StartOAuthFlow)
			authGroup.POST("/oauth/callback", s.oauthController.HandleOAuthCallback)
		}

		// Public model routes
		publicModelGroup := api.Group("/models")
		{
			publicModelGroup.GET("", s.modelController.ListModels)
			publicModelGroup.GET("/search", s.modelController.SearchModels)
			publicModelGroup.GET("/:id", s.modelController.GetModelDetails)
			publicModelGroup.GET("/providers", s.modelController.GetModelProviders)
			publicModelGroup.GET("/categories", s.modelController.GetModelCategories)
		}

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(s.jwtManager))
		{
			// User routes
			userGroup := protected.Group("/user")
			{
				userGroup.GET("/profile", s.userController.GetProfile)
				userGroup.PUT("/profile", s.userController.UpdateProfile)
				userGroup.POST("/password/change", s.userController.ChangePassword)
				userGroup.GET("/api-keys", s.userController.ListAPIKeys)
				userGroup.POST("/api-keys", s.userController.CreateAPIKey)
				userGroup.DELETE("/api-keys/:id", s.userController.DeleteAPIKey)
				userGroup.POST("/api-keys/:id/rotate", s.userController.RotateAPIKey)
				userGroup.GET("/balance", s.userController.GetBalance)
				userGroup.GET("/usage", s.userController.GetUsageStatistics)
			}

			// Auth routes (protected)
			authGroup := protected.Group("/auth")
			{
				// OAuth disconnect route
				authGroup.POST("/oauth/disconnect", s.oauthController.DisconnectOAuthAccount)
			}

			// Billing routes
			billingGroup := protected.Group("/billing")
			{
				billingGroup.GET("/balance", s.billingController.GetBalance)
				billingGroup.GET("/payments", s.billingController.GetPaymentHistory)
				billingGroup.POST("/payments", s.billingController.CreatePayment)
				billingGroup.GET("/records", s.billingController.GetBillingRecords)
				billingGroup.POST("/calculate-cost", s.billingController.CalculateCost)
				billingGroup.POST("/webhook", s.billingController.ProcessPaymentWebhook)
			}

			// Proxy routes for AI model access
			proxyGroup := protected.Group("/chat")
			// Apply user-based rate limiting if Redis is available
			if s.redisClient != nil {
				proxyGroup.Use(middleware.AuthRateLimit(s.redisClient.Client))
			}
			{
				proxyGroup.POST("/completions", s.proxyController.ChatCompletion)
			}
		}
	}

	// Admin routes
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.JWTAuth(s.jwtManager), middleware.RequireAdmin())
	{
		// User management
		adminGroup.GET("/users", s.adminController.ListUsers)
		adminGroup.GET("/users/:id", s.adminController.GetUserDetails)
		adminGroup.PUT("/users/:id", s.adminController.UpdateUser)

		// Model provider management
		adminGroup.POST("/providers", s.adminController.CreateModelProvider)
		adminGroup.PUT("/providers/:id", s.adminController.UpdateModelProvider)

		// Model management
		adminGroup.POST("/models", s.adminController.CreateModel)
		adminGroup.PUT("/models/:id", s.adminController.UpdateModel)

		// System management
		adminGroup.GET("/stats", s.adminController.GetSystemStats)
		adminGroup.PUT("/config/:key", s.adminController.UpdateSystemConfig)
	}

	// Debug route to test routing
	s.router.GET("/debug/routes", func(c *gin.Context) {
		routes := s.router.Routes()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    routes,
		})
	})

	s.router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ERR_404",
				"message": "Endpoint not found",
				"details": "The requested resource does not exist",
			},
		})
	})
}

func (s *Server) Start() error {
	s.logger.Info().Str("port", s.cfg.Server.Port).Str("mode", s.cfg.Server.Mode).Msg("Starting server")

	// Start billing worker for async processing
	if s.billingService != nil {
		s.billingService.StartBillingWorker()
		s.logger.Info().Msg("Billing worker started")
	}

	s.httpServer = &http.Server{
		Addr:         ":" + s.cfg.Server.Port,
		Handler:      s.router,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Stop billing worker
	if s.billingService != nil {
		s.billingService.StopBillingWorker()
	}

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) Router() *gin.Engine {
	return s.router
}
