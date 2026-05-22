package room

import "time"

type Room struct {
	ID         string
	Name       string
	OwnerID    string
	InviteCode string
	CreatedAt  time.Time
}