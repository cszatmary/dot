package client

import (
	"crypto/md5"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/TouchBistro/goutils/file"
	"github.com/cszatmary/dot/dotfile"
	"github.com/pkg/errors"
)

// ErrNotSetup is returned when a dotfile has not been setup to be managed by dot.
var ErrNotSetup = stderrors.New("not setup")

// ErrSetup is returned when dot has already been setup with a different registry.
var ErrSetup = stderrors.New("already setup with a different registry")

type lockfile struct {
	RegistryDir string                 `json:"registryDir"`
	Dotfiles    map[string]dotfileInfo `json:"dotfiles"`
}

type dotfileInfo struct {
	// stringified md5 hash, used to determine if the file has been modified
	DstHash string `json:"dstHash"`
}

// Debugger wraps the Debugf method and represents any type that
// can write debug messages.
type Debugger interface {
	Debugf(format string, args ...interface{})
}

// noopDebugger is a Debugger with a no-op Debugf method.
type noopDebugger struct{}

func (noopDebugger) Debugf(format string, args ...interface{}) {}

// Client provides the API for managing dotfiles with dot.
type Client struct {
	lf       *lockfile
	registry *dotfile.Registry
	// configurable
	homeDir  string
	debugger Debugger
}

// New creates a new Client instance.
func New(opts ...Option) (*Client, error) {
	c := &Client{
		lf: &lockfile{},
	}
	for _, opt := range opts {
		opt(c)
	}
	// Create a noopDebugger if none was set to prevent panics on calls to Debugf
	if c.debugger == nil {
		c.debugger = noopDebugger{}
	}
	if c.homeDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Wrap(err, "failed to find user home directory")
		}
		c.homeDir = homeDir
	}

	lfp := c.lockfilePath()
	f, err := os.Open(lfp)
	if errors.Is(err, os.ErrNotExist) {
		// No lockfile, dot has not been setup
		return c, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open lockfile %s", lfp)
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(c.lf)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse lockfile %s", lfp)
	}
	if !c.IsSetup() {
		return c, nil
	}

	// dot is setup, load registry
	c.registry, err = dotfile.NewRegistry(os.DirFS(c.lf.RegistryDir))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load dot registry at %s", c.lf.RegistryDir)
	}
	return c, nil
}

// Option is a function that takes a Client instance and applies a configuration to it.
type Option func(*Client)

// WithHomeDir sets the directory that the client should use when the home directory is needed.
func WithHomeDir(dir string) Option {
	return func(c *Client) {
		c.homeDir = dir
	}
}

// WithDebugger sets a Debugger that should be used by the client to write debug messages.
func WithDebugger(d Debugger) Option {
	return func(c *Client) {
		c.debugger = d
	}
}

// IsSetup returns whether or not dot has been setup to manage dotfiles.
func (c *Client) IsSetup() bool {
	return c.lf.RegistryDir != ""
}

// configPath returns the root dir where dot stores config.
func (c *Client) configPath() string {
	return filepath.Join(c.homeDir, ".config", "dot")
}

func (c *Client) lockfilePath() string {
	return filepath.Join(c.configPath(), "dot.lock")
}

func (c *Client) dotfileBackupPath(df dotfile.Dotfile) string {
	return filepath.Join(c.configPath(), "backups", df.SrcPath) + ".bak"
}

