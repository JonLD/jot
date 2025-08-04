package storage

type NoteStore interface {
    Create(note Note) (*Note, error)
    Delete(id string) error
    Edit(id string, opts ...UpdateOption) (*Note, error)
    GetByID(id string) (*Note, error)
    GetAll() ([]*Note, error)
    GetInProject(project string) ([]*Note, error)
}
