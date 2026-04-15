package room

type Room struct {
	ID       string
	SourceID string
}

func (r Room) GetID() string {
	return r.ID
}
