package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	installSkillAgent     string
	installSkillDir       string
	installSkillCommand   string
	installSkillOverwrite bool
	installSkillScope     string
)

var installSkillCmd = &cobra.Command{
	Use:   "install-skill",
	Short: "Install queryli SKILL.md for local AI coding agents",
	Long: `Install a queryli skill into local AI coding agent skills directories.

By default, this installs user/global skills for all supported agents:
  Codex:          ~/.agents/skills/queryli/SKILL.md
  Claude Code:    ~/.claude/skills/queryli/SKILL.md
  Cursor:         ~/.cursor/skills/queryli/SKILL.md
  Gemini CLI:     ~/.gemini/skills/queryli/SKILL.md
  Windsurf:       ~/.codeium/windsurf/skills/queryli/SKILL.md

Use --agent to install for one agent, --scope project for project-local skills,
or --dir to write into a custom skills directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targets, err := resolveSkillInstallTargets(installSkillAgent, installSkillScope, installSkillDir)
		if err != nil {
			return err
		}

		for _, target := range targets {
			skillPath, err := installSkill(target.Dir, installSkillCommand, installSkillOverwrite)
			if err != nil {
				return fmt.Errorf("%s: %w", target.Name, err)
			}
			fmt.Printf("Installed queryli skill for %s: %s\n", target.Name, skillPath)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installSkillCmd)
	installSkillCmd.Flags().StringVar(&installSkillAgent, "agent", "all", "agent to install for (all|codex|claude|cursor|gemini|windsurf)")
	installSkillCmd.Flags().StringVar(&installSkillCommand, "command", "queryli", "queryli command or absolute path for the skill to use")
	installSkillCmd.Flags().StringVar(&installSkillDir, "dir", "", "custom skills directory to install into")
	installSkillCmd.Flags().BoolVar(&installSkillOverwrite, "overwrite", false, "replace an existing queryli SKILL.md")
	installSkillCmd.Flags().StringVar(&installSkillScope, "scope", "user", "install scope (user|project)")
}

type skillInstallTarget struct {
	Name string
	Dir  string
}

func resolveSkillInstallTargets(agent, scope, dir string) ([]skillInstallTarget, error) {
	if dir != "" {
		expanded, err := expandHome(dir)
		if err != nil {
			return nil, err
		}
		return []skillInstallTarget{{Name: "custom", Dir: expanded}}, nil
	}

	agent = strings.ToLower(strings.TrimSpace(agent))
	scope = strings.ToLower(strings.TrimSpace(scope))

	if scope != "user" && scope != "project" {
		return nil, fmt.Errorf("unsupported scope %q; use user or project", scope)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("working dir: %w", err)
	}

	allTargets := map[string]skillInstallTarget{
		"claude": {
			Name: "Claude Code",
			Dir:  scopedSkillDir(scope, home, cwd, ".claude", filepath.Join(".claude", "skills")),
		},
		"codex": {
			Name: "Codex",
			Dir:  scopedSkillDir(scope, home, cwd, ".agents", filepath.Join(".agents", "skills")),
		},
		"cursor": {
			Name: "Cursor",
			Dir:  scopedSkillDir(scope, home, cwd, ".cursor", filepath.Join(".cursor", "skills")),
		},
		"gemini": {
			Name: "Gemini CLI",
			Dir:  scopedSkillDir(scope, home, cwd, ".gemini", filepath.Join(".gemini", "skills")),
		},
		"windsurf": {
			Name: "Windsurf",
			Dir:  windsurfSkillDir(scope, home, cwd),
		},
	}

	if agent == "all" {
		return []skillInstallTarget{
			allTargets["codex"],
			allTargets["claude"],
			allTargets["cursor"],
			allTargets["gemini"],
			allTargets["windsurf"],
		}, nil
	}

	target, ok := allTargets[agent]
	if !ok {
		return nil, fmt.Errorf("unsupported agent %q; use all, codex, claude, cursor, gemini, or windsurf", agent)
	}
	return []skillInstallTarget{target}, nil
}

func scopedSkillDir(scope, home, cwd, userConfigDir, projectSkillsDir string) string {
	if scope == "project" {
		return filepath.Join(cwd, projectSkillsDir)
	}
	return filepath.Join(home, userConfigDir, "skills")
}

func windsurfSkillDir(scope, home, cwd string) string {
	if scope == "project" {
		return filepath.Join(cwd, ".windsurf", "skills")
	}
	return filepath.Join(home, ".codeium", "windsurf", "skills")
}

func installSkill(targetDir, queryliCommand string, overwrite bool) (string, error) {
	skillDir := filepath.Join(targetDir, "queryli")
	skillPath := filepath.Join(skillDir, "SKILL.md")

	if _, err := os.Stat(skillPath); err == nil && !overwrite {
		return "", fmt.Errorf("%s already exists; use --overwrite to replace it", skillPath)
	} else if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("check skill file: %w", err)
	}

	if err := os.MkdirAll(skillDir, 0700); err != nil {
		return "", fmt.Errorf("create skill directory: %w", err)
	}

	if err := os.WriteFile(skillPath, []byte(buildQueryliSkillMarkdown(queryliCommand)), 0600); err != nil {
		return "", fmt.Errorf("write SKILL.md: %w", err)
	}

	return skillPath, nil
}

func expandHome(path string) (string, error) {
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		return home, nil
	}

	if len(path) >= 2 && path[0] == '~' && os.IsPathSeparator(path[1]) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}

func buildQueryliSkillMarkdown(queryliCommand string) string {
	if queryliCommand == "" {
		queryliCommand = "queryli"
	}

	return fmt.Sprintf(`---
name: queryli
description: Use queryli to manage local database connection profiles, connect or disconnect the queryli daemon, check database connection status, and run SQL queries or SQL files through the active queryli connection. Use when Codex needs to inspect, query, seed, or validate PostgreSQL, MySQL, SQLite, or Oracle databases with the queryli CLI.
---

# Queryli

Use the local %[1]s CLI for database work when a repository or user workflow expects queryli-managed profiles and daemon connections.

Run queryli with:

`+"```bash"+`
%[2]s
`+"```"+`

## Workflow

1. Check available profiles with %[3]s.
2. Select a profile with %[4]s when needed.
3. Start the connection daemon with %[5]s.
4. Verify the connection with %[6]s or %[7]s.
5. Run SQL with %[8]s or execute a file with %[9]s.
6. Disconnect with %[10]s when the task should not leave a daemon running.

## Output

- Prefer `+"`--format json`"+` for data that will be parsed or compared.
- Use table output for quick human inspection.
- Use CSV output only when the result is intended for a CSV pipeline or file.

## Safety

- Inspect destructive SQL before running it.
- Ask before running schema migrations, deletes, truncates, broad updates, or production-looking profile names.
- Do not store database passwords in profiles. Use `+"`QUERYLI_PASSWORD`"+` or the `+"`--password`"+` flag when credentials are required.

## Useful Commands

`+"```bash"+`
%[2]s profile list
%[2]s profile use <name>
%[2]s connect [profile]
%[2]s status
%[2]s ping
%[2]s query "SELECT 1" --format json
%[2]s exec ./script.sql
%[2]s disconnect
`+"```"+`
`,
		inlineCode("queryli"),
		queryliCommand,
		inlineCode(queryliCommand+" profile list"),
		inlineCode(queryliCommand+" profile use <name>"),
		inlineCode(queryliCommand+" connect [profile]"),
		inlineCode(queryliCommand+" ping"),
		inlineCode(queryliCommand+" status"),
		inlineCode(queryliCommand+` query "SELECT ..."`),
		inlineCode(queryliCommand+" exec path/to/file.sql"),
		inlineCode(queryliCommand+" disconnect"),
	)
}

func inlineCode(value string) string {
	return "`" + value + "`"
}
