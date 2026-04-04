package auth

type repository interface {
	GetUserByUsername(username string) (User, error)
}

type service struct {
	repo repository
}

func NewService(r repository) *service {
	return &service{
		repo: r,
	}
}

func (s service) Login(creds credentials) error {
	// достать пользоваля по имени
	user, err := s.repo.GetUserByUsername(creds.Username)
	if err != nil {

	}
	_ = user
	// сверить хеши
	// сгенерить сессию
	// вернуть сессию
	return nil
}
