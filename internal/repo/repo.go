package repo

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"w2g/internal/auth"
	"w2g/internal/frame"
)

var (
	usersUnit = dataUnit{
		filename: "users.csv",
		fields:   []string{"id", "username", "pass_hash"},
	}

	framesUnit = dataUnit{
		filename: "frames.csv",
		fields:   []string{"id", "name", "iframe"},
	}
)

type dataUnit struct {
	filename string
	fields   []string
}

type csvStorage struct {
	dataStruct map[string]dataUnit
	basePath   string
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

	dataStruct := make(map[string]dataUnit)
	dataStruct["users"] = usersUnit
	dataStruct["frames"] = framesUnit

	for _, unit := range dataStruct {
		if slices.Contains(filenames, unit.filename) {
			continue
		}

		file, err := os.Create(path + unit.filename)
		if err != nil {
			return nil, fmt.Errorf("error when creating %s: %w", file, err)
		}

		file.WriteString(strings.Join(unit.fields, ",") + "\n")

		file.Close()
	}

	return &csvStorage{
		dataStruct: dataStruct,
		basePath:   path,
	}, nil
}

func (s csvStorage) GetUserByUsername(username string) (auth.User, error) {
	return auth.User{}, nil
}

func (s csvStorage) GetAllFrames() ([]frame.Frame, error) {
	filename := s.dataStruct["frames"].filename

	file, err := os.Open(s.basePath + filename)
	if err != nil {
		return nil, fmt.Errorf("error when open file %s: %w", file, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var frames []frame.Frame
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error when read file %s: %w", file, err)
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}

		frame := frame.Frame{
			ID:     row[0],
			Name:   row[1],
			Iframe: row[2],
		}
		frames = append(frames, frame)
	}

	return frames, nil
}
