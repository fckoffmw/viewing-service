package auth

import "time"

type User struct {
	ID        string
	Username  string
	Password  string
	CreatedAt time.Time
}

type Session struct {
	SessionID  string
	UserID     string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
}
