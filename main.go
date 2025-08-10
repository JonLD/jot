package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "github.com/JonLD/jot/internal/storage"
    "github.com/JonLD/jot/internal/ui"
    "github.com/JonLD/jot/internal/config"

    "github.com/spf13/cobra"
    tea "github.com/charmbracelet/bubbletea"
)

type CLIFlags struct {
    OpenNote     string
    BranchNote   bool
    FromNvim     bool
}

type ConfigFlags struct {
    Editor           string
    EditorBackground string
    DefaultMode      string
}

var (
	fromNvim bool
    cliFlags    = &CLIFlags{}
    configFlags = &ConfigFlags{}
	cfg *config.Config
)

var rootCmd = &cobra.Command{
    Use:   "jot",
    Short: "A smart note-taking CLI for developers",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Default action - start TUI
		store, err := initializeApp()
		if err != nil {
			return err
		}
        return runJot(store, ui.FilterDisplayAll)
    },
}

var openCmd = &cobra.Command{
    Use:   "open [note-title]",
    Short: "Open or create a note by title",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        store, err := initializeApp()
        if err != nil {
            return err
        }
		project := getCurrentProject()
		branch := getCurrentBranch()
        return handleOpenNote(store, args[0], project, branch, fromNvim)
    },
}

var branchCmd = &cobra.Command{
    Use:   "branch",
    Short: "Open or create a note for the current Git branch",
    RunE: func(cmd *cobra.Command, args []string) error {
        store, err := initializeApp()
        if err != nil {
            return err
        }

		if len(args) == 1 {
            branch := getCurrentBranch()
            project := getCurrentProject()
			return handleOpenNote(store, args[0], project, branch, fromNvim)
		}
		project := getCurrentProject()
		branch := getCurrentBranch()
        return handleContextNote(store, project, branch, fromNvim)
    },
}

var projectCmd = &cobra.Command{
    Use:   "proj [note-title]",
    Short: "Open or create a project-wide note",
    Args:  cobra.MaximumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        store, err := initializeApp()
        if err != nil {
            return err
        }

        project := getCurrentProject()

        if len(args) == 1 {
            return handleOpenNote(store, args[0], project, "*", fromNvim)
        }

        return handleContextNote(store, project, "*", fromNvim)
    },
}

func initializeApp() (storage.NoteStore, error) {
	store, err := storage.NewSQLiteStore("", cfg)
	if err != nil {
		return nil, fmt.Errorf("error initializing storage: %w", err)
	}

	return store, nil
}

func init() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		log.Printf("Error loading config, using defaults: %v\n", err)
		cfg = &config.Config{}
	}
    // CLI flags
    rootCmd.Flags().StringVarP(&cliFlags.OpenNote,
        "open", "o", "", "Open note by title or ID")
    rootCmd.Flags().BoolVarP(&cliFlags.BranchNote,
        "branch", "b", false, "Open note for current branch")

    // Config flags
    rootCmd.Flags().StringVarP(&configFlags.Editor,
        "editor", "e", "", "Set the editor command")
    rootCmd.Flags().StringVarP(&configFlags.EditorBackground,
        "editor-background", "", "", "Set editor background mode")
    rootCmd.Flags().StringVarP(&configFlags.DefaultMode, "default-mode", "m", "", "Set default mode")

	rootCmd.PersistentFlags().BoolVar(&fromNvim, "fromnvim", false, "Called from Neovim (internal)")

	rootCmd.AddCommand(openCmd)
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(projectCmd)
}

func runJot(store storage.NoteStore, filter ui.FilterFunc) error {
    // Handle config updates first
    if hasConfigFlags(configFlags) {
        if err := updateConfigFromFlags(configFlags); err != nil {
            return fmt.Errorf("error updating config: %w", err)
        }
        fmt.Println("Configuration updated successfully")
        return nil
    }
    startTUI(store, filter)
    return nil
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func hasConfigFlags(flags *ConfigFlags) bool {
    return flags.Editor != "" || flags.EditorBackground != "" || flags.DefaultMode != ""
}

func updateConfigFromFlags(flags *ConfigFlags) error {
    modified := false

    if flags.Editor != "" {
        cfg.Editor = flags.Editor
        modified = true
    }

    if flags.EditorBackground != "" {
        if flags.EditorBackground == "true" || flags.EditorBackground == "false" {
            cfg.EditorBackground = flags.EditorBackground == "true"
            modified = true
        } else {
            return fmt.Errorf("invalid value for editor-background: %s", flags.EditorBackground)
        }
    }

    if flags.DefaultMode != "" {
        if flags.DefaultMode == "normal" || flags.DefaultMode == "search" {
            cfg.DefaultMode = flags.DefaultMode
            modified = true
        } else {
            return fmt.Errorf("invalid default mode: %s", flags.DefaultMode)
        }
    }
    if modified {
        return cfg.Save()
    }
    return nil
}

func startTUI(store storage.NoteStore, filter ui.FilterFunc) {
    model := ui.NewModel(store, filter)
    p := tea.NewProgram(model)
    p.Run()
}

func handleOpenNote(
	store storage.NoteStore,
	query string,
	project string,
	branch string,
	fromNvim bool,
) error {
    // Try to find note by title first, then by ID
    notes, err := store.GetAll()
    if err != nil {
        return fmt.Errorf("error fetching notes: %v", err)
    }

    var foundNote *storage.Note
    for _, note := range notes {
		matchingDetails := note.Title == query && note.Project == project && note.Branch == branch
        if (matchingDetails) || note.ID == query {
            foundNote = note
            break
        }
    }

    // If note doesn't exist, create it
    if foundNote == nil {
        note := storage.Note{
            Title:   query,
            Project: project,
            Branch:  branch,
        }

        createdNote, err := store.Create(note)
        if err != nil {
            return fmt.Errorf("error creating note: %v", err)
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
            return fmt.Errorf( "error opening note: %v", err)
        }
    }
	return nil
}

func handleContextNote(store storage.NoteStore, project string, branch string, fromNvim bool) error {
    // Try to find existing note for this project/branch combination
    notes, err := store.GetAll()
    if err != nil {
        return fmt.Errorf("error fetching notes: %v", err)
    }

    var foundNotes []*storage.Note
    for _, note := range notes {
        if note.Project == project && note.Branch == branch {
            foundNotes = append(foundNotes, note)
        }
    }

    // If note doesn't exist, create it
    if len(foundNotes) == 0 {
        defaultTitle := branch
        if branch == "*" {
            defaultTitle = project
        }

        note := storage.Note{
            Title:   defaultTitle,
            Project: project,
            Branch:  branch,
        }

        createdNote, err := store.Create(note)
        if err != nil {
            return fmt.Errorf("error creating note: %v", err)
        }
        foundNotes = append(foundNotes, createdNote)
    }

	// If multiple notes then open TUI with appropriate filter
	if len(foundNotes) > 1 {
        var filter ui.FilterFunc
        if branch == "*" {
            filter = ui.FilterByProject
        } else {
            filter = ui.FilterByBranch
        }
		return runJot(store, filter)
	} else { // Only one note so open it immediately
		// Check if called from Neovim
		if fromNvim {
			// Just output path for jot.nvim to open
			fmt.Print(foundNotes[0].Path)
		} else {
			// Outside Neovim - open with default editor
			if err := store.Open(foundNotes[0].ID); err != nil {
				return fmt.Errorf("error opening note: %v", err)
			}
		}
	}

    return nil
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
