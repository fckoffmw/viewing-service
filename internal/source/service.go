package source

type repository interface {
	GetAllSources() ([]Source, error)
}

type service struct {
	repo repository
}

func NewService(r repository) *service {
	return &service{
		repo: r,
	}
}

func (s service) GetAllSources() ([]Source, error) {
	return s.repo.GetAllSources()
}
