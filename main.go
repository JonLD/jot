package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "github.com/JonLD/jot/internal/storage"
    "github.com/JonLD/jot/internal/ui"
    "github.com/JonLD/jot/internal/config"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    var (
        openNote     = flag.String("open", "", "Open note by title or ID (creates if doesn't exist)")
        branchNote   = flag.Bool("branch", false, "Open the note for the current branch (creates if doesn't exist)")
        fromNvim     = flag.Bool("fromnvim", false, "Called from Neovim (internal flag)")
        setEditor    = flag.String("editor", "", "Set the editor command for opening notes (e.g., 'nvim', 'code', 'open -a \"Visual Studio Code\"')")
    )
    flag.Parse()

    // Handle setting editor config
    if *setEditor != "" {
        cfg, err := config.Load()
        if err != nil {
            log.Printf("Error loading config: %v\n", err)
            os.Exit(1)
        }
        cfg.Editor = *setEditor
        if err := cfg.Save(); err != nil {
            log.Printf("Error saving config: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("Editor set to: %s\n", *setEditor)
        return
    }

    // Initialize storage
    store, err := storage.NewSQLiteStore("")
    if err != nil {
        log.Printf("Error initializing storage: %v\n", err)
        os.Exit(1)
    }

    // Handle CLI commands
    if *branchNote {
        handleBranchNote(store, *fromNvim)
        return
    }

    if *openNote != "" {
        handleOpenNote(store, *openNote, *fromNvim)
        return
    }

    // Default: start TUI
    startTUI()
}

func startTUI() {
    store , err:= storage.NewSQLiteStore("")
    if err != nil {
        log.Printf("Error initializing storage: %v\n", err)
        panic(err)
    }
    model := ui.NewModel(store)
    p := tea.NewProgram(model)
    p.Run()
}

func handleOpenNote(store storage.NoteStore, query string, fromNvim bool) {
    // Try to find note by title first, then by ID
    notes, err := store.GetAll()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error fetching notes: %v\n", err)
        os.Exit(1)
    }

    var foundNote *storage.Note
    for _, note := range notes {
        if note.Title == query || note.ID == query {
            foundNote = note
            break
        }
    }

    // If note doesn't exist, create it
    if foundNote == nil {
        note := storage.Note{
            Title:   query,
            Project: getCurrentProject(),
            Branch:  getCurrentBranch(),
        }

        createdNote, err := store.Create(note)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error creating note: %v\n", err)
            os.Exit(1)
        }
        foundNote = createdNote
    }

    // Check if called from Neovim
    if fromNvim {
        // Inside Neovim - just output the path
        fmt.Print(foundNote.Path)
    } else {
        // Outside Neovim - open with default editor
        err := store.Open(foundNote.ID)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error opening note: %v\n", err)
            os.Exit(1)
        }
    }
}

func handleBranchNote(store storage.NoteStore, fromNvim bool) {
    project := getCurrentProject()
    branch := getCurrentBranch()

    // Try to find existing note for this project/branch combination
    notes, err := store.GetAll()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error fetching notes: %v\n", err)
        os.Exit(1)
    }

    var foundNote *storage.Note
    for _, note := range notes {
        if note.Project == project && note.Branch == branch {
            foundNote = note
            break
        }
    }

    // If note doesn't exist, create it
    if foundNote == nil {
        note := storage.Note{
            Title:   branch + " notes",
            Project: project,
            Branch:  branch,
        }

        createdNote, err := store.Create(note)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error creating branch note: %v\n", err)
            os.Exit(1)
        }
        foundNote = createdNote
    }

    // Check if called from Neovim
    if fromNvim {
        // Inside Neovim - just output the path
        fmt.Print(foundNote.Path)
    } else {
        // Outside Neovim - open with default editor
        err := store.Open(foundNote.ID)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error opening note: %v\n", err)
            os.Exit(1)
        }
    }
}

// Helper function to get current Git branch
func getCurrentBranch() string {
    cmd := exec.Command("git", "branch", "--show-current")
    output, err := cmd.Output()
    if err != nil {
        return "main" // fallback
    }
    return strings.TrimSpace(string(output))
}

// Helper function to get current Git project name
func getCurrentProject() string {
    // Try to get project name from git remote
    cmd := exec.Command("git", "remote", "get-url", "origin")
    output, err := cmd.Output()
    if err == nil {
        url := strings.TrimSpace(string(output))
        // Extract project name from URL (handle both SSH and HTTPS)
        if strings.Contains(url, "/") {
            parts := strings.Split(url, "/")
            projectName := parts[len(parts)-1]
            // Remove .git suffix if present
            projectName = strings.TrimSuffix(projectName, ".git")
            if projectName != "" {
                return projectName
            }
        }
    }

    // Fallback to directory name
    cwd, err := os.Getwd()
    if err != nil {
        return "unknown"
    }
    return filepath.Base(cwd)
}
