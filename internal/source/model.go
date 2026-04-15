package source

type Source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func (s Source) GetID() string {
	return s.ID
}
