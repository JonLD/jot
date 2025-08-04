# Jot

A smart note-taking CLI/TUI written in Go. Design for developers to quickly capture and organize notes by branch, project, and ticket.

## Features

- **Quick note capture** - Fast, frictionless note-taking from the command line
- **Smart organization** - Notes automatically organized by Git branch, project, and ticket
- **Terminal UI** - Clean, keyboard-driven interface for browsing and managing notes
- **Developer-focused** - Built for the development workflow

## Installation

### Download pre-built binaries

**Linux:**
```bash
# x64
curl -L https://github.com/JonLD/jot/releases/latest/download/jot-latest-linux-amd64.tar.gz | tar -xz && sudo mv jot-linux-amd64 /usr/local/bin/jot

# ARM64
curl -L https://github.com/JonLD/jot/releases/latest/download/jot-latest-linux-arm64.tar.gz | tar -xz && sudo mv jot-linux-arm64 /usr/local/bin/jot
```

**macOS (check with `uname -m`):**
```bash
# Intel Mac (x86_64)
curl -L https://github.com/JonLD/jot/releases/latest/download/jot-latest-darwin-amd64.tar.gz | tar -xz && sudo mv jot-darwin-amd64 /usr/local/bin/jot

# Apple Silicon (arm64)
curl -L https://github.com/JonLD/jot/releases/latest/download/jot-latest-darwin-arm64.tar.gz | tar -xz && sudo mv jot-darwin-arm64 /usr/local/bin/jot
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://github.com/JonLD/jot/releases/latest/download/jot-latest-windows-amd64.zip" -OutFile "jot.zip"
Expand-Archive -Path "jot.zip" -DestinationPath "."
Move-Item "jot-windows-amd64.exe" "$env:USERPROFILE\bin\jot.exe"
```

### Or install with Go

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

Notes will be opened in your default editor for Markdown files.

```bash
# Interactive TUI (default)
jot

# Open/create a note for current Git branch
jot -branch

# Open/create a note by title
jot -open "Bug fix notes"

```

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
