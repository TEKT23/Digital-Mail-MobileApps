package dto

import (
	"time"

	"TugasAkhir/models"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	TokenType    string      `json:"token_type,omitempty"`
	ExpiresAt    time.Time   `json:"expires_at"`
	User         UserSummary `json:"user"`
}

type UserSummary struct {
	ID        uint        `json:"id"`
	Username  string      `json:"username"`
	FirstName string      `json:"first_name,omitempty"`
	LastName  string      `json:"last_name,omitempty"`
	Email     string      `json:"email"`
	Role      models.Role `json:"role"`
	Jabatan   string      `json:"jabatan,omitempty"`
	Atribut   string      `json:"atribut,omitempty"`
}

type RegisterRequest struct {
	Username  string      `json:"username" binding:"required,min=3"`
	FirstName string      `json:"first_name" binding:"omitempty,max=100"`
	LastName  string      `json:"last_name" binding:"omitempty,max=100"`
	Email     string      `json:"email" binding:"required,email"`
	Password  string      `json:"password" binding:"required,min=8"`
	Role      models.Role `json:"role" binding:"required"`
	Jabatan   string      `json:"jabatan" binding:"omitempty,max=150"`
	Atribut   string      `json:"atribut" binding:"omitempty"`
}

type RegisterResponse struct {
	User    UserSummary `json:"user"`
	Message string      `json:"message,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type PasswordResetRequest struct {
	Email string `json:"email" form:"email" binding:"required,email"`
}

type PasswordResetSubmission struct {
	Token           string `json:"token" form:"token" binding:"required"`
	Password        string `json:"password" form:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" binding:"required,eqfield=Password"`
}
