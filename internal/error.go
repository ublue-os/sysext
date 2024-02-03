package internal

import (
	"fmt"
	"strings"
)

// Mostly just to warn the user that a positional argument is missing
type PositionalArgumentError struct {
	Message []string
}

func (e *PositionalArgumentError) Error() string {
	var upperMessage []string
	for _, message := range e.Message {
		upperMessage = append(upperMessage, strings.ToUpper(message))
	}

	if len(e.Message) == 1 {
		return fmt.Sprintf("Required positional argument: %s", upperMessage[0])
	}
	return fmt.Sprintf("Required positional arguments: %s", strings.Join(upperMessage, ", "))
}

func NewPositionalError(message ...string) error {
	return &PositionalArgumentError{Message: message}
}
