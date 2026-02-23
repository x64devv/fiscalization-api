package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fiscalization-api/internal/config"
	"fiscalization-api/internal/database"
	"fiscalization-api/internal/handlers"
	"fiscalization-api/internal/middleware"
	"fiscalization-api/internal/repository"
	"fiscalization-api/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.Server.Mode)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Connect to database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db)

	logger.Info("Database connection established")

	// Initialize repositories
	deviceRepo := repository.NewDeviceRepository(db)
	receiptRepo := repository.NewReceiptRepository(db)
	fiscalDayRepo := repository.NewFiscalDayRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	cryptoSvc, err := service.NewCryptoService(cfg.Crypto)
	if err != nil {
		logger.Fatal("Failed to initialize crypto service", zap.Error(err))
	}

	validationSvc := service.NewValidationService()
	deviceSvc := service.NewDeviceService(deviceRepo, cryptoSvc, logger)
	receiptSvc := service.NewReceiptService(receiptRepo, fiscalDayRepo, deviceRepo, validationSvc, cryptoSvc, logger)
	fiscalDaySvc := service.NewFiscalDayService(fiscalDayRepo, receiptRepo, deviceRepo, cryptoSvc, logger)
	userSvc := service.NewUserService(userRepo, deviceRepo, "your-jwt-secret-key-change-me", logger)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	deviceHandler := handlers.NewDeviceHandler(deviceSvc)
	receiptHandler := handlers.NewReceiptHandler(receiptSvc)
	fiscalDayHandler := handlers.NewFiscalDayHandler(fiscalDaySvc)
	userHandler := handlers.NewUserHandler(userSvc)

	// Initialize Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CORSMiddleware())

	// Setup routes
	setupRoutes(router, healthHandler, deviceHandler, receiptHandler, fiscalDayHandler, userHandler, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func initLogger(mode string) (*zap.Logger, error) {
	if mode == "release" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func setupRoutes(
	router *gin.Engine,
	healthHandler *handlers.HealthHandler,
	deviceHandler *handlers.DeviceHandler,
	receiptHandler *handlers.ReceiptHandler,
	fiscalDayHandler *handlers.FiscalDayHandler,
	userHandler *handlers.UserHandler,
	logger *zap.Logger,
) {
	// Health check
	router.GET("/health", healthHandler.Health)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public device routes (no authentication)
		v1.POST("/device/verify-taxpayer", deviceHandler.VerifyTaxpayer)
		v1.POST("/device/register", deviceHandler.RegisterDevice)
		v1.GET("/server/certificate", deviceHandler.GetServerCertificate)

		// Protected routes (certificate authentication required)
		protected := v1.Group("")
		protected.Use(middleware.CertificateAuthMiddleware(logger))
		{
			// Device management
			device := protected.Group("/device")
			{
				device.POST("/issue-certificate", deviceHandler.IssueCertificate)
				device.GET("/config", deviceHandler.GetConfig)
				device.GET("/status", deviceHandler.GetStatus)
				device.POST("/ping", deviceHandler.Ping)
			}

			// Fiscal day management
			fiscalDay := protected.Group("/fiscal-day")
			{
				fiscalDay.POST("/open", fiscalDayHandler.OpenFiscalDay)
				fiscalDay.POST("/close", fiscalDayHandler.CloseFiscalDay)
				fiscalDay.GET("/status", fiscalDayHandler.GetStatus)
			}

			// Receipt management
			receipt := protected.Group("/receipt")
			{
				receipt.POST("/submit", receiptHandler.SubmitReceipt)
			}

			// Stock management
			stock := protected.Group("/stock")
			{
				stock.GET("/list", deviceHandler.GetStockList)
			}

			// User management
			users := protected.Group("/users")
			{
				users.GET("/list", userHandler.ListUsers)
				users.POST("/create-begin", userHandler.CreateUserBegin)
				users.POST("/create-confirm", userHandler.CreateUserConfirm)
				users.PUT("/update", userHandler.UpdateUser)
				users.PUT("/change-password", userHandler.ChangePassword)
			}
		}

		// Public user routes
		v1.POST("/users/login", userHandler.Login)
	}
}
