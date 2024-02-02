package internal

import "fmt"

// Indicates that some check about the object's integrity is not passing.
type ChecksumError struct {
	Message string
}

func (e *ChecksumError) Error() string {
	return fmt.Sprintf("Unexpected Checksum Error: %s", e.Message)
}
