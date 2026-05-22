package source

type repository interface {
	GetAllSources() ([]Source, error)
	AddSource(*Source) (string, error)
}

type service struct {
	repo repository
}

func NewService(r repository) *service {
	return &service{
		repo: r,
	}
}

func (s *service) GetAll() ([]Source, error) {
	return s.repo.GetAllSources()
}

func (s *service) Add(name, url string) (string, error) {
	return s.repo.AddSource(&Source{
		Name: name,
		URL:  url,
	})
}
