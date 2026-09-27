package main

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

type Entity struct {
	ID      string
	Name    string
	Content string
	Version int64
}

var ErrNotFound = errors.New("entity not found")

type Storage struct {
	mu       sync.RWMutex
	filePath string
	data     map[string]*Entity
}

func NewStorage(path string) (*Storage, error) {
	s := &Storage{filePath: path, data: make(map[string]*Entity)}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) load() error {
	f, err := os.Open(s.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return err
	}
	for i, row := range rows {
		if i == 0 { // заголовок
			continue
		}
		if len(row) < 4 {
			continue
		}
		v, _ := strconv.ParseInt(row[3], 10, 64)
		s.data[row[0]] = &Entity{ID: row[0], Name: row[1], Content: row[2], Version: v}
	}
	return nil
}

func (s *Storage) save() error {
	f, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	_ = w.Write([]string{"id", "name", "content", "version"})
	for _, e := range s.data {
		_ = w.Write([]string{e.ID, e.Name, e.Content, strconv.FormatInt(e.Version, 10)})
	}
	return nil
}

func (s *Storage) Create(name, content string) (*Entity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := &Entity{ID: uuid.NewString(), Name: name, Content: content, Version: 1}
	s.data[e.ID] = e
	if err := s.save(); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Storage) Get(id string) (*Entity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (s *Storage) Update(id, name, content string) (*Entity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	if name != "" {
		e.Name = name
	}
	if content != "" {
		e.Content = content
	}
	e.Version++
	if err := s.save(); err != nil {
		return nil, err
	}
	cp := *e
	return &cp, nil
}

func (s *Storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[id]; !ok {
		return ErrNotFound
	}
	delete(s.data, id)
	return s.save()
}

func (s *Storage) List(page, pageSize int32) ([]*Entity, int32) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]*Entity, 0, len(s.data))
	for _, e := range s.data {
		cp := *e
		all = append(all, &cp)
	}
	total := int32(len(all))

	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * pageSize
	if start >= total {
		return []*Entity{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], total
}