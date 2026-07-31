package source

import (
	"fmt"
	"w2g/internal/utils/str"
)

type repository interface {
	GetAllSources() ([]Source, error)
	GetSourceByID(id string) (*Source, error)
	AddSource(*Source) (string, error)
	UpdateSource(*Source) error
	DeleteSource(id string) error
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
	sources, err := s.repo.GetAllSources()
	if err != nil {
		return nil, fmt.Errorf("get all sources: %w", err)
	}

	return sources, nil
}

func (s *service) Add(name, url string) (string, error) {
	url, err := str.ExtractURLFromIframeTag(url)
	if err != nil {
		return "", fmt.Errorf("validating url: %w", err)
	}

	id, err := s.repo.AddSource(&Source{
		Name: name,
		URL:  url,
	})

	if err != nil {
		return "", fmt.Errorf("add source: %w", err)
	}

	return id, nil
}

func (s *service) Update(id, name, url string) error {
	src, err := s.repo.GetSourceByID(id)
	if err != nil {
		return fmt.Errorf("update source: %w", err)
	}
	if src == nil {
		return fmt.Errorf("update source: source with id %s not found", id)
	}

	src.Name = name
	src.URL = url

	if err := s.repo.UpdateSource(src); err != nil {
		return fmt.Errorf("update source: %w", err)
	}

	return nil
}

func (s *service) Delete(id string) error {
	src, err := s.repo.GetSourceByID(id)
	if err != nil {
		return fmt.Errorf("delete source: %w", err)
	}
	if src == nil {
		return fmt.Errorf("delete source: source with id %s not found", id)
	}

	if err := s.repo.DeleteSource(id); err != nil {
		return fmt.Errorf("delete source: %w", err)
	}

	return nil
}
