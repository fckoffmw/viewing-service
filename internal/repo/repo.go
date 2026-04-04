package repo

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"w2g/internal/auth"
)

var (
	usersUnit = dataUnit{
		filename: "users.csv",
		fields:   []string{"id", "username", "pass_hash"},
	}

	dataStruct = []dataUnit{usersUnit}
)

type dataUnit struct {
	filename string
	fields   []string
}

type csvStorage struct {
}

func NewCSVStorage(path string) (*csvStorage, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("path %s not exists", path)
		}
		return nil, fmt.Errorf("error when init csv storate: %w", err)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error when read dir %s: %w", path, err)
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}

	for _, unit := range dataStruct {
		if slices.Contains(filenames, unit.filename) {
			continue
		}

		file, err := os.Create(path + unit.filename)
		if err != nil {
			return nil, fmt.Errorf("error when creating %s: %w", file, err)
		}

		file.WriteString(strings.Join(unit.fields, ","))

		file.Close()
	}

	return &csvStorage{}, nil
}

func (s csvStorage) GetUserByUsername(username string) (auth.User, error) {
	return auth.User{}, nil
}
