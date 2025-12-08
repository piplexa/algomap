package domain

import "time"

// Session представляет сессию пользователя
type Session struct {
	ID        string    `json:"id"`         // session_key (uuid)
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginRequest - запрос на логин
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse - ответ после успешного логина
type LoginResponse struct {
	SessionKey string `json:"session_key"`
	User       *User  `json:"user"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// TODO: OAuth провайдеры (Google, GitHub)
// type OAuthProvider struct {
//     ID             int64
//     UserID         int64
//     Provider       string  // google, github
//     ProviderUserID string  // ID у провайдера
//     Email          string
//     CreatedAt      time.Time
// }