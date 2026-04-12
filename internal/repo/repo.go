package repo

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"w2g/internal/auth"
	"w2g/internal/room"
	"w2g/internal/source"
)

var (
	usersUnit = dataUnit{
		filename: "users.csv",
		fields:   []string{"id", "username", "pass_hash"},
	}

	sourcesUnit = dataUnit{
		filename: "sources.csv",
		fields:   []string{"id", "name", "url"},
	}

	roomsUnit = dataUnit{
		filename: "rooms.csv",
		fields:   []string{"id", "source_id"},
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
	dataStruct["sources"] = sourcesUnit
	dataStruct["rooms"] = roomsUnit

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

func (s csvStorage) GetAllSources() ([]source.Source, error) {
	filename := s.dataStruct["sources"].filename

	file, err := os.Open(s.basePath + filename)
	if err != nil {
		return nil, fmt.Errorf("error when open file %s: %w", file, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var frames []source.Source
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error when read file %s: %w", file, err)
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}

		frame := source.Source{
			ID:   row[0],
			Name: row[1],
			Url:  row[2],
		}
		frames = append(frames, frame)
	}

	return frames, nil
}

func (s csvStorage) GetSourceById(id string) (*source.Source, error) {
	filename := s.dataStruct["sources"].filename

	file, err := os.Open(s.basePath + filename)
	if err != nil {
		return nil, fmt.Errorf("error when open file %s: %w", file, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var sources []source.Source
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error when read file %s: %w", file, err)
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}

		source := source.Source{
			ID:   row[0],
			Name: row[1],
			Url:  row[2],
		}
		sources = append(sources, source)
	}

	for _, source := range sources {
		if source.ID == id {
			return &source, nil
		}
	}

	return nil, fmt.Errorf("source with id %s not found", id)
}

func (s csvStorage) GetGlobalRoom() (*room.Room, error) {
	filename := s.dataStruct["rooms"].filename

	file, err := os.Open(s.basePath + filename)
	if err != nil {
		return nil, fmt.Errorf("error when open file %s: %w", file, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var rooms []room.Room
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error when read file %s: %w", file, err)
	}

	for i, row := range rows {
		if i == 0 {
			continue
		}

		room := room.Room{
			ID:       row[0],
			SourceID: row[1],
		}
		rooms = append(rooms, room)
	}

	if len(rooms) == 0 {
		return nil, fmt.Errorf("empty rooms was returned")
	}

	return &rooms[0], nil
}

func (s csvStorage) UpdateGlobalRoomSource(source source.Source) (string, error) {
	globalRoom, err := s.GetGlobalRoom()
	if err != nil {
		return "", fmt.Errorf("error when getting global room: %w", err)
	}

	globalRoom.SourceID = source.ID

	filename := s.dataStruct["rooms"].filename

	file, err := os.Open(s.basePath + filename)
	if err != nil {
		return "", fmt.Errorf("error when open file %s: %w", file, err)
	}
	defer file.Close()

	_, err = file.WriteString(strings.Join(s.dataStruct["rooms"].fields, ",") + "\n" + globalRoom.ID + "," + globalRoom.SourceID + "\n")
	if err != nil {
		return "", fmt.Errorf("when writing to file %s: %w", file, err)
	}

	return source.ID, nil
}
