package storage

type NoteStore interface {
    Create(note Note) (*Note, error)
    Open(id string) error
    Delete(id string) error
    Update(id string, opts ...UpdateOption) (*Note, error)
    GetByID(id string) (*Note, error)
    GetAll() ([]*Note, error)
    GetInProject(project string) ([]*Note, error)
    GetProjectMisc(project string) ([]*Note, error)
    GetByTicket(ticket string) ([]*Note, error)
    GetByBranch(branch string) ([]*Note, error)
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
