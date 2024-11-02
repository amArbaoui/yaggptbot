package util

import (
	"os"
)

func FileSizeInKib(path string) (int64, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return 0, nil
	}
	size := fi.Size()
	return size / 1024, nil
}
