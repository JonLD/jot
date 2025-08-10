package ui

import (
    "log"
    "fmt"
    "strings"
    "os/exec"
    "path/filepath"
    "os"

    "github.com/JonLD/jot/internal/storage"
    "github.com/JonLD/jot/themes"

    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/sahilm/fuzzy"
)

// Current theme TODO: add into a config file
var currentTheme = themes.TokyoNightScheme

type State int

const (
    StateNormal State = iota
    StateSearch
    StateNewNote
    StateDeleteConfirm
)

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
        Border(lipgloss.RoundedBorder())

    popupStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color(currentTheme.DefaultFg)).
        BorderForeground(lipgloss.Color(currentTheme.BorderColor)).
        Border(lipgloss.RoundedBorder())
)

func NewModel(store storage.NoteStore) Model {
    newNoteTextInput := textinput.New() // Creates the text input component
    newNoteTextInput.Placeholder = "Enter note title..."
    newNoteTextInput.CharLimit = 100
    newNoteTextInput.Width = 40
    newNoteTextInput.SetValue("")
    newNoteTextInput.Focus()
    newNoteTextInput.Blur()

    searchTextInput := textinput.New() // Creates the text input component
    searchTextInput.CharLimit = 100
    searchTextInput.Width = 40
    searchTextInput.SetValue("")
    searchTextInput.Focus()


    return Model{
        Store:     store,
        NewNoteInputText: newNoteTextInput,
        SearchInputText: searchTextInput,
        State: StateSearch,
    }
}

type Model struct {
    Store             storage.NoteStore
    Notes             []*storage.Note
    FilteredNotes     []*storage.Note
    Cursor            int
    NewNoteInputText  textinput.Model
    SearchInputText   textinput.Model
    State             State
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
        switch model.State {
        case StateDeleteConfirm:
            return model.updateDeleteMode(msg)
        case StateNewNote:
            return model.updateNewNoteMode(msg)
        case StateNormal:
            return model.updateNormalMode(msg)
        case StateSearch:
            return model.updateSearchMode(msg)
        }
    case notesLoadedMsg:
        model.Notes = msg.notes
        model.FilteredNotes = msg.notes
    }
    return model, nil
}

func (model Model) updateNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "n":
        model.State = StateNewNote
        return model, model.NewNoteInputText.Focus()
    case "i":
        model.State = StateSearch
        return model, model.SearchInputText.Focus()

    case "d":
        if len(model.Notes) > 0 && model.Cursor < len(model.Notes) {
            selectedNote := model.Notes[model.Cursor]
            // Set up confirmation dialog
            model.State = StateDeleteConfirm
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
    return model, nil
}

func (model Model) updateSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.Type {
    case tea.KeyEsc:
        model.State = StateNormal
        model.SearchInputText.Blur()
        return model, nil
    case tea.KeyCtrlC:
        return model, tea.Quit
    case tea.KeyCtrlJ:
        if model.Cursor < len(model.Notes)-1 {
            model.Cursor++
        }
        return model, nil
    case tea.KeyCtrlK:
        if model.Cursor > 0 {
            model.Cursor--
        }
        return model, nil
    case tea.KeyCtrlL:
        if len(model.Notes) > 0 && model.Cursor < len(model.Notes) {
            selectedNote := model.Notes[model.Cursor]
            model.Store.Open(selectedNote.ID)
        }
        return model, nil
    }

    // First empty the filtered notes so can later append all matches
    model.FilteredNotes = nil

    var cmd tea.Cmd
    model.SearchInputText, cmd = model.SearchInputText.Update(msg)
    var searchQuery = model.SearchInputText.Value()
    if searchQuery == "" {
        model.FilteredNotes = model.Notes
    }

    var titles []string
    for _, note := range model.Notes {
        titles = append(titles, note.Title)
    }

    matches := fuzzy.Find(searchQuery, titles)
    for _, match := range matches {
        model.FilteredNotes = append(model.FilteredNotes, model.Notes[match.Index])
    }
    return model, cmd
}

func (model Model) updateNewNoteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.Type {
    case tea.KeyEnter, tea.KeyCtrlL:
        title := model.NewNoteInputText.Value()
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
        model.State = StateNormal
        model.NewNoteInputText.Reset()
        model.NewNoteInputText.Blur()
        return model, nil

    case tea.KeyCtrlC, tea.KeyEsc:
        model.State = StateNormal
        model.NewNoteInputText.Blur()
        model.NewNoteInputText.Reset()
        return model, nil
    }

    var cmd tea.Cmd
    model.NewNoteInputText, cmd = model.NewNoteInputText.Update(msg)
    return model, cmd
}

func (model Model) updateDeleteMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "y", "Y":
        // Perform deletion
        err := model.Store.Delete(model.DeleteNoteID)
        if err != nil {
            log.Printf("Error deleting note: %v", err)
        }
        model.State = StateNormal
        // Reload notes
        cmd := tea.Cmd(func() tea.Msg {
            notes, _ := model.Store.GetAll()
            return notesLoadedMsg{notes}
        })
        return model, cmd
    case "n", "N", "esc":
        model.State = StateNormal
        return model, nil
    }
    return model, nil
}


func (model Model) View() string {
    listStyle := windowStyle.
        Padding(1, 2)

    var listContent strings.Builder
    asciiArt := `
  ____   ___   ______
 |    | /   \ |      |
 |__  ||     ||      |
 __|  ||  O  ||_|  |_|
/  |  ||     |  |  |
\  `+ "`"+ `  ||     |  |  |
 \____| \___/   |__|
`
    listContent.WriteString(primaryStyle.Bold(true).Render(asciiArt) + "\n")

    searchBarStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color(currentTheme.PrimaryFg))

    if model.State == StateSearch {
        searchBarStyle = searchBarStyle.Bold(true)
    }

    listContent.WriteString("\n")
    listContent.WriteString(searchBarStyle.Render("Search: ") + model.SearchInputText.View())
    // listContent.WriteString("\n\n")

    // Show filtered results count if searching
    if model.SearchInputText.Value() != "" {
        countStyle := mutedStyle
        listContent.WriteString(countStyle.Render(fmt.Sprintf("(%d/%d)", len(model.FilteredNotes), len(model.Notes))))
    }
    listContent.WriteString("\n\n")


    for i, note := range model.FilteredNotes {
        if model.Cursor == i {
            listContent.WriteString(selectedStyle.Render("▶ "+
                note.Title) + "\n")
        } else {
            listContent.WriteString(mutedStyle.Render("  "+
                note.Title) + "\n")
        }
    }
    helpText := "i: search, j/k: navigate, Enter: open, n: new, d: delete, q: quit"
    if model.State == StateSearch {
          helpText = "Type to search, Esc: exit search mode"
      }
    listContent.WriteString("\n" + mutedStyle.Render(helpText))
    mainView := listStyle.Render(listContent.String())

    if model.State == StateNewNote {
        popupContent := popupStyle.
            Padding(1, 2).
            Width(50).
            Align(lipgloss.Center).
            Render(
            "New Note\n\n" +
            model.NewNoteInputText.View() + "\n\n" +
            "Press Ctrl-l or Enter to save, Ctrl-c or Esc to cancel",
            )

        return lipgloss.Place(
            lipgloss.Width(mainView),
            lipgloss.Height(mainView),
            lipgloss.Center,
            lipgloss.Center,
            popupContent,
        )
    }

    if model.State == StateDeleteConfirm {
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
            lipgloss.Width(mainView),
            lipgloss.Height(mainView),
            lipgloss.Center,
            lipgloss.Center,
            confirmContent,
            )
    }

    return mainView
}
