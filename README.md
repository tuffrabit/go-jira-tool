# go-jira-tool

A lightweight CLI tool for interacting with Jira Cloud. Designed to be called by both humans and AI agents, abstracting away API authentication and Jira's internal document format so callers can work with plain text and JSON.

The entire existance of this tool is predicated on the official JIRA cli tool not supporting downloading ticket attachments. Once the official tool does, this project is pointless.

## Features

- **List tickets** — Search with raw JQL, get structured JSON output
- **Get ticket details** — Downloads a ticket as a markdown file with all attachments resolved to local files
- **Create tickets** — Create new issues from plain text descriptions
- **Transition tickets** — Move an issue to a new status (e.g. "In Progress", "Done")
- **Comment on tickets** — Add a plain-text comment to an issue, inline or from a file

## Installation

### Option A: go install

Requires Go 1.25+.

```bash
go install github.com/tuffrabit/go-jira-tool@latest
```

### Option B: Build from source

```bash
git clone https://github.com/tuffrabit/go-jira-tool.git
cd go-jira-tool
go build
```

## Setup

### 1. Create an API token

1. Go to [id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click **Create API token**
3. Give it a label (e.g. "go-jira-tool")
4. Select these scopes:
   - `read:jira-work` — read issues, comments, attachments, and JQL search
   - `write:jira-work` — create/transition issues, add comments, assign
   - `read:jira-user` — look up the current user (used by `-assign`)

   If your Atlassian UI doesn't offer a scope picker, the classic (unscoped) token also works but grants the token your full Jira permissions.
5. Copy the generated token

### 2. Configure

Create a `config.json` file at `~/.go-jira-tool/config.json`:

```bash
mkdir -p ~/.go-jira-tool
```

```json
{
  "base_url": "https://your-org.atlassian.net",
  "email": "you@example.com",
  "api_token": "your-api-token",
  "default_project": "PROJ"
}
```

The tool looks for `config.json` in the following order:

1. `~/.go-jira-tool/config.json`
2. Next to the executable (for portable/self-contained setups)

| Field | Required | Description |
|-------|----------|-------------|
| `base_url` | Yes | Your Jira Cloud instance URL |
| `email` | Yes | The email address associated with your Atlassian account |
| `api_token` | Yes | API token from step 1 |
| `default_project` | No | Default project key for listing and creating tickets |

## Usage

### List tickets

Search for issues using a [JQL](https://support.atlassian.com/jira-service-management-cloud/docs/use-advanced-search-with-jira-query-language-jql/) query. Returns a JSON array to stdout.

```bash
# Search with JQL
go-jira-tool -l -q "project = PROJ AND status = 'In Progress'"

# List all issues in the default project (uses default_project from config)
go-jira-tool -l
```

Output:

```json
[
  {
    "key": "PROJ-123",
    "summary": "Fix login page crash",
    "status": "In Progress",
    "labels": ["bug"]
  }
]
```

### Get ticket details

Fetches a ticket and writes it to a local folder as a markdown file with all attachments downloaded.

```bash
# Uses default output path: ~/.go-jira-tool/issues/PROJ-123/
go-jira-tool -t PROJ-123

# Override output path
go-jira-tool -t PROJ-123 -p /tmp/tickets
```

This creates a folder containing:

```
PROJ-123/
├── ticket.md           # Summary, metadata, and description as markdown
├── screenshot.png      # Downloaded attachments
├── design-doc.pdf
└── ...
```

The tool outputs the path to the created folder. Image and attachment references in the description are replaced with relative file paths to the downloaded files.

### Create a ticket

Create a new issue in the default project.

```bash
# With an inline description
go-jira-tool -c "Fix the login bug" -d "The login page crashes on special characters"

# With a description from a file
go-jira-tool -c "Fix the login bug" -dp ./description.md
```

Output: the new issue key and a link to it in Jira.

```
PROJ-456 https://your-org.atlassian.net/browse/PROJ-456
```

### Transition a ticket

Move an issue to a new workflow status. The status name is matched case-insensitively against the issue's available transitions.

```bash
go-jira-tool -transition PROJ-123 "In Progress"
go-jira-tool -transition PROJ-123 "Done"
```

Output:

```
PROJ-123 transitioned to In Progress
```

If the target status is not available, the error message lists valid transitions.

### Comment on a ticket

Add a plain-text comment to an issue. The body can be provided inline as a positional argument or read from a file with `-cf`.

```bash
# Inline comment
go-jira-tool -comment PROJ-123 "This is a comment"

# Comment from a file
go-jira-tool -comment PROJ-123 -cf /path/to/comment.txt
```

Output:

```
Comment added to PROJ-123
```

### Help

```bash
go-jira-tool -h
```

### Version

```bash
go-jira-tool -version
```

Prints the build version. Set at build time via `-ldflags "-X main.Version=<version>"` (the `build.sh` / `build.bat` scripts handle this when given a version argument). Defaults to `dev` for local builds.

## Flags

| Flag | Description |
|------|-------------|
| `-l` | List mode — search for issues |
| `-q` | JQL query string (used with `-l`) |
| `-t` | Get mode — fetch a ticket by issue key |
| `-p` | Base path for the ticket output folder (default: `~/.go-jira-tool/issues/`) |
| `-c` | Create mode — create a ticket with the given summary |
| `-d` | Ticket description as a string (used with `-c`) |
| `-dp` | Path to a file containing the ticket description (used with `-c`) |
| `-transition` | Transition mode — move an issue to a new status |
| `-comment` | Comment mode — add a comment to an issue |
| `-cf` | Path to a file containing the comment body (used with `-comment`) |
| `-version` | Print the build version and exit |

## File structure

```
~/.go-jira-tool/
├── config.json          # Your credentials and settings
└── issues/              # Default location for fetched tickets
    ├── PROJ-123/
    │   ├── ticket.md
    │   └── screenshot.png
    └── PROJ-456/
        └── ticket.md
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Security issues: see [SECURITY.md](SECURITY.md).

## License

Apache License 2.0. See [LICENSE](LICENSE) and [NOTICE](NOTICE).
