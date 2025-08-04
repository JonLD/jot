package ui

import (
    "fmt"

    "github.com/charmbracelet/bubbles/textinput"
    "github.com/JonLD/scrib/internal/storage"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"
)

func NewModel(store storage.NoteStore) Model {
    ti := textinput.New()  // Creates the text input component
    ti.Placeholder = "Enter note title..."
    ti.CharLimit = 100
    ti.Width = 40
    ti.SetValue("")
    ti.Focus()
    ti.Blur()

    return Model{
        Store:          store,
        InputText:      ti,
        ShowInput: false,
    }
}

type Model struct {
    Store storage.NoteStore
    Notes []*storage.Note
    Cursor int
    ShowInput bool
    InputText textinput.Model
}

type notesLoadedMsg struct {
    notes []*storage.Note
}

func (m Model) Init() tea.Cmd {
    return tea.Cmd(func() tea.Msg {
        notes, _ := m.Store.GetAll()
        return notesLoadedMsg{notes}
    })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
switch msg := msg.(type) {
case tea.KeyMsg:
    if m.ShowInput {
        switch msg.String() {
        case "enter":
            m.ShowInput = false
            m.InputText.Reset()
            return m, nil
        case "esc":
            m.ShowInput = false
            m.InputText.Reset()
            return m, nil
        }
        inputText, cmd := m.InputText.Update(msg)
        m.InputText = inputText
        return m, cmd
    } else {
            switch msg.String() {
            case "n":
                m.ShowInput = true
                m.InputText.Focus()
            case "q", "ctrl+c":
                return m, tea.Quit
            case "j", "down":
                if m.Cursor < len(m.Notes) - 1 {
                    m.Cursor++
                }
            case "k", "up":
                if m.Cursor > 0 {
                    m.Cursor--
                }
        }
    }
    case notesLoadedMsg:
        m.Notes = msg.notes
    }

    return m, nil
}

func (m Model) View() string {
    s := "Scrib\n\n"

    for i, note := range m.Notes {
        cursor := " "
        if m.Cursor == i {
            cursor = ">"
        }

            s += fmt.Sprintf("%s %s\n", cursor, note.Title)
    }
    s += "\nPress q to quit"

    if m.ShowInput {
        popupStyle := lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("62")).
            Padding(1, 2).
            Width(50).
            Align(lipgloss.Center)

        popup := popupStyle.Render(
            "New Note\n\n" +
            m.InputText.View() + "\n\n" +
            "Press enter to save, Esc to cancel",
        )

        return lipgloss.Place(
            80, 24,
            lipgloss.Center, lipgloss.Center,
            popup,
            lipgloss.WithWhitespaceChars(" "),
            lipgloss.WithWhitespaceForeground(lipgloss.Color("240")),
            ) + "\n" + s
    }
    return s
}