func (c *Client) writeLockfile() error {
	lfp := c.lockfilePath()
	f, err := os.OpenFile(lfp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return errors.Wrapf(err, "failed to create/open file %s", lfp)
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(c.lf)
	if err != nil {
		return errors.Wrapf(err, "failed to write lockfile to %s", lfp)
	}
	return nil
}

// Setup will setup dot to manage dotfiles. If the dotfile destination already exists,
// a backup of it will be made, so the original version can be restored.
// Setup will only setup dotfiles that have not been previously setup. This means
// it can be called multiple times to setup additional dotfiles.
//
// If registryDir is different than the one used by dot, Setup will return ErrSetup
// unless force is true, in which case it will overwrite the current registry dir.
func (c *Client) Setup(registryDir string, force bool) error {
	// Check if already setup
	if c.lf.RegistryDir != "" && c.lf.RegistryDir != registryDir && !force {
		return errors.Wrap(ErrSetup, registryDir)
	}
	var err error
	c.registry, err = dotfile.NewRegistry(os.DirFS(registryDir))
	if err != nil {
		return errors.Wrapf(err, "failed to load dot registry at %s", registryDir)
	}

	// Get hash of each dst dotfile
	// This will be used to determine if the dotfiles are out of date
	c.debugger.Debugf("Backing up existing dotfiles and saving hashes")
	if c.lf.Dotfiles == nil {
		c.lf.Dotfiles = make(map[string]dotfileInfo)
	}

	dfs, err := c.registry.Dotfiles()
	if err != nil {
		return errors.Wrap(err, "failed to get dotfiles from registry")
	}

	for _, df := range dfs {
		if !supportsOS(df) {
			continue
		}
		// Check if already setup, and ignore if so
		if _, ok := c.lf.Dotfiles[df.Name]; ok {
			c.debugger.Debugf("Dotfile %s already setup, skipping", df.Name)
			continue
		}

		df.DstPath = expandTilde(df.DstPath, c.homeDir)
		f, err := os.Open(df.DstPath)
		if errors.Is(err, os.ErrNotExist) {
			// It's fine if dst doesn't exist, it will be created by Apply
			c.lf.Dotfiles[df.Name] = dotfileInfo{}
			continue
		}
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", df.DstPath)
		}

		c.debugger.Debugf("Saving hash of %s", df.DstPath)
		hash, err := md5Hash(f)
		if err != nil {
			return errors.Wrapf(err, "failed to get hash of %s", df.DstPath)
		}

		// Backup dotfile, do this before saving the hash and marking this as "setup"
		c.debugger.Debugf("Creating backup of %s", df.DstPath)
		backupPath := c.dotfileBackupPath(df)
		if err := file.CopyFile(df.DstPath, backupPath); err != nil {
			return errors.Wrapf(err, "failed to backup %s to %s", df.DstPath, backupPath)
		}

		c.lf.Dotfiles[df.Name] = dotfileInfo{DstHash: string(hash)}
	}
	c.debugger.Debugf("Finished backing up dotfiles and saving hashes")

	// Mark as setup
	c.lf.RegistryDir = registryDir
	if err := c.writeLockfile(); err != nil {
		return errors.Wrap(err, "failed to save lockfile")
	}
	return nil
}

