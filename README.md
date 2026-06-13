# queryli

`queryli` is a multi-database CLI for managing connection profiles and running SQL through a persistent local daemon.

It supports PostgreSQL, MySQL, SQLite, and Oracle.

## Features

- Manage multiple database connection profiles.
- Keep one active database connection alive through a background daemon.
- Run ad hoc SQL queries or execute `.sql` files.
- Print query results as table, JSON, or CSV.
- Install a `SKILL.md` helper for local AI coding agents.

## Install

From the repository root:

```bash
go install .
```

Make sure your Go binary directory is on `PATH`. On Windows this is commonly:

```text
%USERPROFILE%\go\bin
```

Verify:

```bash
queryli --help
```

## Quick Start

Create a PostgreSQL profile:

```bash
queryli profile add --name local-pg --type postgres \
  --host localhost --port 5432 --user postgres --db postgres --sslmode disable
```

Create a SQLite profile:

```bash
queryli profile add --name local-sqlite --type sqlite --path ./dev.db
```

Set the active profile:

```bash
queryli profile use local-pg
```

Connect:

```bash
QUERYLI_PASSWORD=your-password queryli connect
```

Run a query:

```bash
queryli query "SELECT 1" --format json
```

Execute a SQL file:

```bash
queryli exec ./script.sql
```

Check status:

```bash
queryli status
queryli ping
```

Disconnect:

```bash
queryli disconnect
```

## Commands

```text
queryli profile add
queryli profile list
queryli profile remove
queryli profile use
queryli connect [profile]
queryli disconnect
queryli status
queryli ping
queryli query <sql>
queryli exec <file.sql>
queryli install-skill
```

## Configuration

Profiles are stored in:

```text
~/.queryli/config.yaml
```

Runtime daemon files are stored in the same directory:

```text
~/.queryli/daemon.pid
~/.queryli/daemon.sock
~/.queryli/daemon.meta
~/.queryli/daemon.log
```

Passwords are not stored by `profile add`. Pass credentials at connect time:

```bash
QUERYLI_PASSWORD=your-password queryli connect local-pg
```

or:

```bash
queryli connect local-pg --password your-password
```

## Output Formats

Use `--format` or `-f`:

```bash
queryli query "SELECT * FROM users LIMIT 5" --format table
queryli query "SELECT * FROM users LIMIT 5" --format json
queryli query "SELECT * FROM users LIMIT 5" --format csv
```

## AI Agent Skill

Install a `queryli` skill for supported local coding agents:

```bash
queryli install-skill
```

By default this installs user-level skills for:

- Codex: `~/.agents/skills/queryli/SKILL.md`
- Claude Code: `~/.claude/skills/queryli/SKILL.md`
- Cursor: `~/.cursor/skills/queryli/SKILL.md`
- Gemini CLI: `~/.gemini/skills/queryli/SKILL.md`
- Windsurf: `~/.codeium/windsurf/skills/queryli/SKILL.md`

Install for one agent:

```bash
queryli install-skill --agent claude
queryli install-skill --agent cursor
```

Install into the current project:

```bash
queryli install-skill --agent all --scope project
```

Use a custom skills directory:

```bash
queryli install-skill --dir ~/.some-agent/skills
```

## Development

Run checks:

```bash
go test ./...
```

Build locally:

```bash
go build .
```
