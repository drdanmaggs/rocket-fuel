package cmd

import (
	"testing"
)

func TestRootCommandExists(t *testing.T) {
	t.Parallel()

	if rootCmd == nil {
		t.Fatal("expected root command to be initialized")
	}

	if rootCmd.Use != "rocket-fuel" {
		t.Errorf("expected root command use to be 'rocket-fuel', got %q", rootCmd.Use)
	}
}

func TestVersionCommandRegistered(t *testing.T) {
	t.Parallel()

	var found bool
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected 'version' subcommand to be registered")
	}
}
