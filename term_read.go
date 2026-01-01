package main

import (
	"syscall"

	"golang.org/x/term"
)

func readPassword() ([]byte, error) {
	b, err := term.ReadPassword(int(syscall.Stdin))
	// term.ReadPassword doesn't print newline; mimic user pressing enter
	return b, err
}
