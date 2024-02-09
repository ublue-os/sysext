package fileio

import (
	"errors"
	"io"
	"os"
)

// Legally stolen from HikariKnight on Github!

// Creates a file and appends the content to the file (ending newline must be supplied with content string)
func FileAppendS(fileName string, content string) (int, error) {
	return FileAppend(fileName, []byte(content))
}

func FileAppend(fileName string, content []byte) (int, error) {
	// Open the file
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var bytes_written int
	bytes_written, err = f.Write(content)
	if err != nil {
		return bytes_written, err
	}
	return bytes_written, nil
}

func FileExist(fileName string) bool {
	if _, err := os.Stat(fileName); !errors.Is(err, os.ErrNotExist) {
		return true
	}
	return false
}

func FileCopy(sourceFile, destFile string) error {
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
