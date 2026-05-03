package initdemo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerate_CreatesExpectedFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "krr-demo")

	if err := Generate(dir, false); err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	expected := []string{
		"README.md",
		"usage.csv",
		"usage.json",
		filepath.Join("reports", "recommendations.md"),
		filepath.Join("reports", "recommendations.json"),
		filepath.Join("reports", "recommendations.csv"),
	}

	for _, rel := range expected {
		path := filepath.Join(dir, rel)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file not found: %s", path)
		}
	}
}

func TestGenerate_FailsIfExistsWithoutForce(t *testing.T) {
	dir := t.TempDir() // already exists

	err := Generate(dir, false)
	if err == nil {
		t.Fatal("expected error when directory exists and --force not set")
	}
}

func TestGenerate_SucceedsWithForce(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "krr-demo")
	// First create
	if err := Generate(dir, false); err != nil {
		t.Fatalf("first Generate error: %v", err)
	}
	// Second create with force
	if err := Generate(dir, true); err != nil {
		t.Fatalf("second Generate (force) error: %v", err)
	}
}
