package domain

import "errors"

var (
	ErrInvalidName       = errors.New("name is required")
	ErrInvalidEmail      = errors.New("email is required")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrCannotFriendSelf  = errors.New("cannot add yourself as a friend")
	ErrAlreadyFriends    = errors.New("already friends")
)
