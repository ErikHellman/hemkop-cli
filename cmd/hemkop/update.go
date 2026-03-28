package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const (
	repo      = "ErikHellman/hemkop-cli"
	releaseURL = "https://api.github.com/repos/" + repo + "/releases/latest"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func runUpdate() {
	fmt.Fprintf(os.Stderr, "Current version: %s\n", version)
	fmt.Fprintf(os.Stderr, "Checking for updates...\n")

	// Fetch latest release info
	resp, err := http.Get(releaseURL)
	if err != nil {
		fatal("checking for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fatal("GitHub API returned %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fatal("parsing release info: %v", err)
	}

	latest := release.TagName
	if latest == version {
		fmt.Fprintf(os.Stderr, "Already up to date (%s)\n", version)
		return
	}

	fmt.Fprintf(os.Stderr, "Updating %s -> %s\n", version, latest)

	// Download the new binary
	assetName := fmt.Sprintf("hemkop-%s-%s", runtime.GOOS, runtime.GOARCH)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, latest, assetName)

	resp, err = http.Get(downloadURL)
	if err != nil {
		fatal("downloading update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fatal("downloading update: %s (no binary for %s/%s?)", resp.Status, runtime.GOOS, runtime.GOARCH)
	}

	// Get path of current executable
	execPath, err := os.Executable()
	if err != nil {
		fatal("finding executable path: %v", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		fatal("resolving executable path: %v", err)
	}

	// Write to a temp file in the same directory (ensures same filesystem for rename)
	dir := filepath.Dir(execPath)
	tmpFile, err := os.CreateTemp(dir, "hemkop-update-*")
	if err != nil {
		fatal("creating temp file: %v", err)
	}
	tmpPath := tmpFile.Name()

	_, err = io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		fatal("writing update: %v", err)
	}

	// Copy permissions from existing binary
	info, err := os.Stat(execPath)
	if err != nil {
		os.Remove(tmpPath)
		fatal("reading file permissions: %v", err)
	}
	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		os.Remove(tmpPath)
		fatal("setting permissions: %v", err)
	}

	// Atomic replace
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		fatal("replacing binary: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Updated to %s\n", latest)
}
