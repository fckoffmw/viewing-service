package repo

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"w2g/internal/auth"
	"w2g/internal/room"
	"w2g/internal/source"
)

var (
	globalRoomID = "1"

	usersTable   = "users"
	sourcesTable = "sources"
	roomsTable   = "rooms"

	usersUnit = dataUnit{
		filename: usersTable + ".csv",
		fields:   []string{"id", "username", "pass_hash", "created_at"},
	}

	sourcesUnit = dataUnit{
		filename: sourcesTable + ".csv",
		fields:   []string{"id", "name", "url"},
	}

	roomsUnit = dataUnit{
		filename: roomsTable + ".csv",
		fields:   []string{"id", "source_id"},
	}
)

type dataUnit struct {
	filename string
	fields   []string
	lastID   int
}

type hasID interface {
	GetID() string
}

type csvStorage struct {
	dataStruct map[string]*dataUnit
	basePath   string

	mu sync.RWMutex
}

func NewCSVStorage(path string) (*csvStorage, error) {
	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("path %s not exists", path)
		}
		return nil, fmt.Errorf("when init csv storate: %w", err)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("when read dir %s: %w", path, err)
	}

	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}

	dataStruct := make(map[string]*dataUnit)
	dataStruct[usersTable] = &usersUnit
	dataStruct[sourcesTable] = &sourcesUnit
	dataStruct[roomsTable] = &roomsUnit

	for _, unit := range dataStruct {
		if slices.Contains(filenames, unit.filename) {
			lines, err := countLinesWithoutHeader(path + unit.filename)
			if err != nil {
				return nil, fmt.Errorf("when counting lines in %s: %w", unit.filename, err)
			}
			unit.lastID = lines

			continue
		}

		file, err := os.Create(path + unit.filename)
		if err != nil {
			return nil, fmt.Errorf("when creating %s: %w", unit.filename, err)
		}

		if _, err := file.WriteString(strings.Join(unit.fields, ",") + "\n"); err != nil {
			return nil, fmt.Errorf("when writing to %s: %w", unit.filename, err)
		}

		file.Close()
	}

	return &csvStorage{
		dataStruct: dataStruct,
		basePath:   path,
	}, nil
}

func (s *csvStorage) GetUserByUsername(username string) (*auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := s.dataStruct[usersTable].filename

	rows, err := readAllFromFile(s.basePath + filename)
	if err != nil {
		return nil, fmt.Errorf("when read file %s: %w", filename, err)
	}

	users, err := rowsTo[auth.User](rows)
	if err != nil {
		return nil, fmt.Errorf("when convert csv rows to user struct: %w", err)
	}

	for _, u := range users {
		if u.Username == username {
			return &u, nil
		}
	}

	return nil, nil
}

func (s *csvStorage) GetAllSources() ([]source.Source, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := s.dataStruct[sourcesTable].filename

	rows, err := readAllFromFile(s.basePath + filename)
	if err != nil {
		return nil, fmt.Errorf("when read file %s: %w", filename, err)
	}

	sources, err := rowsTo[source.Source](rows)
	if err != nil {
		return nil, fmt.Errorf("when convert csv rows to source struct: %w", err)
	}

	return sources, nil
}

func (s *csvStorage) GetSourceById(id string) (*source.Source, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := s.dataStruct[sourcesTable].filename

	return getByIDFrom[source.Source](id, s.basePath+filename)
}

func (s *csvStorage) GetRoomByID(id string) (*room.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := s.dataStruct[roomsTable].filename

	return getByIDFrom[room.Room](id, s.basePath+filename)
}

func (s *csvStorage) GetGlobalRoom() (*room.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filename := s.dataStruct[roomsTable].filename

	return getByIDFrom[room.Room](globalRoomID, s.basePath+filename)
}

func (s *csvStorage) UpdateGlobalRoomSource(sourceID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	filename := s.dataStruct[roomsTable].filename

	globalRoom, err := getByIDFrom[room.Room](globalRoomID, s.basePath+filename)
	if err != nil {
		return "", fmt.Errorf("when getting global room: %w", err)
	}

	globalRoom.SourceID = sourceID

	file, err := os.Create(s.basePath + filename)
	if err != nil {
		return "", fmt.Errorf("when open file %s: %w", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(s.dataStruct[roomsTable].fields); err != nil {
		return "", fmt.Errorf("when writing header: %w", err)
	}

	if err := writer.Write([]string{globalRoom.ID, globalRoom.SourceID}); err != nil {
		return "", fmt.Errorf("when writing row: %w", err)
	}

	return sourceID, nil
}

func (s *csvStorage) AddSource(source source.Source) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.basePath+s.dataStruct[sourcesTable].filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	id := s.getNewSourceID()

	record := []string{
		id,
		source.Name,
		source.Url,
	}
	if err := writer.Write(record); err != nil {
		return "", err
	}

	s.dataStruct[sourcesTable].lastID += 1

	return id, nil
}

func (s *csvStorage) AddUser(user auth.User) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(s.basePath+s.dataStruct[usersTable].filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	id := s.getNewSourceID()

	record := []string{
		id,
		user.Username,
		user.Password,
		time.Now().Format(),
	}
	if err := writer.Write(record); err != nil {
		return "", err
	}

	s.dataStruct[sourcesTable].lastID += 1

	return id, nil
}

func countLinesWithoutHeader(path string) (int, error) {
	lines, err := readAllFromFile(path)
	if err != nil {
		return 0, err
	}

	return len(lines) - 1, nil
}

func (s *csvStorage) getNewSourceID() string {
	return strconv.Itoa(s.dataStruct[sourcesTable].lastID + 1)
}

func readAllFromFile(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("when open file %s: %w", path, err)
	}
	defer file.Close()

	return csv.NewReader(file).ReadAll()
}

func rowsTo[T any](rows [][]string) ([]T, error) {
	var res []T

	rt := reflect.TypeOf((*T)(nil)).Elem()

	for i, row := range rows {
		if i == 0 {
			continue
		}

		val := reflect.New(rt).Elem().Interface().(T)
		v := reflect.ValueOf(&val).Elem()

		for i := 0; i < rt.NumField() && i < len(row); i++ {
			field := rt.Field(i)

			if field.Anonymous {
				continue
			}

			f := v.Field(i)
			if !f.CanSet() {
				continue
			}

			switch f.Kind() {
			case reflect.String:
				f.SetString(row[i])
			case reflect.Int, reflect.Int64:
				num, err := strconv.ParseInt(row[i], 10, 64)
				if err != nil {
					return nil, err
				}

				f.SetInt(num)
			case reflect.Float64:
				num, err := strconv.ParseFloat(row[i], 64)
				if err != nil {
					return nil, err
				}

				f.SetFloat(num)

			}

		}

		res = append(res, val)
	}

	return res, nil
}

func getByIDFrom[T hasID](id, path string) (*T, error) {
	rows, err := readAllFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("when read file %s: %w", path, err)
	}

	items, err := rowsTo[T](rows)
	if err != nil {
		return nil, fmt.Errorf("when convert csv rows to struct: %w", err)
	}

	for i := range items {
		if items[i].GetID() == id {
			return &items[i], nil
		}
	}

	return nil, fmt.Errorf("item with id %s not found", id)
}
