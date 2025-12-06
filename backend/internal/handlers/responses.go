package handlers

// ErrorResponse стандартный формат ошибки
type ErrorResponse struct {
    Error   string `json:"error" example:"invalid credentials"`
    Message string `json:"message,omitempty" example:"Email or password is incorrect"`
}

// LoginRequest запрос на вход
type LoginRequest struct {
    Email    string `json:"email" example:"user@example.com"`
    Password string `json:"password" example:"password123"`
}

// LoginResponse ответ при входе
type LoginResponse struct {
    UserID string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
    Email  string `json:"email" example:"user@example.com"`
}