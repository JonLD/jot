package storage

import "time"

type Note struct {
    ID        string
    Title     string
    Path      string
    Project   string
    Branch    string
    Ticket    string
    Tags      []string
    CreatedAt time.Time
    ModifiedAt time.Time
}
