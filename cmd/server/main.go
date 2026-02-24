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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger, err := initLogger(cfg.Server.Mode)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db)

	logger.Info("Database connection established")

	deviceRepo    := repository.NewDeviceRepository(db)
	receiptRepo   := repository.NewReceiptRepository(db)
	fiscalDayRepo := repository.NewFiscalDayRepository(db)
	userRepo      := repository.NewUserRepository(db)
	adminRepo     := repository.NewAdminRepository(db)

	cryptoSvc, err := service.NewCryptoService(cfg.Crypto)
	if err != nil {
		logger.Fatal("Failed to initialize crypto service", zap.Error(err))
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-jwt-secret-key-change-me"
	}

	validationSvc := service.NewValidationService()
	deviceSvc     := service.NewDeviceService(deviceRepo, cryptoSvc, logger)
	receiptSvc    := service.NewReceiptService(receiptRepo, fiscalDayRepo, deviceRepo, validationSvc, cryptoSvc, logger)
	fiscalDaySvc  := service.NewFiscalDayService(fiscalDayRepo, receiptRepo, deviceRepo, cryptoSvc, logger)
	userSvc       := service.NewUserService(userRepo, deviceRepo, jwtSecret, logger)
	adminSvc      := service.NewAdminService(adminRepo, jwtSecret, logger)

	healthHandler    := handlers.NewHealthHandler()
	deviceHandler    := handlers.NewDeviceHandler(deviceSvc)
	receiptHandler   := handlers.NewReceiptHandler(receiptSvc)
	fiscalDayHandler := handlers.NewFiscalDayHandler(fiscalDaySvc)
	userHandler      := handlers.NewUserHandler(userSvc)
	adminHandler     := handlers.NewAdminHandler(adminSvc)

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.LoggerMiddleware(logger))
	router.Use(middleware.CORSMiddleware())

	setupRoutes(router, healthHandler, deviceHandler, receiptHandler, fiscalDayHandler, userHandler, adminHandler, jwtSecret, logger)

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
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
	adminHandler *handlers.AdminHandler,
	jwtSecret string,
	logger *zap.Logger,
) {
	router.GET("/health", healthHandler.Health)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/device/verify-taxpayer", deviceHandler.VerifyTaxpayer)
		v1.POST("/device/register", deviceHandler.RegisterDevice)
		v1.GET("/server/certificate", deviceHandler.GetServerCertificate)
		v1.POST("/users/login", userHandler.Login)

		protected := v1.Group("")
		protected.Use(middleware.CertificateAuthMiddleware(logger))
		{
			device := protected.Group("/device")
			device.POST("/issue-certificate", deviceHandler.IssueCertificate)
			device.GET("/config", deviceHandler.GetConfig)
			device.GET("/status", deviceHandler.GetStatus)
			device.POST("/ping", deviceHandler.Ping)

			fd := protected.Group("/fiscal-day")
			fd.POST("/open", fiscalDayHandler.OpenFiscalDay)
			fd.POST("/close", fiscalDayHandler.CloseFiscalDay)
			fd.GET("/status", fiscalDayHandler.GetStatus)

			protected.Group("/receipt").POST("/submit", receiptHandler.SubmitReceipt)
			protected.Group("/stock").GET("/list", deviceHandler.GetStockList)

			users := protected.Group("/users")
			users.GET("/list", userHandler.ListUsers)
			users.POST("/create-begin", userHandler.CreateUserBegin)
			users.POST("/create-confirm", userHandler.CreateUserConfirm)
			users.PUT("/update", userHandler.UpdateUser)
			users.PUT("/change-password", userHandler.ChangePassword)
		}
	}

	// Admin API - separate prefix, separate auth
	admin := router.Group("/api/admin")
	admin.POST("/login", adminHandler.Login)

	ap := admin.Group("")
	ap.Use(middleware.AdminAuthMiddleware(jwtSecret))
	{
		ap.GET("/stats", adminHandler.GetStats)

		co := ap.Group("/companies")
		co.GET("", adminHandler.ListCompanies)
		co.POST("", adminHandler.CreateCompany)
		co.GET("/:id", adminHandler.GetCompany)
		co.PUT("/:id", adminHandler.UpdateCompany)
		co.PATCH("/:id/status", adminHandler.SetCompanyStatus)
		co.GET("/:id/devices", adminHandler.ListCompanyDevices)

		dv := ap.Group("/devices")
		dv.GET("", adminHandler.ListAllDevices)
		dv.POST("", adminHandler.ProvisionDevice)
		dv.PATCH("/:deviceID/status", adminHandler.SetDeviceStatus)
		dv.PATCH("/:deviceID/mode", adminHandler.SetDeviceMode)

		ap.GET("/fiscal-days", adminHandler.ListFiscalDays)
		ap.GET("/receipts", adminHandler.ListReceipts)
		ap.GET("/audit", adminHandler.ListAuditLogs)
	}
}
