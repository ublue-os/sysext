package fileio

import (
	"errors"
	"io"
	"os"
)

// Legally stolen from HikariKnight on Github!

func FileExist(fileName string) bool {
	if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) {
		return true
	}
	return false
}

func FileCopy(sourceFile, destFile string) error {
	filestat, err := os.Stat(sourceFile)
	if err != nil {
		return err
	}

	if !filestat.Mode().IsRegular() {
		return nil
	}
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return err
	}
	return nil
}