// Apply will copy dotfile sources from a registry to their destination.
// Optionally, a list of dotfile names can be provided to only apply specific dotfiles.
// If no names are provided, all dotfiles will be applied.
//
// By default, Apply will check if the dotfile destination file has been manually modified.
// If a modification is detected, the dotfile will not be applied and an error will be
// returned. If force is set to true, this check is skipped and the dotfile is always applied.
func (c *Client) Apply(force bool, names ...string) error {
	retrieved, err := c.registry.Dotfiles(names...)
	if err != nil {
		return errors.Wrap(err, "failed to get dotfiles from registry")
	}
	// Filter out dotfiles not supported by the current OS
	var dfs []dotfile.Dotfile
	for _, df := range retrieved {
		if supportsOS(df) {
			dfs = append(dfs, df)
		}
	}

	// Make sure it is safe to apply updates
	// If there are any dotfiles whose hash is not equal to the hash
	// in the lockfile then it has been manually modified
	c.debugger.Debugf("Checking if dotfiles have been modified")
	for _, df := range dfs {
		dfInfo, ok := c.lf.Dotfiles[df.Name]
		// Make sure dotfile was setup
		if !ok {
			return errors.Wrap(ErrNotSetup, df.Name)
		}

		df.DstPath = expandTilde(df.DstPath, c.homeDir)
		f, err := os.Open(df.DstPath)
		if errors.Is(err, os.ErrNotExist) {
			// Dst doesn't exist, will be created below
			continue
		}
		if err != nil {
			return errors.Wrapf(err, "failed to open file %s", df.DstPath)
		}

		hash, err := md5Hash(f)
		if err != nil {
			return errors.Wrapf(err, "failed to get hash of %s", df.DstPath)
		}
		if string(hash) == dfInfo.DstHash {
			c.debugger.Debugf("No modifications detected to %s", df.DstPath)
			continue
		}
		if !force {
			return errors.Errorf("%s was manually modified", df.DstPath)
		}
		c.debugger.Debugf("%s was manually modified, but force mode is enabled", df.DstPath)
	}

	// Check if lockfiles are out of date
	type outdatedLockfile struct {
		df      dotfile.Dotfile
		newHash string
	}
	var outdated []outdatedLockfile
	c.debugger.Debugf("Checking if dotfiles are outdated")
	for _, df := range dfs {
		f, err := c.registry.OpenDotfile(df.Name)
		if err != nil {
			return errors.Wrapf(err, "failed to open dotfile %s", df.Name)
		}
		hash, err := md5Hash(f)
		if err != nil {
			return errors.Wrapf(err, "failed to get hash of %s", df.DstPath)
		}
		dfInfo := c.lf.Dotfiles[df.Name]
		strHash := string(hash)
		if force || strHash != dfInfo.DstHash {
			c.debugger.Debugf("%s is out of date, updating", df.Name)
			outdated = append(outdated, outdatedLockfile{df, strHash})
		}
	}

	// Apply src to dest
	// TODO would be nice if this behaved like an automic transaction
	// i.e. if one failed any successful ones would be rolled back
	// and the user could retry rather than leaving in a partially successful state
	// for now the user will just need to manually retry the ones that failed though
	for _, o := range outdated {
		c.debugger.Debugf("Applying changes to dotfile %s", o.df.Name)
		if err := c.copyDotfile(o.df); err != nil {
			return errors.Wrapf(err, "failed to apply changes to %s", o.df.Name)
		}
		c.lf.Dotfiles[o.df.Name] = dotfileInfo{o.newHash}
	}
	c.debugger.Debugf("Finished applying changes to dotfiles")
	if err := c.writeLockfile(); err != nil {
		return errors.Wrap(err, "failed to save lockfile")
	}
	return nil
}

func (c *Client) copyDotfile(df dotfile.Dotfile) error {
	dir := filepath.Dir(df.DstPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	srcFile, err := c.registry.OpenDotfile(df.SrcPath)
	if err != nil {
		return fmt.Errorf("failed to open source dotfile %q: %w", df.SrcPath, err)
	}
	defer srcFile.Close()

	stat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat %s", df.SrcPath)
	}
	dstFile, err := os.OpenFile(df.DstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, stat.Mode())
	if err != nil {
		return fmt.Errorf("failed to open/create file %q: %w", df.DstPath, err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy %q to %q: %w", df.SrcPath, df.DstPath, err)
	}
	return nil
}

// Utils

// supportsOS checks whether the dotfile supports the current OS.
func supportsOS(df dotfile.Dotfile) bool {
	// No OSes defined means all are supported
	if len(df.OS) == 0 {
		return true
	}
	currentOS := runtime.GOOS
	for _, os := range df.OS {
		if os == currentOS {
			return true
		}
		// macOS is supported as an alias for darwin
		if os == "macOS" && currentOS == "darwin" {
			return true
		}
	}
	return false
}

// md5Hash returns the md5 hash of the data read from rc.
// md5Hash will close rc when it is finished.
func md5Hash(rc io.ReadCloser) ([]byte, error) {
	defer rc.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, rc); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// expandTilde replaces a ~ at the start of a path with the given homeDir.
func expandTilde(p, homeDir string) string {
	if strings.HasPrefix(p, "~") {
		base := strings.TrimPrefix(p, "~")
		p = filepath.Join(homeDir, base)
	}
	return p
}
