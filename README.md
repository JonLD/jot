# Jot

A smart note-taking CLI/TUI written in Go. Design for developers to quickly capture and organize notes by branch, project, and ticket.

## Features

- **Quick note capture** - Fast, frictionless note-taking from the command line
- **Smart organization** - Notes automatically organized by Git branch, project, and ticket
- **Terminal UI** - Clean, keyboard-driven interface for browsing and managing notes
- **Developer-focused** - Built for the development workflow

## Installation

### Install with Go

```bash
go install github.com/JonLD/jot@latest
```

### Or build from source

```bash
git clone https://github.com/JonLD/jot.git
cd jot
go build
```

## Usage

```bash
# Interactive TUI (default)
jot

# Open/create a note for current Git branch
jot -branch

# Open/create a note by title
jot -open "Bug fix notes"
```

## Configuration

### Editor

By default, notes open in your system's default Markdown editor. To use a specific editor:

```bash
# Set Neovim as your editor
jot -editor "nvim"

# Set VS Code
jot -editor "code"

# Set Vim
jot -editor "vim"

# macOS: Set VS Code using open command
jot -editor "open -a 'Visual Studio Code'"
```

Configuration is saved to `~/.jot/config.json`.

## Neovim Integration

For seamless note-taking from within Neovim, check out the companion plugin:

**[JonLD/jot.nvim](https://github.com/JonLD/jot.nvim)**

## Development

```bash
# Run the application
go run .

# Build
go build

# Run tests
go test ./...

# Format code
go fmt ./...
```

## License

MIT
