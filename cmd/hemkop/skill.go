package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed skill.md
var skillContent []byte

func runInstallSkill() {
	home, err := os.UserHomeDir()
	if err != nil {
		fatal("finding home directory: %v", err)
	}

	skillDir := filepath.Join(home, ".claude", "skills", "hemkop")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		fatal("creating skill directory: %v", err)
	}

	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, skillContent, 0644); err != nil {
		fatal("writing skill file: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Installed hemkop skill to %s\n", skillPath)
}
