//go:build integration

package client

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

func loadEnv(t *testing.T) (username, password string) {
	t.Helper()

	f, err := os.Open("../../.env")
	if err != nil {
		t.Fatalf("Failed to open .env: %v", err)
	}
	defer f.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.Trim(value, `"'`)
		vars[key] = value
	}

	username = vars["HEMKOP_USERNAME"]
	password = vars["HEMKOP_PASSWORD"]
	if username == "" || password == "" {
		t.Fatal("HEMKOP_USERNAME and HEMKOP_PASSWORD must be set in .env")
	}
	return username, password
}

func TestLoginIntegration(t *testing.T) {
	username, password := loadEnv(t)

	c := NewClient()

	if err := c.Login(username, password); err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	customer, err := c.GetCustomer()
	if err != nil {
		t.Fatalf("GetCustomer failed: %v", err)
	}

	if customer.UID == "anonymous" || customer.UID == "" {
		t.Fatal("Expected authenticated customer, got anonymous")
	}

	if customer.Name == "" {
		t.Fatal("Expected customer name to be non-empty")
	}

	t.Logf("Logged in as: %s (%s)", customer.Name, customer.Email)
}
