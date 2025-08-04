package storage

import (
    "fmt"
    "time"

    "github.com/google/uuid"
)

type InMemoryStore struct {
    notes map[string]*Note
}

func NewInMemoryStore() *InMemoryStore {
    return &InMemoryStore{
        notes: make(map[string]*Note),
    }
}

func (s *InMemoryStore) Create(note Note) (*Note, error) {
    note.ID = uuid.NewString()
    note.CreatedAt = time.Now()
    note.ModifiedAt = time.Now()
    s.notes[note.ID] = &note
    return &note, nil
}

func (s *InMemoryStore) Delete(id string) error {
    delete(s.notes, id)
    return nil
}

type UpdateOption func(*Note)

func WithTitle(title string) UpdateOption {
    return func(n *Note) { n.Title = title }
}

func WithPath(path string) UpdateOption {
    return func(n *Note) { n.Path = path }
}

func WithProject(project string) UpdateOption {
    return func(n *Note) { n.Project = project }
}

func WithBranch(branch string) UpdateOption {
    return func(n *Note) { n.Branch = branch }
}

func WithTicket(ticket string) UpdateOption {
    return func(n *Note) { n.Ticket = ticket }
}

func WithTags(tags []string) UpdateOption {
    return func(n *Note) { n.Tags = tags }
}

func (s *InMemoryStore) Edit(id string, opts ...UpdateOption) (*Note, error) {
    if note, exists := s.notes[id]; exists {
        for _, opt := range opts {
            opt(note)
        }
        note.ModifiedAt = time.Now()
        return note, nil
    } else {
        return nil, fmt.Errorf("note with id %s note found", id)
    }
}

func (s *InMemoryStore) GetByID(id string) (*Note, error) {

    if note, exists := s.notes[id]; exists {
        return note, nil
    } else {
        return nil, fmt.Errorf("note with id %s note found", id)
    }
}

func (s *InMemoryStore) GetAll() ([]*Note, error) {
    var noteList []*Note
    for _, note := range s.notes {
        noteList = append(noteList, note)
    }
    return noteList, nil
}

func (s *InMemoryStore) GetInProject(project string) ([]*Note, error) {
    var noteList []*Note
    for _, note := range s.notes {
        if note.Project == project {
            noteList = append(noteList, note)
        }
    }
    return noteList, nil
}

