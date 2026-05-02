package room

type Room struct {
	ID          string
	Name       string
	OwnerID    string
	InviteCode string
	CreatedAt  string
}

func (r Room) GetID() string {
	return r.ID
}