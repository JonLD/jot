package main

import (
    "github.com/JonLD/scrib/internal/storage"
    "github.com/JonLD/scrib/internal/ui"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    startTUI()
}

func startTUI() {
    store := storage.NewInMemoryStore()
    testNote1 := storage.Note{Title: "Bug investigation", Project: "scrib", Branch: "main"}
    testNote2 := storage.Note{Title: "Feature planning", Project: "scrib", Branch: "feature/tui"}
    store.Create(testNote1)
    store.Create(testNote2)
    model := ui.NewModel(store)
    p := tea.NewProgram(model)
    p.Run()
}
