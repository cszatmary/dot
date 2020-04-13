package util

import (
	"crypto/md5"
	"runtime"

	"github.com/pkg/errors"
)

func MD5Checksum(buf []byte) ([]byte, error) {
	hash := md5.New()
	_, err := hash.Write(buf)
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
