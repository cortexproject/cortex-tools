package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTemplates(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "template")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	// Write some data to the file
	text := "This is a test template"
	if _, err := tmpfile.Write([]byte(text)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Call createTemplates
	templates, err := createTemplates([]string{tmpfile.Name()})
	if err != nil {
		t.Fatal(err)
	}

	// Check the returned map
	if len(templates) != 1 {
		t.Fatalf("Expected 1 template, got %d", len(templates))
	}
	if templates[filepath.Base(tmpfile.Name())] != text {
		t.Fatalf("Expected template content to be '%s', got '%s'", text, templates[filepath.Base(tmpfile.Name())])
	}
}

func TestCreateTemplates_DuplicateFilenames(t *testing.T) {
	// Create two temporary files with the same base name in different directories
	dir1 := filepath.Join(os.TempDir(), "dir1")
	if err := os.Mkdir(dir1, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir1) // clean up
	tmpfile1, err := os.Create(filepath.Join(os.TempDir(), "dir1", "fool.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile1.Name()) // clean up

	dir2 := filepath.Join(os.TempDir(), "dir2")
	if err := os.Mkdir(dir2, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir2) // clean up
	tmpfile2, err := os.Create(filepath.Join(os.TempDir(), "dir2", "fool.tmpl"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile2.Name()) // clean up

	// Call createTemplates
	_, err = createTemplates([]string{tmpfile1.Name(), tmpfile2.Name()})
	if err == nil {
		t.Fatal("Expected error due to duplicate filenames, got nil")
	}

	// Check that the error message contains "duplicate template file name"
	if !strings.Contains(err.Error(), "duplicate template file name") {
		t.Fatalf("Expected error message to contain 'duplicate template file name', got '%s'", err.Error())
	}
}
