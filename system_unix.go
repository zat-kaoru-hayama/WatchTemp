//go:build !windows
// +build !windows

package main

import (
	"os"
)

func System(cmdline string) (*os.Process, error) {
	shell := os.Getenv("SHELL")
	return os.StartProcess(
		shell,
		[]string{shell, "-c", cmdline},
		&os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		})
}
