package fsutil

import (
	"bytes"
	"github.com/c2h5oh/datasize"
	"github.com/pkg/errors"
	"io"
	"os"
)

// Exists returns true if path is an existing file or directory, otherwise it
// returns false. If followLinks is true, then Exists will attempt to follow
// links to their target and report said target's existence. If followLinks is
// false, Exist will operate on the link itself.
func Exists(path string, followLinks bool) (bool, error) {
	var err error
	if followLinks {
		_, err = os.Stat(path)
	} else {
		_, err = os.Lstat(path)
	}
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.Wrapf(err, "stat %#v failed", path)
}

// SameFile wraps os.SameFile and handles calling os.Stat on both paths.
func SameFile(pathA, pathB string) (bool, error) {
	aInfo, err := os.Stat(pathA)
	if err != nil {
		return false, errors.Wrapf(err, "stat %#v failed", pathA)
	}
	bInfo, err := os.Stat(pathB)
	if err != nil {
		return false, errors.Wrapf(err, "stat %#v failed", pathB)
	}
	return os.SameFile(aInfo, bInfo), nil
}

// IsLink returns true if path represents a symlink, otherwise it returns false.
func IsLink(path string) (bool, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode()&os.ModeSymlink != 0, nil
}

// IsRegularFile returns true if path represents a regular file, otherwise it returns false.
func IsRegularFile(path string) (bool, error) {
	// If we use os.Stat, it'll follow links, which we don't want.
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode().IsRegular(), nil
}

// SameContents checks that two files contain the same bytes
func SameContents(pathA, pathB string) (bool, error) {
	fileA, err := os.Open(pathA)
	if err != nil {
		return false, errors.Wrapf(err, "open %#v failed", pathA)
	}
	defer fileA.Close()
	fileB, err := os.Open(pathB)
	if err != nil {
		return false, errors.Wrapf(err, "open %#v failed", pathB)
	}
	defer fileB.Close()

	bytesA := make([]byte, 8*datasize.MB)
	bytesB := make([]byte, 8*datasize.MB)
	for {
		nBytesReadA, errA := fileA.Read(bytesA)
		isEndOfFileA := errA == io.EOF
		if errA != nil && !isEndOfFileA {
			return false, errors.Wrapf(errA, "read %#v failed", pathA)
		}
		nBytesReadB, errB := fileB.Read(bytesB)
		isEndOfFileB := errB == io.EOF
		if errB != nil && !isEndOfFileB {
			return false, errors.Wrapf(errB, "read %#v failed", pathB)
		}
		if nBytesReadA != nBytesReadB {
			return false, nil
		}
		if !bytes.Equal(bytesA, bytesB) {
			return false, nil
		}
		if isEndOfFileA != isEndOfFileB {
			return false, nil
		} else if isEndOfFileA {
			return true, nil
		}
	}
}
