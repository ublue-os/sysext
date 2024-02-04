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

type InvalidOptionError struct {
	Message []string
}

func (e *InvalidOptionError) Error() string {
	var upperMessage []string
	for _, message := range e.Message {
		upperMessage = append(upperMessage, strings.ToUpper(message))
	}

	if len(e.Message) == 1 {
		return fmt.Sprintf("Invalid option: %s", upperMessage[0])
	}
	return fmt.Sprintf("Invalid options: %s", strings.Join(upperMessage, ", "))
}

func NewInvalidOptionError(message ...string) error {
	return &InvalidOptionError{Message: message}
}
