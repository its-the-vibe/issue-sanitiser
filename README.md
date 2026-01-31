# issue-sanitiser

A copilot agent to tidy up GitHub issue descriptions. This command-line tool uses GitHub Copilot to analyze and rewrite GitHub issues with better formatting, context, and clarity.

## Features

- Analyzes GitHub issues for improvements
- Fixes typos and grammatical errors
- Adds proper markdown formatting
- Enhances context and adds relevant links
- Suggests clear acceptance criteria
- Embeds expert agent instructions for consistent results

## Prerequisites

- Go 1.21 or higher (tested with Go 1.24)
- GitHub Copilot subscription
- GitHub Copilot CLI (authenticated)

## Installation

### From Source

```bash
git clone https://github.com/its-the-vibe/issue-sanitiser.git
cd issue-sanitiser
go build -o issue-sanitiser
```

### Using go install

```bash
go install github.com/its-the-vibe/issue-sanitiser@latest
```

## Usage

Simply provide a GitHub issue URL as an argument:

```bash
./issue-sanitiser https://github.com/owner/repo/issues/123
```

Or if installed via go install:

```bash
issue-sanitiser https://github.com/owner/repo/issues/123
```

The tool will:
1. Parse the issue URL
2. Fetch the issue details
3. Analyze the content
4. Generate a sanitized version with improvements
5. Display the result in your terminal

## Example

```bash
$ ./issue-sanitiser https://github.com/example/project/issues/42

Analyzing issue: https://github.com/example/project/issues/42

# Sanitised Issue

## Suggested Title
Fix authentication timeout in production environment

## Sanitised Body
[Complete rewritten issue with proper formatting, context, and links]

---

## Key Improvements Made
- Fixed typos in original description
- Added reproduction steps
- Linked to authentication module code
- Added acceptance criteria
- Included environment details

✅ Issue sanitisation complete!
```

## How It Works

The tool embeds the agent description from `.github/agents/issue-sanitiser-command.agent.md` directly into the binary, so it can run from anywhere. The agent uses GitHub Copilot's API to:

- Read the issue content
- Search the codebase for context
- Find relevant documentation
- Rewrite the issue following best practices

## Development

```bash
# Run without building
go run main.go https://github.com/owner/repo/issues/123

# Build
go build -o issue-sanitiser

# Test
./issue-sanitiser https://github.com/owner/repo/issues/123
```

## License

See repository license.
