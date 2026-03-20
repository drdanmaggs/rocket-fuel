package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// NOTE: Cobra commands are not thread-safe. Tests that call SetArgs/SetOut/Execute
// on the package-level rootCmd must NOT use t.Parallel(). Tests that only read
// rootCmd fields can be parallel.

func TestRootCommandExists(t *testing.T) {
	t.Parallel()

	if rootCmd == nil {
		t.Fatal("expected root command to be initialized")
	}

	if rootCmd.Use != "rf" {
		t.Errorf("expected root command use to be 'rf', got %q", rootCmd.Use)
	}
}

func TestVersionCommandRegistered(t *testing.T) {
	// Not parallel — reads rootCmd.Commands() which sorts internally.

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

func TestVersionCommandOutput(t *testing.T) {
	// Not parallel — mutates rootCmd via SetOut/SetArgs.

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "rf") {
		t.Errorf("expected output to start with 'rocket-fuel', got %q", out)
	}

	// Reset for other tests.
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}

func TestHelpContainsDescription(t *testing.T) {
	// Not parallel — mutates rootCmd via SetOut/SetArgs.

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Visionary/Integrator") {
		t.Errorf("expected help to mention 'Visionary/Integrator', got %q", out)
	}

	// Reset for other tests.
	rootCmd.SetOut(nil)
	rootCmd.SetArgs(nil)
}
