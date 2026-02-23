package service

import (
	"fmt"
	"time"

	"fiscalization-api/internal/models"
	"fiscalization-api/internal/repository"
	"fiscalization-api/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo   repository.UserRepository
	deviceRepo repository.DeviceRepository
	jwtSecret  string
	logger     *zap.Logger
}

func NewUserService(
	userRepo repository.UserRepository,
	deviceRepo repository.DeviceRepository,
	jwtSecret string,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		userRepo:   userRepo,
		deviceRepo: deviceRepo,
		jwtSecret:  jwtSecret,
		logger:     logger,
	}
}

// Login authenticates a user
func (s *UserService) Login(req models.LoginRequest) (*models.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, models.NewAPIError(401, "Invalid credentials", models.ErrCodeUSER02)
	}

	// Check if user is active
	if user.Status != models.UserStatusActive {
		return nil, models.NewAPIError(401, "User account is not active", models.ErrCodeUSER03)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, models.NewAPIError(401, "Invalid credentials", models.ErrCodeUSER02)
	}

	// Generate JWT token
	token, expiresAt, err := s.generateJWT(user)
	if err != nil {
		s.logger.Error("Failed to generate JWT", zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	s.logger.Info("User logged in", zap.String("username", user.Username))

	return &models.LoginResponse{
		OperationID: utils.GenerateOperationID(),
		Token:       token,
		ExpiresAt:   expiresAt,
		User:        *user,
	}, nil
}

// CreateUserBegin initiates user creation by sending security code
func (s *UserService) CreateUserBegin(req models.CreateUserBeginRequest) (*models.CreateUserBeginResponse, error) {
	// Validate device exists
	device, err := s.deviceRepo.GetByDeviceID(req.DeviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	// Check if username already exists
	existing, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, models.NewAPIError(422, "Username already exists", models.ErrCodeUSER06)
	}

	// Generate security code
	securityCode, err := utils.GenerateSecurityCode(6)
	if err != nil {
		s.logger.Error("Failed to generate security code", zap.Error(err))
		return nil, fmt.Errorf("failed to generate security code")
	}

	// Create temporary user with auto-generated password (will be set during confirmation)
	tempPassword, _ := utils.GenerateActivationKey()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password")
	}

	user := &models.User{
		TaxpayerID:    device.TaxpayerID,
		Username:      req.Username,
		PasswordHash:  string(passwordHash),
		PersonName:    req.PersonName,
		PersonSurname: req.PersonSurname,
		UserRole:      req.UserRole,
		Status:        models.UserStatusNotConfirmed,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, fmt.Errorf("failed to create user")
	}

	// Save security code (expires in 15 minutes)
	expiresAt := time.Now().Add(15 * time.Minute)
	if err := s.userRepo.SaveSecurityCode(user.ID, securityCode, expiresAt); err != nil {
		s.logger.Error("Failed to save security code", zap.Error(err))
		return nil, fmt.Errorf("failed to save security code")
	}

	// In production, send security code via email/SMS
	s.logger.Info("Security code generated for user creation",
		zap.String("username", user.Username),
		zap.String("code", securityCode), // Don't log in production!
	)

	return &models.CreateUserBeginResponse{
		OperationID: utils.GenerateOperationID(),
	}, nil
}

// CreateUserConfirm confirms user creation with security code
func (s *UserService) CreateUserConfirm(req models.CreateUserConfirmRequest) (*models.CreateUserConfirmResponse, error) {
	// Get user by username (not by ID, as CreateUserConfirmRequest uses username)
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, models.NewAPIError(422, "User not found", models.ErrCodeUSER01)
	}

	// Check if user is not confirmed
	if user.Status != models.UserStatusNotConfirmed {
		return nil, models.NewAPIError(422, "User is not in pending state", models.ErrCodeUSER03)
	}

	// Get security code
	savedCode, expiresAt, err := s.userRepo.GetSecurityCode(user.ID)
	if err != nil {
		return nil, err
	}
	if savedCode == "" {
		return nil, models.NewAPIError(422, "Security code not found or expired", models.ErrCodeUSER04)
	}

	// Verify security code
	if savedCode != req.SecurityCode {
		return nil, models.NewAPIError(422, "Invalid security code", models.ErrCodeUSER04)
	}

	// Check expiration
	if time.Now().After(expiresAt) {
		return nil, models.NewAPIError(422, "Security code expired", models.ErrCodeUSER05)
	}

	// Hash the new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password")
	}

	// Update user password and activate
	if err := s.userRepo.UpdatePassword(user.ID, string(passwordHash)); err != nil {
		s.logger.Error("Failed to update password", zap.Error(err))
		return nil, fmt.Errorf("failed to update password")
	}

	user.Status = models.UserStatusActive
	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to activate user", zap.Error(err))
		return nil, fmt.Errorf("failed to activate user")
	}

	// Delete security code
	s.userRepo.DeleteSecurityCode(user.ID)

	// Generate JWT token
	token, _, err := s.generateJWT(user)
	if err != nil {
		s.logger.Error("Failed to generate JWT", zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	s.logger.Info("User created successfully", zap.String("username", user.Username))

	return &models.CreateUserConfirmResponse{
		OperationID: utils.GenerateOperationID(),
		User:        *user,
		JWTToken:    token,
	}, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(req models.UpdateUserRequest) (*models.UpdateUserResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, models.NewAPIError(422, "User not found", models.ErrCodeUSER01)
	}

	// Update fields
	user.PersonName = req.PersonName
	user.PersonSurname = req.PersonSurname
	user.UserRole = req.UserRole
	user.Status = req.UserStatus

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return nil, fmt.Errorf("failed to update user")
	}

	return &models.UpdateUserResponse{
		OperationID: utils.GenerateOperationID(),
	}, nil
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(req models.ChangePasswordRequest) (*models.ChangePasswordResponse, error) {
	// Get user
	user, err := s.userRepo.GetByID(req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, models.NewAPIError(422, "User not found", models.ErrCodeUSER01)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return nil, models.NewAPIError(422, "Invalid old password", models.ErrCodeUSER02)
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to hash password")
	}

	// Update password
	if err := s.userRepo.UpdatePassword(user.ID, string(passwordHash)); err != nil {
		s.logger.Error("Failed to update password", zap.Error(err))
		return nil, fmt.Errorf("failed to update password")
	}

	s.logger.Info("Password changed", zap.String("username", user.Username))

	return &models.ChangePasswordResponse{
		OperationID: utils.GenerateOperationID(),
		Message:     "Password changed successfully",
	}, nil
}

// ListUsers lists all users for a taxpayer
func (s *UserService) ListUsers(deviceID int, offset, limit int) (*models.ListUsersResponse, error) {
	// Get device to get taxpayer ID
	device, err := s.deviceRepo.GetByDeviceID(deviceID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, models.NewAPIError(422, "Device not found", models.ErrCodeDEV01)
	}

	users, total, err := s.userRepo.List(device.TaxpayerID, offset, limit)
	if err != nil {
		return nil, err
	}

	// Convert to response format (exclude password hash)
	userInfos := make([]models.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = models.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.PersonName,
			Surname:  user.PersonSurname,
			Email:    &user.Email,
			Phone:    &user.PhoneNo,
			Role:     user.UserRole,
			Status:   user.Status.String(),
		}
	}

	return &models.ListUsersResponse{
		Total: total,
		Users: userInfos,
	}, nil
}

// Helper methods

func (s *UserService) generateJWT(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.UserRole,
		"exp":      expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *UserService) ValidateJWT(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	userID := int64(claims["user_id"].(float64))
	return s.userRepo.GetByID(userID)
}
