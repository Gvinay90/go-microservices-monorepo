package domain

// User represents a user in the system
// This is a pure domain model with no external dependencies (no GORM, no proto tags)
type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string   // Bcrypt hashed password
	Friends      []string // Friend IDs only
}

// NewUser creates a new User with validation
func NewUser(id, name, email, passwordHash string) (*User, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if email == "" {
		return nil, ErrInvalidEmail
	}

	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Friends:      []string{},
	}, nil
}

// AddFriend adds a friend to the user's friend list
func (u *User) AddFriend(friendID string) error {
	if friendID == u.ID {
		return ErrCannotFriendSelf
	}

	// Check if already friends
	for _, fid := range u.Friends {
		if fid == friendID {
			return ErrAlreadyFriends
		}
	}

	u.Friends = append(u.Friends, friendID)
	return nil
}
