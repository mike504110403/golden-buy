package user

import "errors"

var (
	// ErrUserNotFound 用戶未找到
	ErrUserNotFound = errors.New("user not found")

	// ErrInsufficientBalance 餘額不足
	ErrInsufficientBalance = errors.New("insufficient balance")

	// ErrInvalidUserID 無效的用戶 ID
	ErrInvalidUserID = errors.New("invalid user id")
)

