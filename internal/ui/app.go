package ui

import (
    "log"
    "strings"

    "github.com/JonLD/scrib/internal/storage"
    "github.com/JonLD/scrib/themes"

    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"
)

// Current theme TODO: add into a config file
var currentTheme = themes.TokyoNightScheme

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
    ti := textinput.New() // Creates the text input component
    ti.Placeholder = "Enter note title..."
    ti.CharLimit = 100
    ti.Width = 40
    ti.SetValue("")
    ti.Focus()
    // ti.Cursor.Style = cursorStyle
    ti.Blur()

    return Model{
        Store:     store,
        InputText: ti,
        ShowInput: false,
    }
}

type Model struct {
    Store     storage.NoteStore
    Notes     []*storage.Note
    Cursor    int
    ShowInput bool
    InputText textinput.Model
}

type notesLoadedMsg struct {
    notes []*storage.Note
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        tea.Cmd(func() tea.Msg {
            notes, _ := m.Store.GetAll()
            return notesLoadedMsg{notes}
        }),
        textinput.Blink,
        )
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.ShowInput {
            switch msg.Type {
            case tea.KeyEnter, tea.KeyCtrlL:
                title := m.InputText.Value()
                if title != "" {
                    newNote := storage.Note{
                        Title:   title,
                        Project: "scrib",
                        Branch:  "main",
                    }
                    createdNote, err := m.Store.Create(newNote)
                    if err != nil {
                        log.Printf("Error creating note: %v", err)
                    } else {
                        m.Notes = append(m.Notes, createdNote)
                    }
                }
                m.ShowInput = false
                m.InputText.Reset()
                m.InputText.Blur()
                return m, nil
            case tea.KeyCtrlC, tea.KeyEsc:
                m.ShowInput = false
                m.InputText.Blur()
                m.InputText.Reset()
                return m, nil
            }
            var cmd tea.Cmd
            m.InputText, cmd = m.InputText.Update(msg)
            return m, cmd
        } else {
            switch msg.String() {
            case "n":
                m.ShowInput = true
                return m, m.InputText.Focus()
            case "q", tea.KeyCtrlC.String(), tea.KeyEsc.String():
                return m, tea.Quit
            case "j", tea.KeyDown.String():
                if m.Cursor < len(m.Notes)-1 {
                    m.Cursor++
                }
            case "k", tea.KeyUp.String():
                if m.Cursor > 0 {
                    m.Cursor--
                }
            case tea.KeyCtrlL.String():
                if len(m.Notes) > 0 && m.Cursor < len(m.Notes) {
                    selectedNote := m.Notes[m.Cursor]
                    // TODO: Open note in editor
                    log.Printf("Opening note: %s", selectedNote.Title)
                }
            }
        }
    case notesLoadedMsg:
        m.Notes = msg.notes
    }

    return m, nil
}

func (m Model) View() string {
    listStyle := windowStyle.
        Padding(1, 2)

    var listContent strings.Builder
    listContent.WriteString(primaryStyle.Bold(true).Render("Scrib Notes") + "\n")

    for i, note := range m.Notes {
        if m.Cursor == i {
            listContent.WriteString(selectedStyle.Render("▶ "+
                note.Title) + "\n")
        } else {
            listContent.WriteString(mutedStyle.Render("  "+
                note.Title) + "\n")
        }
    }

    listContent.WriteString("\n" + mutedStyle.Render("Press 'n' for new note, 'q' to quit"))
    mainView := listStyle.Render(listContent.String())

    if m.ShowInput {
        popupContent := popupStyle.
            Padding(1, 2).
            Width(50).
            Align(lipgloss.Center).
            Render(
            "New Note\n\n" +
            m.InputText.View() + "\n\n" +
            "Press Ctrl-l or Enter to save, Ctrl-c or Esc to cancel",
            )

        return lipgloss.Place(
            80, 24,
            lipgloss.Center, lipgloss.Center,
            popupContent,
            )
    }
    return mainView
}
