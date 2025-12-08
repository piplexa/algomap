package domain

import "time"

// User представляет пользователя системы
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserRequest - запрос на создание пользователя (регистрация)
type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Password string `json:"password"`
}

// UpdateUserRequest - запрос на обновление пользователя
type UpdateUserRequest struct {
	Name *string `json:"name,omitempty"`
}