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

func (s service) GetAllSources() ([]Source, error) {
	return s.repo.GetAllSources()
}

func (s service) AddSource(name, url string) (string, error) {
	return s.repo.AddSource(&Source{
		Name: name,
		Url:  url,
	})
}
