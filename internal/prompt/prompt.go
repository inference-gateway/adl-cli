package prompt

import (
	"fmt"
	"log"
	"strings"

	"github.com/chzyer/readline"
)

// ReadString reads a line of input from the terminal with support for arrow keys and editing
func ReadString(promptText, defaultValue string) (string, error) {
	var prompt string
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s [%s]: ", promptText, defaultValue)
	} else {
		prompt = fmt.Sprintf("%s: ", promptText)
	}

	rl, err := readline.New(prompt)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := rl.Close(); closeErr != nil {
			log.Printf("failed to close readline: %v", closeErr)
		}
	}()

	line, err := rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt {
			return "", err
		}
		return "", err
	}

	line = strings.TrimSpace(line)

	if line == "" {
		return defaultValue, nil
	}

	return line, nil
}

// ReadPassword reads a password from the terminal (characters are hidden)
func ReadPassword(promptText string) (string, error) {
	rl, err := readline.New(fmt.Sprintf("%s: ", promptText))
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := rl.Close(); closeErr != nil {
			log.Printf("failed to close readline: %v", closeErr)
		}
	}()

	password, err := rl.ReadPassword(fmt.Sprintf("%s: ", promptText))
	if err != nil {
		if err == readline.ErrInterrupt {
			return "", err
		}
		return "", err
	}

	return string(password), nil
}
