package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/TouchBistro/goutils/file"
	"github.com/cszatma/dot/util"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	lockfileName = ".dot.lock"
)

var (
	config   DotConfig
	lockfile Lockfile
)

type DotfileConfig struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
	OS   string `yaml:"os"`
}

type DotConfig struct {
	Dotfiles map[string]DotfileConfig `yaml:"dotfiles"`
}

type DotfileInfo struct {
	SrcHash  string `yaml:"srcHash"`
	DestHash string `yaml:"destHash"`
}

type Lockfile struct {
	DotfilesDir string                 `yaml:"dotfilesDir"`
	IsSetup     bool                   `yaml:"setup"`
	Dotfiles    map[string]DotfileInfo `yaml:"dotfiles"`
}

func Config() *DotConfig {
	return &config
}

func loadDotConfig() error {
	configPath := filepath.Join(lockfile.DotfilesDir, "dot.yml")
	if !file.FileOrDirExists(configPath) {
		return errors.Errorf("No such file %s", configPath)
	}

	cf, err := os.Open(configPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to open config file at path %s", configPath)
	}
	defer cf.Close()

	dec := yaml.NewDecoder(cf)
	err = dec.Decode(&config)
	if err != nil {
		return errors.Wrap(err, "Failed to decode config file")
	}

	for name, dotfile := range config.Dotfiles {
		// Remove dotfiles that are not for the current os
		if dotfile.OS != "*" {
			osList := strings.Split(dotfile.OS, ",")
			currentOS := util.CurrentOS()
			isValidOS := false

			for _, os := range osList {
				if os == currentOS {
					isValidOS = true
					break
				}
			}

			if !isValidOS {
				delete(config.Dotfiles, name)
			}
		}

		// Make src paths absolute
		if !filepath.IsAbs(dotfile.Src) {
			dotfile.Src = filepath.Join(lockfile.DotfilesDir, dotfile.Src)
		}

		// Expand tilde in dest paths
		if strings.HasPrefix(dotfile.Dest, "~") {
			base := strings.TrimPrefix(dotfile.Dest, "~")
			dotfile.Dest = filepath.Join(os.Getenv("HOME"), base)
		}

		config.Dotfiles[name] = dotfile
	}

	return nil
}

func Init() error {
	lockfilePath := filepath.Join(os.Getenv("HOME"), lockfileName)
	if !file.FileOrDirExists(lockfilePath) {
		lockfile = Lockfile{}
		return nil
	}

	lf, err := os.Open(lockfilePath)
	if err != nil {
		return errors.Wrapf(err, "Failed to open lockfile at path %s", lockfilePath)
	}
	defer lf.Close()

	dec := yaml.NewDecoder(lf)
	err = dec.Decode(&lockfile)
	if err != nil {
		return errors.Wrap(err, "Failed to decode lockfile")
	}

	if !lockfile.IsSetup {
		return nil
	}

	err = loadDotConfig()
	return errors.Wrap(err, "Failed to load dot config file")
}

func SaveLockfile() error {
	lockfilePath := filepath.Join(os.Getenv("HOME"), ".dot.lock")
	lf, err := os.Create(lockfilePath)
	if err != nil {
		return errors.Wrapf(err, "Failed to create lockfile at %s", lockfilePath)
	}
	defer lf.Close()

	// Add header comment
	_, err = lf.WriteString("# THIS IS AN AUTOGENERATED FILE. DO NOT EDIT THIS FILE DIRECTLY.\n\n")
	if err != nil {
		return errors.Wrap(err, "Failed to add header comment to lockfile")
	}

	enc := yaml.NewEncoder(lf)
	err = enc.Encode(&lockfile)
	return errors.Wrapf(err, "Failed to write lockfile to %s", lockfilePath)
}

func IsSetup() bool {
	return lockfile.IsSetup
}

func Setup(dotfilesDir string) error {
	lockfile.DotfilesDir = dotfilesDir
	lockfile.IsSetup = true

	err := loadDotConfig()
	if err != nil {
		return errors.Wrap(err, "Failed to load dot config file")
	}

	// Get hashes for src and dest
	// This will be used to determine if the dotfiles are out of date
	log.Debugln("Saving hashes of dotfiles")
	lockfile.Dotfiles = make(map[string]DotfileInfo)
	for name, dotfile := range config.Dotfiles {
		log.Debugf("Saving hashes of %s\n", name)
		// Do lazy way of reading for now
		// Dotfiles shouldn't be so bit that this is an issue
		srcBuf, err := ioutil.ReadFile(dotfile.Src)
		if err != nil {
			return errors.Wrapf(err, "failed to read contents of %s", dotfile.Src)
		}

		destBuf, err := ioutil.ReadFile(dotfile.Dest)
		if err != nil {
			return errors.Wrapf(err, "failed to read contents of %s", dotfile.Dest)
		}

		srcHash, err := util.MD5Checksum(srcBuf)
		if err != nil {
			return errors.Wrapf(err, "failed to get checksum of %s", dotfile.Src)
		}

		destHash, err := util.MD5Checksum(destBuf)
		if err != nil {
			return errors.Wrapf(err, "failed to get checksum of %s", dotfile.Dest)
		}

		lockfile.Dotfiles[name] = DotfileInfo{
			SrcHash:  string(srcHash),
			DestHash: string(destHash),
		}
	}
	log.Debugln("Finished saving hashes")

	// Create backups of dotfiles
	// This way we still have the original since dot will be messing with them
	log.Debugln("Creating backups of dotfiles")
	for name, dotfile := range config.Dotfiles {
		log.Debugf("Creating backup of %s\n", name)

		backupPath := dotfile.Dest + ".bak"
		err = file.CopyFile(dotfile.Dest, backupPath)
		if err != nil {
			return errors.Wrapf(err, "Failed to create backup of %s at %s", name, backupPath)
		}
	}
	log.Debugln("Finished creating backups")

	err = SaveLockfile()
	return errors.Wrap(err, "Failed to save lockfile")
}
