# ✂️ Snip

```text
 _____       _
/  ___|     (_)
\ `--. _ __  _ _ __
 `--. \ '_ \| | '_ \
/\__/ / | | | | |_) |
\____/|_| |_|_| .__/
              | |
              |_|
              get just what you need.
```

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/mukailasam/snip)](https://goreportcard.com/report/github.com/mukailasam/snip)

Snip is a command-line tool that lets you download a specific folder or file from a GitHub repository without cloning the entire project.

_If you only need one example, one subfolder, or a single config file, Snip gets it for you directly._

## Features

- Download a specific folder or a single file

- Works for both public and private GitHub repositories

- No Git clone - uses the GitHub REST API directly

- Messages for rate limits and large repositories

- Automatically detects the repo’s default branch

- Built with [Cobra](https://github.com/spf13/cobra) for a clean CLI

## Installation

Install it from the sources:

```bash
git clone https://github.com/mukailasam/snip
cd snip
go install
```

Install it from the repository:

```bash
go install github.com/mukailasam/snip
```

## Usage

🗂️ Snip a folder

```
$ snip repo github.com/mukailasam/codelab --dir linear
```

📄 Snip a single file

```
$ snip repo github.com/mukailasam/codelab --file avl.go
```

📂 Specify a destination (optional)

```
$ snip repo github.com/mukailasam/codelab --dir array --dest /users/sam/desktop/workspace/temp
```

If you don’t specify a destination using --dest, Snip saves the result in your current working directory.

## 🔐 Private Repositories

To access private repositories, set your token via environment variable:

```
macOS / Linux (bash/zsh)
export GITHUB_TOKEN=ghp_xxxxxxxxx

Windows PowerShell
$env:GITHUB_TOKEN = "ghp_xxxxxxxxx"
```

Snip automatically detects your token and uses it for private repo access.

## Project Structure

```text
cmd/
├── helper.go
├── root.go         # Cobra CLI entrypoint
├── repo.go         # Handles 'snip repo' command
internal/
├── provider/
│   ├── github.go   # GitHub API implementation
utils/
|    └── utils.go
main.go             # Main entrypoint
```

## How Snip Works

- Parses the provided repo URL

- Detects whether the repo is public or private.

- Retrieves the default branch (if not specified).

- Lists all files and directories via GitHub’s Tree API

- Searches for the requested file or folder name.

- Downloads matching content into the specified (or default) destination.

Snip does not clone the repository or use Git — it talks directly to the provider’s REST API, keeping things lightweight and fast.

## Example Output

```
$ go run snip.go repo github.com/mukailasam/codelab --file avl.go

🌍 Public repository detected.
🔍 Listing repository tree for mukailasam/codelab (branch: main)...
⬇️ Downloaded: avl.go
✅ Done.
```

## Why Snip?

Cloning an entire repository just to get one file or folder is inefficient.
Snip saves time, bandwidth, and storage by letting you fetch only what you need directly from the GitHub API.

It’s perfect for developers who want a quick way to grab example files, configs, or small components without dealing with full clones or large repo histories.

## Inspiration

I needed a tool to quickly grab a single project, subfolder, configuration, or component from my <a href="https://github.com/mukailasam/Codelab" style="color: grey;">Codelab</a> without cloning the entire repository.

## Contributing

Contributions are welcome!
To contribute:

- Fork the repository

- Create a new branch

- Implement your feature or fix

- Open a pull request

Make sure to follow Go’s code formatting and keep your commits clean and descriptive.

## License

Snip is released under the MIT License. See LICENSE
