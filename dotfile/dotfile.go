// Package dotfile holds functionality for working with dotfiles and registries.
package dotfile

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrNotFound is returned when a dotfile is not found.
var ErrNotFound = errors.New("dotfile not found")

// Dotfile represents a dotfile managed by a registry.
type Dotfile struct {
	// Name is the name of the dotfile used to uniquely identify it in the registry.
	Name string `yaml:"-"`
	// SrcPath is the path to the dotfile source within the registry.
	// It must be relative and cannot start with '.' or '..'.
	SrcPath string `yaml:"src"`
	// DstPath is the path dotfile on the OS filesystem.
	// It must be absolute i.e. start with a slash.
	// The one exception to this rule is it maybe start with '~/'.
	// It is up to the caller to decide how to handle '~'.
	DstPath string `yaml:"dst"`
	// OS is a list of supported operating systems for this dotfile.
	// If OS is empty, it is interpreted as all operating systems being supported.
	OS []string `yaml:"os"`
}

// config represents a `dot.yml` file.
type config struct {
	Dotfiles map[string]Dotfile `yaml:"dotfiles"`
}

// Registry represents a dot registry.
// A registry is a directory containing dotfile sources and configuration.
// Registries are read-only.
type Registry struct {
	fs  fs.FS
	cfg config
}

// NewRegistry creates a new Registry object from fsys. fsys must contain
// a `dot.yml` file that holds the configuration for the registry.
// NewRegistry will read `dot.yml` and return an validation errors encountered.
func NewRegistry(fsys fs.FS) (*Registry, error) {
	const filename = "dot.yml"
	f, err := fsys.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s from registry: %w", filename, err)
	}
	defer f.Close()

	var cfg config
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", filename, err)
	}

	// Validate and normalize dotfiles
	var errs ErrorList
	for n, df := range cfg.Dotfiles {
		df.Name = n
		var msgs []string
		// Validate SrcPath
		if fs.ValidPath(df.SrcPath) {
			_, err := fs.Stat(fsys, df.SrcPath)
			if errors.Is(err, fs.ErrNotExist) {
				msgs = append(msgs, fmt.Sprintf("%q does not exist", df.SrcPath))
			} else if err != nil {
				msgs = append(msgs, fmt.Sprintf("failed to stat %q: %s", df.SrcPath, err))
			}
		} else {
			msgs = append(msgs, "src path is invalid")
		}

		// Validate DstPath. DstPath must be an absolute path (i.e. begin with `/`),
		// with the one exception being it may start with `~`.
		if !strings.HasPrefix(df.DstPath, "~") && !filepath.IsAbs(df.DstPath) {
			msgs = append(msgs, "dst must be an absolute path")
		}

		if len(msgs) > 0 {
			errs = append(errs, &ValidationError{
				DotfileName: n,
				Messages:    msgs,
			})
		}
		cfg.Dotfiles[n] = df
	}
	if len(errs) > 0 {
		return nil, errs
	}
	return &Registry{fsys, cfg}, nil
}

// Dotfiles returns a list of dotfiles contained in the registry.
// A list of names can be provided to filter which dotfiles are returned.
// If no names are provided, all dotfiles are returned.
func (r *Registry) Dotfiles(names ...string) ([]Dotfile, error) {
	var dotfiles []Dotfile
	if len(names) == 0 {
		for _, df := range r.cfg.Dotfiles {
			dotfiles = append(dotfiles, df)
		}
		// Sort the dotfiles so the returned order is deterministic
		sort.Slice(dotfiles, func(i, j int) bool {
			return dotfiles[i].Name < dotfiles[j].Name
		})
		return dotfiles, nil
	}

	var notFound []string
	for _, name := range names {
		df, ok := r.cfg.Dotfiles[name]
		if !ok {
			notFound = append(notFound, name)
			continue
		}
		dotfiles = append(dotfiles, df)
	}
	if len(notFound) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, strings.Join(notFound, ", "))
	}
	return dotfiles, nil
}

// OpenDotfile opens the dotfile and returns a fs.File allowing access to the data.
func (r *Registry) OpenDotfile(name string) (fs.File, error) {
	df, ok := r.cfg.Dotfiles[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	f, err := r.fs.Open(df.SrcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", df.SrcPath, err)
	}
	return f, nil
}

// ValidationError represents a dotfile having failed validation.
// It contains the dotfile name and a list of validation failure messages.
type ValidationError struct {
	DotfileName string
	Messages    []string
}

func (ve *ValidationError) Error() string {
	var sb strings.Builder
	sb.WriteString(ve.DotfileName)
	sb.WriteString(": ")
	for i, msg := range ve.Messages {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(msg)
	}
	return sb.String()
}

// ErrorList is a list of errors encountered.
type ErrorList []error

func (e ErrorList) Error() string {
	strs := make([]string, len(e))
	for i, err := range e {
		strs[i] = err.Error()
	}
	return strings.Join(strs, "\n")
}
