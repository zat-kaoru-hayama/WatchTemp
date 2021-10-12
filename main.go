package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
)

const (
	_STAMP_STYLE = "15:04:05.000"
)

type FileStatus struct {
	ModTime time.Time
	Size    int64
}

var (
	flagRoot  = flag.String("target", "", "Set the target `directory`")
	flagOnAdd = flag.String("add", "", "execute `commandline`({} is replaced to the path) on new file found")
	flagOnUpd = flag.String("upd", "", "execute `commandline({} is replaced to the path)` on file updated")
	flagOnDel = flag.String("del", "", "execute `commandline({} is replaced to the path)` on file deleted")
)

func system(cmdline string) func() {
	os.Setenv("CMDLINE", cmdline)
	cmd := exec.Command("cmd.exe", "/S", "/C", "%CMDLINE%")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return func() { cmd.Wait() }
}

func eventAction(cmdline, filename string) {
	cmdline = strings.ReplaceAll(cmdline, `{}`, `"`+filename+`"`)
	fmt.Println(cmdline)
	system(cmdline)
}

func watch(rootPath string, out io.Writer) error {
	previous := make(map[string]FileStatus)
	filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		previous[path] = FileStatus{
			ModTime: info.ModTime(),
			Size:    info.Size(),
		}
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
		case next, ok := <-ticker.C:
			if !ok {
				return errors.New("Closed timer")
			}
			current := make(map[string]FileStatus)
			stamp := next.Format(_STAMP_STYLE)

			filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
				relPath, err := filepath.Rel(rootPath, path)
				if err != nil {
					relPath = path
				}
				if pre, ok := previous[path]; ok {
					if info != nil && !info.IsDir() && (pre.Size != info.Size() || pre.ModTime != info.ModTime()) {
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
				if info != nil {
					current[path] = FileStatus{
						ModTime: info.ModTime(),
						Size:    info.Size(),
					}
				} else {
					current[path] = FileStatus{}
				}
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
