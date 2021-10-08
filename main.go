package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-colorable"
)

const (
	_STAMP_STYLE = "15:04:05.000"
)

func mains() error {
	if dispose := colorable.EnableColorsStdout(nil); dispose != nil {
		defer dispose()
	}
	out := colorable.NewColorableStdout()

	tempPath := os.TempDir()
	previous := make(map[string]struct{})
	filepath.Walk(tempPath, func(_path string, info fs.FileInfo, err error) error {
		path := _path[len(tempPath):]
		previous[path] = struct{}{}
		return nil
	})

	tick := time.Tick(time.Second / 5)
	for next := range tick {
		current := make(map[string]struct{})
		stamp := next.Format(_STAMP_STYLE)

		filepath.Walk(tempPath, func(_path string, info fs.FileInfo, err error) error {
			path := _path[len(tempPath):]
			if _, ok := previous[path]; ok {
				delete(previous, path)
			} else {
				fmt.Fprintf(out, "\x1B[32;1m%s Add %s\x1B[0m\n", stamp, path)
			}
			current[path] = struct{}{}
			return nil
		})
		for path := range previous {
			fmt.Fprintf(out, "\x1B[31;1m%s Del %s\x1B[0m\n", stamp, path)
		}
		previous = current
	}
	return nil
}

func main() {
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
