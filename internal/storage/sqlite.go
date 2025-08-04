package storage

import (
	"fmt"
	"time"
	"database/sql"
	"os"
	"path/filepath"
	"encoding/json"
	"os/exec"
	"runtime"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// In your SQLiteStore initialization
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Default to ~/.jot/notes.db if no path is provided
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = filepath.Join(homeDir, ".jot", "notes.db")

		// Create the directory if it doesn't exist
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create table if it doesn't exist
	schema := `
	CREATE TABLE IF NOT EXISTS notes (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	path TEXT,
	project TEXT,
	branch TEXT,
	ticket TEXT,
	tags TEXT,
	created_at DATETIME,
	modified_at DATETIME
	);`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

type SQLiteStore struct {
	db *sql.DB
}

func (store *SQLiteStore) Create(note Note) (*Note, error) {
	note.ID = uuid.NewString()
	note.CreatedAt = time.Now()
	note.ModifiedAt = time.Now()

	// Build the file path if not provided
	if note.Path == "" {
		// Create logical path: project/ticket/branch/ or project/branch/
		var logicalPath string
		if note.Ticket != "" {
			logicalPath = filepath.Join(note.Project, note.Ticket, note.Branch)
		} else {
			logicalPath = filepath.Join(note.Project, note.Branch)
		}

		// Get notes directory (same as database directory)
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		notesDir := filepath.Join(homeDir, ".jot", "notes", logicalPath)
		note.Path = filepath.Join(notesDir, note.Title+".md")
	}

	// Create the directory structure
	dir := filepath.Dir(note.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create the markdown file with basic content
	content := fmt.Sprintf("# %s\n\nCreated: %s\nProject: %s\nBranch: %s\n",
		note.Title, note.CreatedAt.Format("2006-01-02 15:04:05"), note.Project, note.Branch)
	if note.Ticket != "" {
		content += fmt.Sprintf("Ticket: %s\n", note.Ticket)
	}
	content += "\n---\n\n"

	if err := os.WriteFile(note.Path, []byte(content), 0644); err != nil {
		return nil, err
	}

	// Convert tags slice to JSON string
	tagsJSON, err := json.Marshal(note.Tags)
	if err != nil {
		return nil, err
	}

	_, err = store.db.Exec(`
		INSERT INTO notes (id, title, path, project, branch, ticket, tags, created_at, modified_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		note.ID, note.Title, note.Path, note.Project, note.Branch, note.Ticket,
		string(tagsJSON), note.CreatedAt, note.ModifiedAt)

	if err != nil {
		return nil, err
	}

	return &note, nil
}

func (store *SQLiteStore) Delete(id string) error {
	// Get the note first to find the file path
	note, err := store.GetByID(id)
	if err != nil {
		return err
	}

	// Delete the file if it exists
	if _, err := os.Stat(note.Path); err == nil {
		if err := os.Remove(note.Path); err != nil {
			return fmt.Errorf("failed to delete file: %v", err)
		}
	}

	// Delete from database
	_, err = store.db.Exec("DELETE FROM notes WHERE id = ?", id)
	return err
}

func (store *SQLiteStore) Update(id string, opts ...UpdateOption) (*Note, error) {
	note, err := store.GetByID(id)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(note)
	}
	note.ModifiedAt = time.Now()

	tagsJSON, err := json.Marshal(note.Tags);
	if err != nil {
		return nil, err
	}

	_, err = store.db.Exec(`
		UPDATE notes
		SET title = ?, path = ?, project = ?, branch = ?, ticket = ?, tags = ?, modified_at = ?
		WHERE id = ?`,
		note.Title, note.Path, note.Project, note.Branch, note.Ticket,
		string(tagsJSON), note.ModifiedAt, id)

	if err != nil {
		return nil, err
	}
	return note, nil
}

func (store *SQLiteStore) GetByID(id string) (*Note, error) {
	var note Note
	var tagsJSON string

	err := store.db.QueryRow(`
		SELECT id, title, path, project, branch, ticket, tags, created_at, modified_at
		FROM notes WHERE id = ?`, id).Scan(
		&note.ID, &note.Title, &note.Path, &note.Project, &note.Branch,
		&note.Ticket, &tagsJSON, &note.CreatedAt, &note.ModifiedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note with id %s not found", id)
		}
		return nil, err
	}

	// Unmarshal tags JSON
	err = json.Unmarshal([]byte(tagsJSON), &note.Tags)
	if err != nil {
		return nil, err
	}

	return &note, nil
}

func (store *SQLiteStore) GetAll() ([]*Note, error) {
	rows, err := store.db.Query(`
		SELECT id, title, path, project, branch, ticket, tags, created_at, modified_at
		FROM notes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string

		err := rows.Scan(&note.ID, &note.Title, &note.Path, &note.Project,
			&note.Branch, &note.Ticket, &tagsJSON, &note.CreatedAt, &note.ModifiedAt)
		if err != nil {
			return nil, err
		}

		// Unmarshal tags JSON
		err = json.Unmarshal([]byte(tagsJSON), &note.Tags)
		if err != nil {
			return nil, err
		}

		notes = append(notes, &note)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (store *SQLiteStore) GetInProject(project string) ([]*Note, error) {
	rows, err := store.db.Query(`
		SELECT id, title, path, project, branch, ticket, tags, created_at, modified_at
		FROM notes WHERE project = ?`, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string

		err := rows.Scan(&note.ID, &note.Title, &note.Path, &note.Project,
			&note.Branch, &note.Ticket, &tagsJSON, &note.CreatedAt, &note.ModifiedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(tagsJSON), &note.Tags)
		if err != nil {
			return nil, err
		}

		notes = append(notes, &note)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (store *SQLiteStore) GetByBranch(branch string) ([]*Note, error) {
	rows, err := store.db.Query(`
		SELECT id, title, path, project, branch, ticket, tags, created_at, modified_at
		FROM notes WHERE branch = ?`, branch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string

		err := rows.Scan(&note.ID, &note.Title, &note.Path, &note.Project,
			&note.Branch, &note.Ticket, &tagsJSON, &note.CreatedAt, &note.ModifiedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(tagsJSON), &note.Tags)
		if err != nil {
			return nil, err
		}

		notes = append(notes, &note)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (store *SQLiteStore) GetByTicket(ticket string) ([]*Note, error) {
	rows, err := store.db.Query(`
		SELECT id, title, path, project, branch, ticket, tags, created_at, modified_at
		FROM notes WHERE ticket = ?`, ticket)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string

		err := rows.Scan(&note.ID, &note.Title, &note.Path, &note.Project,
			&note.Branch, &note.Ticket, &tagsJSON, &note.CreatedAt, &note.ModifiedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(tagsJSON), &note.Tags)
		if err != nil {
			return nil, err
		}

		notes = append(notes, &note)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (store *SQLiteStore) GetProjectMisc(project string) ([]*Note, error) {
	rows, err := store.db.Query(`
		SELECT id, title, path, project, branch, ticket, tags, created_at, modified_at
		FROM notes WHERE project = ? AND (branch = '' OR branch IS NULL)`, project)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*Note
	for rows.Next() {
		var note Note
		var tagsJSON string

		err := rows.Scan(&note.ID, &note.Title, &note.Path, &note.Project,
			&note.Branch, &note.Ticket, &tagsJSON, &note.CreatedAt, &note.ModifiedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(tagsJSON), &note.Tags)
		if err != nil {
			return nil, err
		}

		notes = append(notes, &note)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (store *SQLiteStore) Open(id string) error {
	note, err := store.GetByID(id)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", note.Path)
	case "darwin":
		cmd = exec.Command("open", note.Path)
	default: // linux
		cmd = exec.Command("xdg-open", note.Path)
	}

	return cmd.Start()
}
