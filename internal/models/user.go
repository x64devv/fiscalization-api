package models

import (
	"time"
)

// User represents a system user
type User struct {
	ID            int64      `json:"-" db:"id"`
	TaxpayerID    int64      `json:"-" db:"taxpayer_id"`
	Username      string     `json:"userName" db:"username"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	PersonName    string     `json:"personName" db:"person_name"`
	PersonSurname string     `json:"personSurname" db:"person_surname"`
	UserRole      string     `json:"userRole" db:"user_role"`
	Email         string     `json:"email" db:"email"`
	PhoneNo       string     `json:"phoneNo" db:"phone_no"`
	Status        UserStatus `json:"userStatus" db:"status"`
	SecurityCode  *string    `json:"-" db:"security_code"`
	SecurityCodeExpiry *time.Time `json:"-" db:"security_code_expiry"`
	CreatedAt     time.Time  `json:"-" db:"created_at"`
	UpdatedAt     time.Time  `json:"-" db:"updated_at"`
}

// GetUsersListRequest represents users list request
type GetUsersListRequest struct {
	DeviceID int `json:"deviceID" binding:"required"`
}

// GetUsersListResponse represents users list response
type GetUsersListResponse struct {
	Total       int    `json:"total"`
	OperationID string `json:"operationID"`
	Rows        []User `json:"rows"`
}

// LoginRequest represents login request
type LoginRequest struct {
	DeviceID int    `json:"deviceID" binding:"required"`
	Username string `json:"userName" binding:"required,max=100"`
	Password string `json:"password" binding:"required,max=100"`
}

// LoginResponse represents login response
type LoginResponse struct {
	User        User      `json:"user"`
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expiresAt"`
	OperationID string    `json:"operationID"`
}

// CreateUserBeginRequest represents user creation start request
type CreateUserBeginRequest struct {
	DeviceID      int    `json:"deviceID" binding:"required"`
	Username      string `json:"userName" binding:"required,max=100"`
	PersonName    string `json:"personName" binding:"required,max=100"`
	PersonSurname string `json:"personSurname" binding:"required,max=100"`
	UserRole      string `json:"userRole" binding:"required,max=100"`
}

// CreateUserBeginResponse represents user creation start response
type CreateUserBeginResponse struct {
	OperationID string `json:"operationID"`
}

// CreateUserConfirmRequest represents user creation confirmation request
type CreateUserConfirmRequest struct {
	DeviceID     int    `json:"deviceID" binding:"required"`
	Username     string `json:"userName" binding:"required,max=100"`
	SecurityCode string `json:"securityCode" binding:"required,max=10"`
	Password     string `json:"password" binding:"required,max=100"`
}

// CreateUserConfirmResponse represents user creation confirmation response
type CreateUserConfirmResponse struct {
	User        User   `json:"user"`
	JWTToken    string `json:"jwtToken"`
	OperationID string `json:"operationID"`
}

// SendSecurityCodeRequest represents security code sending request
type SendSecurityCodeRequest struct {
	DeviceID int    `json:"deviceID" binding:"required"`
	Username string `json:"userName" binding:"required,max=100"`
}

// SendSecurityCodeResponse represents security code sending response
type SendSecurityCodeResponse struct {
	OperationID string `json:"operationID"`
}

// SendSecurityCodeContactChangeRequest represents contact change security code request
type SendSecurityCodeContactChangeRequest struct {
	DeviceID  int     `json:"deviceID" binding:"required"`
	PhoneNo   *string `json:"phoneNo,omitempty"`
	UserEmail *string `json:"userEmail,omitempty"`
	Token     string  `json:"token" binding:"required,max=1000"`
}

// SendSecurityCodeContactChangeResponse represents contact change security code response
type SendSecurityCodeContactChangeResponse struct {
	OperationID string `json:"operationID"`
}

// ConfirmUserContactChangeRequest represents contact change confirmation request
type ConfirmUserContactChangeRequest struct {
	DeviceID     int                `json:"deviceID" binding:"required"`
	Channel      SendSecurityCodeTo `json:"channel" binding:"required"`
	SecurityCode string             `json:"securityCode" binding:"required,max=10"`
	Token        string             `json:"token" binding:"required,max=1000"`
}

// ConfirmUserContactChangeResponse represents contact change confirmation response
type ConfirmUserContactChangeResponse struct {
	User        User   `json:"user"`
	OperationID string `json:"operationID"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
	DeviceID      int        `json:"deviceID" binding:"required"`
	Username      string     `json:"userName" binding:"required,max=100"`
	PersonName    string     `json:"personName" binding:"required,max=100"`
	PersonSurname string     `json:"personSurname" binding:"required,max=100"`
	UserRole      string     `json:"userRole" binding:"required,max=100"`
	UserStatus    UserStatus `json:"userStatus" binding:"required"`
	Token         string     `json:"token" binding:"required,max=1000"`
}

// UpdateUserResponse represents user update response
type UpdateUserResponse struct {
	OperationID string `json:"operationID"`
}

// ChangeUserPasswordRequest represents password change request
type ChangeUserPasswordRequest struct {
	DeviceID    int    `json:"deviceID" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,max=100"`
	OldPassword string `json:"oldPassword" binding:"required,max=100"`
	Token       string `json:"token" binding:"required,max=1000"`
}

// ChangeUserPasswordResponse represents password change response
type ChangeUserPasswordResponse struct {
	User        User   `json:"user"`
	Token       string `json:"token"`
	OperationID string `json:"operationID"`
}

// ResetUserPasswordBeginRequest represents password reset start request
type ResetUserPasswordBeginRequest struct {
	DeviceID int                `json:"deviceID" binding:"required"`
	Username string             `json:"userName" binding:"required,max=100"`
	Channel  SendSecurityCodeTo `json:"channel" binding:"required"`
}

// ResetUserPasswordBeginResponse represents password reset start response
type ResetUserPasswordBeginResponse struct {
	OperationID string `json:"operationID"`
}

// ResetUserPasswordConfirmRequest represents password reset confirmation request
type ResetUserPasswordConfirmRequest struct {
	DeviceID     int    `json:"deviceID" binding:"required"`
	Username     string `json:"userName" binding:"required,max=100"`
	NewPassword  string `json:"newPassword" binding:"required,max=100"`
	SecurityCode string `json:"securityCode" binding:"required,max=10"`
}

// ResetUserPasswordConfirmResponse represents password reset confirmation response
type ResetUserPasswordConfirmResponse struct {
	OperationID string `json:"operationID"`
	User        User   `json:"user"`
	Token       string `json:"token"`
}

// ChangePasswordRequest represents password change request (simplified)
type ChangePasswordRequest struct {
	UserID      int64  `json:"userID" binding:"required"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// ChangePasswordResponse represents password change response (simplified)
type ChangePasswordResponse struct {
	OperationID string `json:"operationID"`
	Message     string `json:"message"`
}

// ListUsersResponse represents list users response (simplified)
type ListUsersResponse struct {
	Total int        `json:"total"`
	Users []UserInfo `json:"users"`
}

// UserInfo represents user information (simplified for responses)
type UserInfo struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Name     string  `json:"name"`
	Surname  string  `json:"surname"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Role     string  `json:"role"`
	Status   string  `json:"status"`
}
