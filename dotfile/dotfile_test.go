package dotfile_test

import (
	"errors"
	"io"
	"io/fs"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/cszatmary/dot/dotfile"
)

func TestNewRegistryValidationError(t *testing.T) {
	mfs := fstest.MapFS{
		"dot.yml": {
			Data: []byte(`dotfiles:
  git:
    src: ../git/gitconfig
    dst: ~/.gitconfig
  zsh:
    src: zsh/zshrc
    dst: home/.zshrc
`),
		},
	}
	_, err := dotfile.NewRegistry(mfs)
	var errs dotfile.ErrorList
	if !errors.As(err, &errs) {
		t.Fatalf("got error %v with type %T, wanted a dotfiles.ErrorList", err, err)
	}

	var gotNames []string
	for _, err := range errs {
		var validationErr *dotfile.ValidationError
		if !errors.As(err, &validationErr) {
			t.Errorf("got error %v with type %T, wanted a *dotfiles.ValidationError", err, err)
		}
		gotNames = append(gotNames, validationErr.DotfileName)
	}
	sort.Strings(gotNames)
	want := []string{"git", "zsh"}
	if !reflect.DeepEqual(gotNames, want) {
		t.Errorf("got dotfile names %v, want %v", gotNames, want)
	}
}

func TestRegistryDotfiles(t *testing.T) {
	registry, err := dotfile.NewRegistry(createRegistryFixture())
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	got, err := registry.Dotfiles()
	if err != nil {
		t.Errorf("want nil error, got %v", err)
	}
	want := []dotfile.Dotfile{
		{Name: "git", SrcPath: "git/gitconfig", DstPath: "~/.gitconfig"},
		{Name: "vim", SrcPath: "vim/vimrc", DstPath: "~/.vimrc"},
		{Name: "zsh", SrcPath: "zsh/zshrc", DstPath: "~/.zshrc"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got dotfiles %v, want %v", got, want)
	}
}

func TestRegistryDotfilesFiltered(t *testing.T) {
	registry, err := dotfile.NewRegistry(createRegistryFixture())
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	got, err := registry.Dotfiles("git", "zsh")
	if err != nil {
		t.Errorf("want nil error, got %v", err)
	}
	want := []dotfile.Dotfile{
		{Name: "git", SrcPath: "git/gitconfig", DstPath: "~/.gitconfig"},
		{Name: "zsh", SrcPath: "zsh/zshrc", DstPath: "~/.zshrc"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got dotfiles %v, want %v", got, want)
	}
}

func TestRegistryDotfilesNotFound(t *testing.T) {
	registry, err := dotfile.NewRegistry(createRegistryFixture())
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	_, err = registry.Dotfiles("git", "foo")
	if !errors.Is(err, dotfile.ErrNotFound) {
		t.Errorf("got %v, want dotfiles.ErrNotFound", err)
	}
}

func TestRegistryOpenDotfile(t *testing.T) {
	registry, err := dotfile.NewRegistry(createRegistryFixture())
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	f, err := registry.OpenDotfile("zsh")
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	t.Cleanup(func() {
		f.Close()
	})
	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	got := string(data)
	want := `export PATH="$(go env GOPATH)/bin:$PATH"`
	if got != want {
		t.Errorf("got dotfile contents %q, want %v", got, want)
	}
}

func TestRegistryOpenDotfileNotFound(t *testing.T) {
	registry, err := dotfile.NewRegistry(createRegistryFixture())
	if err != nil {
		t.Fatalf("want nil error, got %v", err)
	}
	_, err = registry.OpenDotfile("foo")
	if !errors.Is(err, dotfile.ErrNotFound) {
		t.Errorf("got %v, want dotfiles.ErrNotFound", err)
	}
}

func createRegistryFixture() fs.FS {
	return fstest.MapFS{
		"dot.yml": {
			Data: []byte(`dotfiles:
  git:
    src: git/gitconfig
    dst: ~/.gitconfig
  vim:
    src: vim/vimrc
    dst: ~/.vimrc
  zsh:
    src: zsh/zshrc
    dst: ~/.zshrc
`),
		},
		"git/gitconfig": {
			Data: []byte(`[pull]
    ff = only
    rebase = true
`),
		},
		"vim/vimrc": {
			Data: []byte(`set nocompatible              " be iMproved, required
filetype off                  " required
syntax on
`),
		},
		"zsh/zshrc": {
			Data: []byte(`export PATH="$(go env GOPATH)/bin:$PATH"`),
		},
	}
}
