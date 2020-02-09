package util

import (
	"bytes"
	"crypto/md5"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
)

// FileOrDirExists returns a bool indicating if a file or directory at the given path exists.
func FileOrDirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func Exec(name, dir string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "Exec failed to run %s %s", name, arg)
	}

	return nil
}

func ExecOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	return stdout.String(), errors.Wrapf(err, "exec failed for command %s: %s", name, stderr.String())
}

func MD5Checksum(buf []byte) ([]byte, error) {
	hash := md5.New()
	_, err := hash.Write(buf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write to hash")
	}

	return hash.Sum(nil), nil
}

func CopyFile(src, dest string) error {
	srcData, err := ioutil.ReadFile(src)
	if err != nil {
		return errors.Wrapf(err, "Failed to read source file %s", src)
	}

	stat, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "Failed to get info of source file %s", src)
	}

	err = ioutil.WriteFile(dest, srcData, stat.Mode().Perm())
	return errors.Wrapf(err, "Failed to write file to %s", dest)
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
