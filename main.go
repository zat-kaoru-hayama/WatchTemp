package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/mattn/go-colorable"
)

const (
	_STAMP_STYLE = "15:04:05.000"
)

func watch(rootPath string, out io.Writer) error {
	previous := make(map[string]struct{})
	filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		previous[path] = struct{}{}
		return nil
	})

	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	ticker := time.NewTicker(time.Second / 5)
	defer ticker.Stop()

	fmt.Fprintf(out, "\x1B[37;1mWatch Start: %s\x1B[0m\n", rootPath)
	for {
		select {
		case <-ctrlc:
			fmt.Fprintln(out, "\x1B[37;1mDone\x1B[0m")
			return nil
		case next := <-ticker.C:
			current := make(map[string]struct{})
			stamp := next.Format(_STAMP_STYLE)

			filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
				if _, ok := previous[path]; ok {
					delete(previous, path)
				} else {
					relPath, err := filepath.Rel(rootPath, path)
					if err != nil {
						relPath = path
					}
					fmt.Fprintf(out, "\x1B[32;1m%s Add %s\x1B[0m\n", stamp, relPath)
				}
				current[path] = struct{}{}
				return nil
			})
			for path := range previous {
				relPath, err := filepath.Rel(rootPath, path)
				if err != nil {
					relPath = path
				}
				fmt.Fprintf(out, "\x1B[31;1m%s Del %s\x1B[0m\n", stamp, relPath)
			}
			previous = current
		}
	}
}

func mains() error {
	if dispose := colorable.EnableColorsStdout(nil); dispose != nil {
		defer dispose()
	}
	return watch(os.TempDir(), colorable.NewColorableStdout())
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
