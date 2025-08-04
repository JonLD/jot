package ui

import (
    "log"
    "strings"
    "os/exec"
    "path/filepath"
    "os"

    "github.com/JonLD/jot/internal/storage"
    "github.com/JonLD/jot/themes"

    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"
)

// Current theme TODO: add into a config file
var currentTheme = themes.TokyoNightScheme

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

var (
    primaryStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(currentTheme.PrimaryFg))

    selectedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(currentTheme.SelectedFg)).Bold(true)

    mutedStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(currentTheme.MutedFg))

    errorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(currentTheme.ErrorFg))

    windowStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(currentTheme.DefaultFg)).
        BorderForeground(lipgloss.Color(currentTheme.BorderColor)).
        Background(lipgloss.Color(currentTheme.DefaultBg)).
        Border(lipgloss.RoundedBorder())

    popupStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(currentTheme.DefaultFg)).
        BorderForeground(lipgloss.Color(currentTheme.BorderColor)).
        Background(lipgloss.Color(currentTheme.PopupBg)).
        Border(lipgloss.RoundedBorder())
)

func NewModel(store storage.NoteStore) Model {
    textInput := textinput.New() // Creates the text input component
    textInput.Placeholder = "Enter note title..."
    textInput.CharLimit = 100
    textInput.Width = 40
    textInput.SetValue("")
    textInput.Focus()
    // ti.Cursor.Style = cursorStyle
    textInput.Blur()

    return Model{
        Store:     store,
        InputText: textInput,
        ShowInput: false,
    }
}

type Model struct {
    Store             storage.NoteStore
    Notes             []*storage.Note
    Cursor            int
    ShowInput         bool
    InputText         textinput.Model
    ShowDeleteConfirm bool
    DeleteNoteID      string
    DeleteNoteTitle   string
}

type notesLoadedMsg struct {
    notes []*storage.Note
}

func (model Model) Init() tea.Cmd {
    return tea.Batch(
        tea.Cmd(func() tea.Msg {
            notes, _ := model.Store.GetAll()
            return notesLoadedMsg{notes}
        }),
        textinput.Blink,
        )
}

func (model Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if model.ShowDeleteConfirm {
            switch msg.String() {
            case "y", "Y":
                // Perform deletion
                err := model.Store.Delete(model.DeleteNoteID)
                if err != nil {
                    log.Printf("Error deleting note: %v", err)
                }
                model.ShowDeleteConfirm = false
                // Reload notes
                cmd := tea.Cmd(func() tea.Msg {
                    notes, _ := model.Store.GetAll()
                    return notesLoadedMsg{notes}
                })
                return model, cmd
            case "n", "N", "esc":
                model.ShowDeleteConfirm = false
                return model, nil
            }
            return model, nil
        } else if model.ShowInput {
            switch msg.Type {
            case tea.KeyEnter, tea.KeyCtrlL:
                title := model.InputText.Value()
                if title != "" {
                    newNote := storage.Note{
                        Title:   title,
                        Project: getCurrentProject(),
                        Branch:  getCurrentBranch(),
                    }
                    createdNote, err := model.Store.Create(newNote)
                    if err != nil {
                        log.Printf("Error creating note: %v", err)
                    } else {
                        model.Notes = append(model.Notes, createdNote)
                        // Open the newly created note
                        if err := model.Store.Open(createdNote.ID); err != nil {
                            log.Printf("Error opening created note: %v", err)
                        }
                    }
                }
                model.ShowInput = false
                model.InputText.Reset()
                model.InputText.Blur()
                return model, nil
            case tea.KeyCtrlC, tea.KeyEsc:
                model.ShowInput = false
                model.InputText.Blur()
                model.InputText.Reset()
                return model, nil
            }
            var cmd tea.Cmd
            model.InputText, cmd = model.InputText.Update(msg)
            return model, cmd
        } else {
            switch msg.String() {
            case "n":
                model.ShowInput = true
                return model, model.InputText.Focus()
            case "d":
                if len(model.Notes) > 0 && model.Cursor < len(model.Notes) {
                    selectedNote := model.Notes[model.Cursor]
                    // Set up confirmation dialog
                    model.ShowDeleteConfirm = true
                    model.DeleteNoteID = selectedNote.ID
                    model.DeleteNoteTitle = selectedNote.Title
                    return model, nil
                }
            case "q", tea.KeyCtrlC.String(), tea.KeyEsc.String():
                return model, tea.Quit
            case "j", tea.KeyDown.String():
                if model.Cursor < len(model.Notes)-1 {
                    model.Cursor++
                }
            case "k", tea.KeyUp.String():
                if model.Cursor > 0 {
                    model.Cursor--
                }
            case tea.KeyCtrlL.String(), "enter":
                if len(model.Notes) > 0 && model.Cursor < len(model.Notes) {
                    selectedNote := model.Notes[model.Cursor]
                    model.Store.Open(selectedNote.ID)
                }
            }
        }
    case notesLoadedMsg:
        model.Notes = msg.notes
    }

    return model, nil
}

func (model Model) View() string {
    listStyle := windowStyle.
        Padding(1, 2)

    var listContent strings.Builder
    listContent.WriteString(primaryStyle.Bold(true).Render("Jot Notes") + "\n")

    for i, note := range model.Notes {
        if model.Cursor == i {
            listContent.WriteString(selectedStyle.Render("▶ "+
                note.Title) + "\n")
        } else {
            listContent.WriteString(mutedStyle.Render("  "+
                note.Title) + "\n")
        }
    }

    listContent.WriteString("\n" + mutedStyle.Render("Press 'n' for new note, 'q' to quit"))
    mainView := listStyle.Render(listContent.String())

    if model.ShowInput {
        popupContent := popupStyle.
            Padding(1, 2).
            Width(50).
            Align(lipgloss.Center).
            Render(
            "New Note\n\n" +
            model.InputText.View() + "\n\n" +
            "Press Ctrl-l or Enter to save, Ctrl-c or Esc to cancel",
            )

        return lipgloss.Place(
            80, 24,
            lipgloss.Center, lipgloss.Center,
            popupContent,
            )
    }

    if model.ShowDeleteConfirm {
        confirmContent := popupStyle.
            Padding(1, 2).
            Width(50).
            Align(lipgloss.Center).
            BorderForeground(lipgloss.Color(currentTheme.ErrorFg)).
            Render(
            "Delete Confirmation\n\n" +
            "Delete note '" + model.DeleteNoteTitle + "'?\n\n" +
            "[Y]es / [N]o",
            )

        return lipgloss.Place(
            80, 24,
            lipgloss.Center, lipgloss.Center,
            confirmContent,
            )
    }

    return mainView
}
