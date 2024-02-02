package internal

import (
	"hash"
	"io"
	"os"
	"reflect"
)

func GetFileChecksum(file *os.File, hash hash.Hash) ([]byte, error) {
	fstat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	var fbuffer []byte = make([]byte, fstat.Size())
	_, err = file.Read(fbuffer)
	if err != nil {
		return nil, err
	}

	if err != nil && err != io.EOF {
		return nil, err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	hash.Reset()
	hash.Write(fbuffer)
	return hash.Sum(nil), nil
}

func CheckFilesAreEqual(hashing_algo hash.Hash, files ...*os.File) (bool, error) {
	var last_file_sum []byte
	var err error
	last_file_sum, err = GetFileChecksum(files[0], hashing_algo)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		checksum, err := GetFileChecksum(file, hashing_algo)
		if err != nil {
			return false, err
		}

		if !reflect.DeepEqual(checksum, last_file_sum) {
			return false, &ChecksumError{
				Message: "Failure to verify integrity between file written to cache and target file.",
			}
		}
		last_file_sum = checksum
	}

	return true, nil
}
