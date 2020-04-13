package util

import (
	"crypto/md5"
	"io/ioutil"
	"runtime"

	"github.com/pkg/errors"
)

func FileChecksum(path string) ([]byte, error) {
	// Do lazy way of reading for now
	// Dotfiles shouldn't be so bit that this is an issue
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read contents of %s", path)
	}

	hash := md5.New()
	_, err = hash.Write(buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write to hash")
	}

	return hash.Sum(nil), nil
}

func CurrentOS() string {
	os := runtime.GOOS
	switch os {
	case "darwin":
		return "macos"
	default:
		return os
	}
}
