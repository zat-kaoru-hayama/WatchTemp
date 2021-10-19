package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
)

const (
	_STAMP_STYLE = "15:04:05.000"
)

var (
	flagRoot  = flag.String("target", "", "Set the target `directory`")
	flagOnAdd = flag.String("add", "", "execute `commandline`({} is replaced to the path) on new file found")
	flagOnUpd = flag.String("upd", "", "execute `commandline({} is replaced to the path)` on file updated")
	flagOnDel = flag.String("del", "", "execute `commandline({} is replaced to the path)` on file deleted")
)

func eventAction(cmdline, filename string) {
	cmdline = strings.ReplaceAll(cmdline, `{}`, `"`+filename+`"`)
	fmt.Println(cmdline)
	System(cmdline)
}

func filesEqual(left, right fs.FileInfo) bool {
	var size1 int64 = 0
	var time1 time.Time
	if left != nil && !left.IsDir() {
		size1 = left.Size()
		time1 = left.ModTime()
	}
	var size2 int64
	var time2 time.Time
	if right != nil && !right.IsDir() {
		size2 = right.Size()
		time2 = right.ModTime()
	}
	return size1 == size2 && time1 == time2
}

func watch(rootPath string, out io.Writer) error {
	previous := make(map[string]fs.FileInfo)
	filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		previous[path] = info
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
		case next := <-ticker.C:
			current := make(map[string]fs.FileInfo)
			stamp := next.Format(_STAMP_STYLE)

			filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
				relPath, err := filepath.Rel(rootPath, path)
				if err != nil {
					relPath = path
				}
				if pre, ok := previous[path]; ok {
					if !filesEqual(pre, info) {
						if *flagOnUpd != "" {

							eventAction(*flagOnUpd, path)
						}
						fmt.Fprintf(out, "\x1B[33;1m%s Upd %s\x1B[0m\n", stamp, relPath)
					}
					delete(previous, path)
				} else {
					fmt.Fprintf(out, "\x1B[32;1m%s Add %s\x1B[0m\n", stamp, relPath)
					if *flagOnAdd != "" {
						eventAction(*flagOnAdd, path)
					}
				}
				current[path] = info
				return nil
			})
			for path := range previous {
				relPath, err := filepath.Rel(rootPath, path)
				if err != nil {
					relPath = path
				}
				fmt.Fprintf(out, "\x1B[31;1m%s Del %s\x1B[0m\n", stamp, relPath)
				if *flagOnDel != "" {
					eventAction(*flagOnDel, path)
				}
			}
			previous = current
		case <-ctrlc:
			fmt.Fprintln(out, "\x1B[37;1mDone\x1B[0m")
			return nil
		}
	}
}

func mains() error {
	if dispose := colorable.EnableColorsStdout(nil); dispose != nil {
		defer dispose()
	}
	target := *flagRoot
	if target == "" {
		target = os.TempDir()
	}
	return watch(target, colorable.NewColorableStdout())
}

func main() {
	flag.Parse()
	if err := mains(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
