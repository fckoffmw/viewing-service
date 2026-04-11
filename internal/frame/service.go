package frame

type repository interface {
	GetAllFrames() ([]Frame, error)
}

type service struct {
	repo repository
}

func NewService(r repository) *service {
	return &service{
		repo: r,
	}
}

func (s service) GetAllFrames() ([]Frame, error) {
	return s.repo.GetAllFrames()
}
