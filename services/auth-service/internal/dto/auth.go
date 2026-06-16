package dto

// RegisterRequest is what the API expects for POST /register
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=100"`
	Password string `json:"password" binding:"required,min=8"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type GoogleLoginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type WalletLoginRequest struct {
	WalletAddress string `json:"wallet_address"`
	Signature     string `json:"signature"`
	Message       string `json:"message"`
}

type UserPublic struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`

	Role string `json:"role"`

	Status string `json:"status"`

	AvatarURL *string `json:"avatar_url,omitempty"`

	IsVerified bool `json:"is_verified"`
}

// RegisterResponse is what we send back on success
type RegisterResponse struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	IsVerified bool   `json:"is_verified"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`

	ExpiresAt int64 `json:"expires_at"`

	User UserPublic `json:"user"`
}
