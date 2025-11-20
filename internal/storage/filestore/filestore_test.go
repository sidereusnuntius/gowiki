package filestore

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/wiki/internal/storage"
)

var store storage.Storage
var path string

func TestMain(m *testing.M) {
	var err error
	path, err = os.MkdirTemp(".", "tempdir")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup tests")
		return
	}

	store = &FileStore{
		Root: path,
	}

	m.Run()
	if err =os.RemoveAll(path); err != nil {
		log.Fatal().Err(err).Msg("removal of temporary directory failed")
	}
}

func TestCreate(t *testing.T) {
	cases := []struct{
		Casename string
		Path string
		Content string
		Err error
	}{
		{"create file", "f1.txt", "hello, world!", nil},
		{"create duplicate file", "f1.txt", "hello, world!", storage.ErrAlreadyExists},
	}

	for _, c := range cases {
		t.Run(c.Casename, func(t *testing.T) {
			err := store.Create(strings.NewReader(c.Content), c.Path)
			if err != nil {
				if c.Err == nil {
					t.Error("unexpected error:", err)
					return
				} else if !errors.Is(err, c.Err) {
					t.Errorf("unexpected error type.\nexpected: %s\ngot: %s\n", c.Err, err)
				} else {
					return
				}
			}

			f, err := os.Open(filepath.Join(path, c.Path))
			if err != nil {
				t.Errorf("failed to open file: %s", err)
				return
			}
			defer f.Close()

			if fullPath := filepath.Join(path, c.Path); fullPath != f.Name() {
				t.Errorf("expected filename %s, got %s", c.Path, f.Name())
			}
			content, err := io.ReadAll(f)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if string(content) != c.Content {
				t.Errorf("expected \"%s\", got \"%s\"", c.Content, content)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	name := "moribundus"
	newpath := filepath.Join(path, name)
	f, err := os.Create(newpath)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	f.Close()

	err = store.Delete(name)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	name = "none"
	err = store.Delete(name)
	if err == nil || !errors.Is(err, storage.ErrNotExist) {
		t.Errorf("unexpected err: %s\nexpected \"%s\"", err, storage.ErrNotExist)
	}
}